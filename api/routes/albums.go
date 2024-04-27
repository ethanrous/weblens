package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

type albumUpdateData struct {
	AddMedia    []types.FileId    `json:"newMedia"`
	AddFolders  []types.FileId    `json:"newFolders"`
	RemoveMedia []types.ContentId `json:"removeMedia"`
	Cover       types.ContentId   `json:"cover"`
	NewName     string            `json:"newName"`
	Users       []types.Username  `json:"users"`
	RemoveUsers []types.Username  `json:"removeUsers"`
}

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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body not in propper format"})
		return
	}

	err = dataStore.CreateAlbum(albumData.Name, user)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
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

	var update albumUpdateData
	err = json.Unmarshal(jsonData, &update)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Decoded but unable to parse request body"})
		return
	}

	ms := []types.Media{}
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms = util.Map(update.AddMedia, func(fId types.FileId) types.Media {
			f := dataStore.FsTreeGet(fId)
			m := dataStore.MediaMapGet(f.GetContentId())
			return m
		})
		ms = util.Filter(ms, func(m types.Media) bool { return m != nil })
		if len(ms) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid media Ids in request"})
			return
		}
	}

	if update.AddFolders != nil && len(update.AddFolders) != 0 {

		ms = append(ms, util.Map(dataStore.RecursiveGetMedia(update.AddFolders...), func(mId types.ContentId) types.Media {
			m := dataStore.MediaMapGet(mId)
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
				coverMediaId = ms[0].Id()
			}
			err = a.SetCover(coverMediaId)
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
		err = a.SetCover(update.Cover)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		a.Rename(update.NewName)
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare user(s)"})
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
	db := dataStore.NewDB()
	a, err := db.GetAlbum(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || err != nil || !a.CanUserAccess(user.GetUsername()) {
		util.ErrTrace(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare themself from it
	if a.Owner != user.GetUsername() {
		a.RemoveUsers([]types.Username{user.GetUsername()})
		ctx.Status(http.StatusOK)
		return
	}

	err = db.DeleteAlbum(albumId)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}
