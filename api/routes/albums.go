package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

type albumUpdateData struct {
	AddMedia []string `json:"newMedia"`
	AddFolders []string `json:"newFolders"`
	RemoveMedia []string `json:"removeMedia"`
	Cover string `json:"cover"`
	NewName string `json:"newName"`
	Users []string	`json:"users"`
	RemoveUsers []string `json:"removeUsers"`
}

func getAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

	raw := ctx.Query("raw") == "true"

	db := dataStore.NewDB(ctx.GetString("username"))
	albumData, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Unable to find album information"})
		return
	}

	medias, err := db.GetFilteredMedia("createDate", ctx.GetString("username"), 1, []string{albumId}, raw, false)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filtered media"})
	}

	ctx.JSON(http.StatusOK, gin.H{"albumMeta": albumData, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	username := ctx.GetString("username")
	filter := ctx.Query("filter")
	db := dataStore.NewDB(username)
	albums := db.GetAlbumsByUser(username, filter, true)
	ctx.JSON(http.StatusOK, gin.H{"albums": albums})
}

func createAlbum(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	type albumCreateData struct {
		Name string `json:"name"`
	}
	var albumData albumCreateData
	json.Unmarshal(jsonData, &albumData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body not in propper format"})
		return
	}

	db := dataStore.NewDB(ctx.GetString("username"))
	db.CreateAlbum(albumData.Name, ctx.GetString("username"))
}

func updateAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

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

	db := dataStore.NewDB(ctx.GetString("username"))
	album, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	ms := []string{}
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms := util.Filter(update.AddMedia, func(m string) bool {return m != ""})
		if len(ms) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid media Ids in request"})
			return
		}
	}
	if update.AddFolders != nil && len(update.AddFolders) != 0 {
		for _, dId := range update.AddFolders {
			d := dataStore.FsTreeGet(dId)
			if d == nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder id"})
				return
			}
			d.RecursiveMap(func(f *dataStore.WeblensFileDescriptor) {
				if !f.IsDir() {
					m, err := f.GetMedia()
					if err != nil {
						util.DisplayError(err)
						return
					}
					if m != nil {
						ms = append(ms, m.FileHash)
					}
				}
			})
		}
	}

	addedCount := 0
	if len(ms) != 0 {
		addedCount, err = db.AddMediaToAlbum(albumId, ms)
		if err != nil {
			util.Error.Printf("Failed to add media to album (%s) %v", albumId, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if album.Cover == "" || album.PrimaryColor == "" {
			coverMediaId := album.Cover
			if coverMediaId == "" {
				coverMediaId = ms[0]
			}
			err = setAlbumCover(albumId, coverMediaId, db)
			if err != nil {
				util.DisplayError(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
				return
			}
		}
	}

	if update.RemoveMedia != nil {
		db.RemoveMediaFromAlbum(albumId, update.RemoveMedia)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove media from album"})
			return
		}
	}

	if update.Cover != "" {
		err = setAlbumCover(albumId, update.Cover, db)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err = db.SetAlbumName(albumId, update.NewName)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
			return
		}
	}

	if update.RemoveUsers != nil {
		err = db.UnshareAlbum(albumId, update.RemoveUsers)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare user(s)"})
			return
		}
	}

	if update.Users != nil {
		err = db.ShareAlbum(albumId, update.Users)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share user(s)"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"errors": []string{}, "addedCount": addedCount})
}

func setAlbumCover(albumId, mediaId string, db *dataStore.Weblensdb) error {
	m := db.GetMedia(mediaId, false)
	i := (&m).GetImage()
	prom, err := prominentcolor.Kmeans(i)
	if err != nil {
		return err
	}

	err = db.SetAlbumCover(albumId, mediaId, prom[0].AsString(), prom[1].AsString())
	if err != nil {
		return fmt.Errorf("failed to set album cover")
	}
	return nil
}

func deleteAlbum(ctx *gin.Context) {
	username := ctx.GetString("username")
	albumId := ctx.Param("albumId")
	db := dataStore.NewDB(username)
	album, err := db.GetAlbum(albumId)

	// err or user does not have access to this album, claim not found
	if err != nil || (album.Owner != username && slices.Contains(album.SharedWith, username)){
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user has access, but is not the owner,
	// be more helpful and explain they are not allowed to delete the album
	if album.Owner != username {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete an album you do not own"})
		return
	}

	err = db.DeleteAlbum(albumId)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}
