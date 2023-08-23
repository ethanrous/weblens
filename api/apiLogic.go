package api

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func listPhotos(ctx *gin.Context) {
	imageArray := database.getAllImages()
	ctx.JSON(http.StatusOK, imageArray)
}

func scan(ctx *gin.Context) {
	importMedia.scanAllMedia()
}

func getPhotoThumb(ctx *gin.Context) {
	fmt.Printf("Hash: %s\n", ctx.Query("filehash"))
	var thumb64 = ""
	//thumb64 := getImageThumb(ctx.Query("filehash"))
	bytes, _ := b64.StdEncoding.DecodeString(thumb64)

	fmt.Printf("Thumb64: %s\n", thumb64)

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

	_, err = importMedia.handleNewImage(imgPath, fileHeader.Filename)
	if err != nil {
        panic(err)
    }
}

func getPhoto(ctx *gin.Context) {
	return
	imagePath := fmt.Sprintf("/Users/ethan/repos/weblens/images/%s", ctx.Param("filename"))
	image, _ := imaging.Open(imagePath)

	sideWidth, sideHeight := 100, 100
	thumb := imaging.Thumbnail(image, sideWidth, sideHeight, imaging.CatmullRom)

	buf := new(bytes.Buffer)
	jpeg.Encode(buf, thumb, nil)

	ctx.Writer.Write(buf.Bytes())
}
