package routes

import (
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	_ "image/png"

	"github.com/ethrousseau/weblens/api/database"
	"github.com/ethrousseau/weblens/api/importMedia"
	"github.com/ethrousseau/weblens/api/interfaces"
	log "github.com/ethrousseau/weblens/api/utils"

	"github.com/gin-gonic/gin"
)

func getPagedMedia(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	group := ctx.Query("group")
	if group == "" {
		group = "false"
	}
	if group != "true" && group != "false" {
		ctx.JSON(http.StatusBadRequest, "group, if passed, must be a boolean: true | false")
	}

	var err error

	var skip int
	if (ctx.Query("skip") != "") {
		skip, err = strconv.Atoi(ctx.Query("skip"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "skip paramater must be an interger")
			return
		}
	} else {
		skip = 0
	}

	var limit int
	if (ctx.Query("limit") != "") {
		limit, err = strconv.Atoi(ctx.Query("limit"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "limit paramater must be an interger")
			return
		}
	} else {
		limit = 0
	}

	db := database.New()
	media := db.GetPagedMedia(sort, group, skip, limit)

	ctx.JSON(http.StatusOK, media)
}

func scan(ctx *gin.Context) {
	importMedia.ScanAllMedia()
}

func getPhotoThumb(ctx *gin.Context) {
	db := database.New()

	i := db.GetImage(ctx.Query("filehash"))
	bytes, err := b64.StdEncoding.DecodeString(i.Thumbnail64)
	if err != nil {
		panic(err)
	}
	ctx.Writer.Write(bytes)

}

func uploadPhoto(ctx *gin.Context) () {
	inFile, fileHeader, err := ctx.Request.FormFile("uploadMedia")
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	imgPath := "/Users/ethan/repos/weblens/images"
	outFile, err := os.Create(fmt.Sprintf("%s/%s", imgPath, fileHeader.Filename))
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, inFile)
    if err != nil {
        panic(err)
    }

	ctx.Writer.Header().Set("Content-Type", "application/json")
	ctx.Writer.WriteHeader(http.StatusCreated)

	m := new(interfaces.Media)
	db := database.New()

	importMedia.HandleNewImage(imgPath, m, db)

}

func getFillresMedia(ctx *gin.Context) {
	db := database.New()
	m := db.GetImage(ctx.Param("filehash"))
	fileBytes, err := os.ReadFile(m.Filepath)
	if err != nil {
		log.Error.Printf("Trying to open full res image %v\n", err)
		return
	}

	ctx.Writer.Write(fileBytes)
}
