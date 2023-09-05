package interfaces

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"os/exec"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"

	log "github.com/ethrousseau/weblens/api/utils"
)

type Media struct {
	FileHash		string				`bson:"_id"`
	Filepath 		string 				`bson:"filepath"`
	MediaType		mediaType			`bson:"mediaType"`
	BlurHash 		string 				`bson:"blurHash"`
	Thumbnail64 	string		 		`bson:"thumbnail"`
	ThumbWidth 		int					`bson:"width"`
	ThumbHeight 	int 				`bson:"height"`
	CreateDate		time.Time			`bson:"createDate"`
}

func (m Media) MarshalBinary() ([]byte, error) {
    return json.Marshal(m)
}


func (m *Media) ExtractExif() {
	et, err := exiftool.NewExiftool()
	if err != nil {
		panic(err)
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(m.Filepath)
	if fileInfos[0].Err != nil {
		panic(fileInfos[0].Err)
	}

	exifData := fileInfos[0].Fields

	r, ok := exifData["SubSecCreateDate"]
	if ok {
		m.CreateDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
	} else {
		m.CreateDate, err = time.Now(), nil
	}
	if err != nil {
		panic(err)
	}

	mimeType, ok := exifData["MIMEType"].(string)
	if !ok {
		panic(fmt.Errorf("refusing to parse file without MIMEType"))
	}
	m.MediaType = ParseMediaType(mimeType)

}

func (m *Media) makeTempReadableRaw() (string) {
	cmd := exec.Command("/Users/ethan/Downloads/LibRaw-0.21.1/bin/simple_dcraw", "-e", m.Filepath)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return (m.Filepath + ".thumb.jpg")

}

func (m *Media) ReadFullres() ([]byte) {
	var readableFilepath string
	if m.MediaType.IsRaw {
		readableFilepath = m.makeTempReadableRaw()
		defer os.Remove(readableFilepath)
	} else {
		readableFilepath = m.Filepath
	}

	mediaBytes, err := os.ReadFile(readableFilepath)
	if err != nil {
		panic(err)
	}

	return mediaBytes
}

func (m *Media) ReadFile() (image.Image) {
	var readableFilepath string
	if m.MediaType.IsRaw {
		readableFilepath = m.makeTempReadableRaw()
		defer os.Remove(readableFilepath)
	} else {
		readableFilepath = m.Filepath
	}

	file, err := os.Open(readableFilepath)
	if err != nil {
		panic(nil)
	}
	defer file.Close()

	i, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		panic(err)
	}

	file.Seek(0, 0)

	h := sha256.New()
	_, err = io.Copy(h, file)
	if err != nil {
		panic(err)
	}

	m.FileHash = base64.URLEncoding.EncodeToString(h.Sum(nil))

	return i

}

func (m *Media) calculateThumbSize(i image.Image) {
	dimentions := i.Bounds()
	width, height := dimentions.Dx(), dimentions.Dy()

	aspectRatio := float64(width) / float64(height)

	var newWidth, newHeight float64

	var bindSize = 800.0
	if aspectRatio > 1 {
		newWidth = bindSize
		newHeight = math.Floor(bindSize / aspectRatio)
	} else {
		newWidth = math.Floor(bindSize * aspectRatio)
		newHeight = bindSize
	}

	if newWidth == 0 || newHeight == 0 {
		panic(fmt.Errorf("thumbnail width or height is 0"))
	}
	m.ThumbWidth = int(newWidth)
	m.ThumbHeight = int(newHeight)

}

func (m *Media) GenerateBlurhash(thumb *image.NRGBA) {
	m.BlurHash, _ = blurhash.Encode(4, 3, thumb)
}

func (m *Media) GenerateThumbnail(i image.Image) (*image.NRGBA) {
	m.calculateThumbSize(i)
	thumb := imaging.Thumbnail(i, m.ThumbWidth, m.ThumbHeight, imaging.CatmullRom)

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		log.Error.Fatal(err)
	}

	thumbBytesBuf := new(bytes.Buffer)
	webp.Encode(thumbBytesBuf, thumb, options)

	m.Thumbnail64 = base64.StdEncoding.EncodeToString(thumbBytesBuf.Bytes())

	return thumb

}