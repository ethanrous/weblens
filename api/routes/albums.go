package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

type albumUpdateData struct {
	AddMedia    []string `json:"newMedia"`
	AddFolders  []string `json:"newFolders"`
	RemoveMedia []string `json:"removeMedia"`
	Cover       string   `json:"cover"`
	NewName     string   `json:"newName"`
	Users       []string `json:"users"`
	RemoveUsers []string `json:"removeUsers"`
	CleanMedia  bool     `json:"cleanMissing"`
}

func getAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

	raw := ctx.Query("raw") == "true"

	db := dataStore.NewDB()
	albumData, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Unable to find album information"})
		return
	}

	medias, err := db.GetFilteredMedia("createDate", ctx.GetString("username"), 1, []string{albumId}, raw)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filtered media"})
	}

	ctx.JSON(http.StatusOK, gin.H{"albumMeta": albumData, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	username := ctx.GetString("username")
	filter := ctx.Query("filter")
	db := dataStore.NewDB()
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

	db := dataStore.NewDB()
	db.CreateAlbum(albumData.Name, ctx.GetString("username"))
}

func updateAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")
	a, err := dataStore.GetAlbum(albumId)

	if err != nil {
		util.DisplayError(err)
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

	var ms []string
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms = util.Filter(update.AddMedia, func(m string) bool { return m != "" })
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
			d.RecursiveMap(func(f *dataStore.WeblensFile) {
				if !f.IsDir() {
					m, err := f.GetMedia()
					if err != nil {
						util.DisplayError(err)
						return
					}
					if m != nil {
						ms = append(ms, m.MediaId)
					}
				}
			})
		}
	}

	addedCount := 0
	if len(ms) != 0 {
		addedCount, err = a.AddMedia(ms)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if a.Cover == "" || a.PrimaryColor == "" {
			coverMediaId := a.Cover
			if coverMediaId == "" {
				coverMediaId = ms[0]
			}
			err = a.SetCover(coverMediaId)
			if err != nil {
				util.DisplayError(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
				return
			}
		}
	}

	if update.RemoveMedia != nil {
		err = a.RemoveMedia(update.RemoveMedia)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove media from album"})
			return
		}
	}

	if update.CleanMedia {
		err = a.CleanMissingMedia()
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed clean missing album media"})
		}
	}

	if update.Cover != "" {
		err = a.SetCover(update.Cover)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		a.Rename(update.NewName)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
			return
		}
	}

	if update.RemoveUsers != nil {
		err = a.RemoveUsers(update.RemoveUsers)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare user(s)"})
			return
		}
	}

	if update.Users != nil {
		err = a.AddUsers(update.Users)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share user(s)"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"errors": []string{}, "addedCount": addedCount})
}

func deleteAlbum(ctx *gin.Context) {
	username := ctx.GetString("username")
	albumId := ctx.Param("albumId")
	db := dataStore.NewDB()
	a, err := db.GetAlbum(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || err != nil || !a.CanUserAccess(username) {
		util.DisplayError(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare themself from it
	if a.Owner != username {
		a.RemoveUsers([]string{username})
		ctx.Status(http.StatusOK)
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
