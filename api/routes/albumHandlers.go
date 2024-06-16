package routes

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/album"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func getAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albumId := types.AlbumId(ctx.Param("albumId"))

	raw := ctx.Query("raw") == "true"

	db := dataStore.NewDB()
	albumData, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Unable to find album information"})
		return
	}

	medias, err := db.GetFilteredMedia("createDate", user.GetUsername(), 1, []types.AlbumId{albumId}, raw)
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filtered media"})
	}

	ctx.JSON(http.StatusOK, gin.H{"albumMeta": albumData, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	filter := ctx.Query("filter")
	db := dataStore.NewDB()
	albums := db.GetAlbumsByUser(user.GetUsername(), filter, true)
	ctx.JSON(http.StatusOK, gin.H{"albums": albums})
}

func createAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	type albumCreateData struct {
		Name string `json:"name"`
	}
	var albumData albumCreateData
	err = json.Unmarshal(jsonData, &albumData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body not in proper format"})
		return
	}

	newAlbum := album.New(albumData.Name, user)
	err = rc.AlbumManager.Add(newAlbum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Album creation failed"})
	}
}

func updateAlbum(ctx *gin.Context) {
	albumId := types.AlbumId(ctx.Param("albumId"))
	a, err := dataStore.GetAlbum(albumId)

	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusNotFound)
	}

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var update updateAlbumBody
	err = json.Unmarshal(jsonData, &update)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Decoded but unable to parse request body"})
		return
	}

	var ms []types.Media
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms = util.FilterMap(update.AddMedia, func(mId types.ContentId) (types.Media, bool) {
			m := rc.MediaRepo.Get(mId)
			if m != nil {
				return m, true
			} else {
				return m, false
			}
		})

		if len(ms) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid media Ids in request"})
			return
		}
	}

	if update.AddFolders != nil && len(update.AddFolders) != 0 {
		folders := util.Map(update.AddFolders, func(fId types.FileId) types.WeblensFile {
			return rc.FileTree.Get(fId)
		})

		ms = append(ms, util.Map(dataStore.RecursiveGetMedia(rc.MediaRepo, folders...), func(mId types.ContentId) types.Media {
			m := rc.MediaRepo.Get(mId)
			return m
		})...)
	}

	addedCount := 0
	if len(ms) != 0 {
		addedCount, err = a.AddMedia(ms)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if a.Cover == "" || a.PrimaryColor == "" {
			coverMediaId := a.Cover
			if coverMediaId == "" {
				coverMediaId = ms[0].ID()
			}
			err = a.SetCover(coverMediaId, rc.FileTree)
			if err != nil {
				util.ErrTrace(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
				return
			}
		}
	}

	if update.RemoveMedia != nil {
		err = a.RemoveMedia(update.RemoveMedia)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove media from album"})
			return
		}
	}

	if update.Cover != "" {
		err = a.SetCover(update.Cover, rc.FileTree)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err = a.Rename(update.NewName)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
			return
		}
	}

	if update.RemoveUsers != nil {
		err = a.RemoveUsers(update.RemoveUsers)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
	}

	if update.Users != nil {
		err = a.AddUsers(update.Users)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share user(s)"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"errors": []string{}, "addedCount": addedCount})
}

func deleteAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albumId := types.AlbumId(ctx.Param("albumId"))

	a, err := dataStore.GetAlbum(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || err != nil || !a.CanUserAccess(user.GetUsername()) {
		util.ErrTrace(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare them from the album
	if a.Owner != user.GetUsername() {
		err = a.RemoveUsers([]types.Username{user.GetUsername()})
		ctx.Status(http.StatusOK)
		return
	}

	err = dataStore.DeleteAlbum(albumId)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func unshareMeAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	albumId := types.AlbumId(ctx.Param("albumId"))
	album, err := dataStore.GetAlbum(albumId)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	if !album.CanUserAccess(user.GetUsername()) {
		ctx.Status(http.StatusNotFound)
		return
	}

	err = album.RemoveUsers([]types.Username{user.GetUsername()})
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func albumPreviewMedia(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

	a, err := dataStore.GetAlbum(types.AlbumId(albumId))
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	albumMs := a.Medias
	randomMs := make([]types.ContentId, 0, 9)

	for len(albumMs) != 0 && len(randomMs) < 9 {
		index := rand.Intn(len(albumMs))
		m := rc.MediaRepo.Get(albumMs[index])
		if m != nil && !m.GetMediaType().IsRaw() && m.ID() != a.Cover {
			randomMs = append(randomMs, m.ID())
		}

		albumMs = util.Banish(albumMs, index)
	}

	ctx.JSON(http.StatusOK, gin.H{"mediaIds": randomMs})
}
