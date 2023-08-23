package importMedia

import (
	"bytes"
	"crypto/sha256"
	b64 "encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/interfaces"

	"github.com/barasher/go-exiftool"
	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
)

func readFile(filepath string) (image.Image, error) {
	loadedImg, err := imaging.Open(filepath)
	if err != nil {
		return nil, err
	}

	return loadedImg, nil

}

func getImageHash(i image.Image) (string) {
	imgBuf := new(bytes.Buffer)
	jpeg.Encode(imgBuf, i, nil)

	h := sha256.New()
	_, err := io.Copy(h, imgBuf)
	if err != nil {
		panic(err)
	}

	return b64.URLEncoding.EncodeToString(h.Sum(nil))
}

func calculateThumbSize(i image.Image) (int, int) {
	dimentions := i.Bounds()
	width, height := dimentions.Dx(), dimentions.Dy()

	aspectRatio := float64(width) / float64(height)

	var newWidth, newHeight int

	if aspectRatio > 1 {
		newWidth = 1024
		newHeight = int(1024 / aspectRatio)
	} else {
		newWidth = int(1024 * aspectRatio)
		newHeight = 1024
	}

	return newWidth, newHeight

}

func generateThumbnail(i image.Image, width, height int) (*image.NRGBA, string, error) {
	thumb := imaging.Thumbnail(i, width, height, imaging.CatmullRom)

	thumbBytesBuf := new(bytes.Buffer)
	jpeg.Encode(thumbBytesBuf, thumb, nil)

	thumb64 := b64.StdEncoding.EncodeToString(thumbBytesBuf.Bytes())

	return thumb, thumb64, nil

}

func generateBlurhash(thumb *image.NRGBA) (string, error) {
	blurStr, err := blurhash.Encode(4, 3, thumb)
	if err != nil {
		return "", err
	}

	return blurStr, nil

}

func extractExif(filepath string) (time.Time, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return time.Time{}, err
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(filepath)

	var createDate time.Time

	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}

		createString, ok := fileInfo.Fields["SubSecCreateDate"]
		if ok {
			createDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", createString.(string))
		} else {
			createDate, err = time.Now(), nil
		}
	}

	return createDate, err

}

func handleNewImage(imagesDir, filename string) (*interfaces.Media, error) {
	filepath := (imagesDir + "/" + filename)

	i, err := readFile(filepath)
	if err != nil {
		return nil, err
	}

	hash := getImageHash(i)

	width, height := calculateThumbSize(i)
	thumb, thumb64, err := generateThumbnail(i, width, height)
	if err != nil {
		return nil, err
	}

	blur, err := generateBlurhash(thumb)
	if err != nil {
		return nil, err
	}

	createDate, err := extractExif(filepath)
	if err != nil {
		return nil, err
	}

	m := &interfaces.Media{
		Filename: filename,
		FileHash: hash,
		BlurHash: blur,
		CreateDate: createDate,
		Thumbnail64: thumb64,
	}

	database.addImage(m)

	return m, nil

}

func scanAllMedia() {
	fmt.Print("Beginning file scan\n")
	imagesDir := "/Users/ethan/repos/weblens/images"

	files, _ := os.ReadDir(imagesDir)

	var wg sync.WaitGroup
	wg.Add(len(files))

	for _, file := range files {
		go func(file fs.DirEntry) {
			defer wg.Done()

			_, exists := database.getImageByFilename(file.Name())

			if !exists {
				_, err := handleNewImage(imagesDir, file.Name())
				if err != nil {
					fmt.Printf("ERR handleNewImage: %s\n", err)
					return
				}

			}

		}(file)
	}

	wg.Wait()

	fmt.Print("Finished file scan")

}