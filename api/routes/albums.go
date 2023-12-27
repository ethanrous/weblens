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

type albumCreateData struct {
	Name string `json:"name"`
}

type albumUpdateData struct {
	Media []string `json:"media"`
	Cover string `json:"cover"`
}

func getAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

	db := dataStore.NewDB(ctx.GetString("username"))
	albumData, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Unable to find album information"})
		return
	}

	medias := util.Map(albumData.Medias, func(mediaHash string) (m dataStore.Media) {m = db.GetMedia(mediaHash, false); return})
	slices.SortFunc(medias, func(a, b dataStore.Media) int {return a.CreateDate.Compare(b.CreateDate)})
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

	var albumData albumCreateData
	json.Unmarshal(jsonData, &albumData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body not in propper format"})
		return
	}

	db := dataStore.NewDB(ctx.GetString("username"))
	db.CreateAlbum(albumData.Name, ctx.GetString("username"))
}

func addToAlbum(ctx *gin.Context) {
	albumId := ctx.Param("albumId")

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var update albumUpdateData
	json.Unmarshal(jsonData, &update)
	db := dataStore.NewDB(ctx.GetString("username"))
	album, err := db.GetAlbum(albumId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	if update.Media != nil {
		err = db.AddMediaToAlbum(albumId, update.Media)
		if err != nil {
			util.Error.Printf("Failed to add media to album (%s) %v", albumId, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}


		if album.Cover == "" || album.PrimaryColor == "" {
			coverMediaId := album.Cover
			if coverMediaId == "" {
				coverMediaId = album.Medias[0]
			}
			setAlbumCover(album.Id, coverMediaId, db)
		}
	}

	if update.Cover != "" {
		setAlbumCover(album.Id, update.Cover, db)
	}

	ctx.Status(http.StatusOK)
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
