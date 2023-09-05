package routes

import (
	b64 "encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "image/png"

	"github.com/ethrousseau/weblens/api/database"
	"github.com/ethrousseau/weblens/api/importMedia"
	"github.com/ethrousseau/weblens/api/interfaces"
	util "github.com/ethrousseau/weblens/api/utils"

	"github.com/gin-gonic/gin"
)

func getPagedMedia(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
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

	var raw bool
	if ctx.Query("raw") == "" || ctx.Query("raw") == "false" {
		raw = false
	} else {
		raw = true
	}

	db := database.New()
	media, moreMedia := db.GetPagedMedia(sort, skip, limit, raw, true)

	res := struct{
		Media []interfaces.Media
		MoreMedia bool
	} {
		media,
		moreMedia,
	}
	ctx.JSON(http.StatusOK, res)
}

func getMediaItem(ctx *gin.Context) {
	fileHash := ctx.Param("filehash")
	includeMeta := util.BoolFromString(ctx.Query("meta"), true)
	includeThumbnail := util.BoolFromString(ctx.Query("thumbnail"), true)
	includeFullres := util.BoolFromString(ctx.Query("fullres"), true)

	if !(includeMeta || includeThumbnail || includeFullres) {
		// At least one option must be selected
		ctx.AbortWithStatus(http.StatusBadRequest)
	} else if includeFullres && (includeMeta || includeThumbnail) {
		// Full res must be the only option if selected
		ctx.AbortWithStatus(http.StatusBadRequest)
	}

	db := database.New()
	data := db.GetMedia(fileHash, includeThumbnail)

	if includeFullres {
		fullResBytes := data.ReadFullres()
		ctx.Writer.Write(fullResBytes)
	} else if !includeMeta && includeThumbnail {
		thumbBytes, err := b64.StdEncoding.DecodeString(data.Thumbnail64)
		if err != nil {
			panic(err)
		}
		ctx.Writer.Write(thumbBytes)
	} else {
		ctx.JSON(http.StatusOK, data)
	}
}

func scan(ctx *gin.Context) {
	importMedia.ScanAllMedia()
}

type fileUpload struct{
	Filename string `json:"filename"`
	Item64 string `json:"item64"`
}

func uploadItem(ctx *gin.Context) () {
	requestString, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}
	item := fileUpload{}
	json.Unmarshal(requestString, &item)

	imgPath := "/Users/ethan/repos/weblens/images/upload" + item.Filename
	outFile, err := os.Create(imgPath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	index := strings.Index(item.Item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item.Item64[index + 1:])
	if err != nil {
		panic(err)
	}

	_, err = outFile.Write(itemBytes)
	if err != nil {
		panic(err)
	}

	db := database.New()

	importMedia.HandleNewImage(imgPath, db)

}
