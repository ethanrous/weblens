package interfaces

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/tiff"

	util "github.com/ethrousseau/weblens/api/utils"
)

type Media struct {
	FileHash		string				`bson:"_id"`
	Filepath 		string 				`bson:"filepath"`
	MediaType		mediaType			`bson:"mediaType"`
	BlurHash 		string 				`bson:"blurHash"`
	Thumbnail64 	string		 		`bson:"thumbnail"`
	MediaWidth 		int					`bson:"width"`
	MediaHeight 	int 				`bson:"height"`
	ThumbWidth 		int					`bson:"thumbWidth"`
	ThumbHeight 	int 				`bson:"thumbHeight"`
	CreateDate		time.Time			`bson:"createDate"`
}

func (m Media) MarshalBinary() ([]byte, error) {
    return json.Marshal(m)
}


func (m *Media) ExtractExif() (error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		panic(err)
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(m.Filepath)
	if fileInfos[0].Err != nil {
		util.Debug.Panicf("Cound not extract metadata for %s: %s", m.Filepath, fileInfos[0].Err)
	}

	exifData := fileInfos[0].Fields

	r, ok := exifData["SubSecCreateDate"]
	if !ok {
		r, ok = exifData["MediaCreateDate"]
	}
	if ok {
		m.CreateDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
		if err != nil {
			m.CreateDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
		}
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
	m.MediaType, err = ParseMediaType(mimeType)
	if err != nil {
		return err
	}

	var dimentions string
	if m.MediaType.IsVideo {
		dimentions = exifData["VideoSize"].(string)
		} else {
		dimentions = exifData["ImageSize"].(string)
	}
	dimentionsList := strings.Split(dimentions, "x")
	m.MediaHeight, _ = strconv.Atoi(dimentionsList[0])
	m.MediaWidth, _ = strconv.Atoi(dimentionsList[1])

	return nil
}

func (m *Media) rawImageReader() (io.Reader, error) {
	absolutePath := util.GuaranteeAbsolutePath(m.Filepath)
	cmd := exec.Command("/Users/ethan/Downloads/LibRaw-0.21.1/bin/dcraw_emu", "-mem", "-T", "-Z", "-", absolutePath)
	//cmd := exec.Command("/Users/ethan/Downloads/LibRaw-0.21.1/bin/dcraw_emu", "-mem", "-T", absolutePath)


	stdout, err := cmd.StdoutPipe()
	if err != nil {
		util.Debug.Panicf("could not get tiff stdout: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		util.Debug.Panicf("could not start tiff build cmd: %s", err)
	}
	buf := new (bytes.Buffer)
	_, err = buf.ReadFrom(stdout)
	if err != nil {
		return nil, err
	}
	cmd.Wait()

	i, err := tiff.Decode(buf)
	if err != nil {
		util.Debug.Panicf("could not read dcraw tiff output into go image: %s", err)
	}



	buf = new(bytes.Buffer)
	err = jpeg.Encode(buf, i, nil)
	if err != nil {
		util.Debug.Panicf("could not convert image to jpeg: %s", err)
	}

	return buf, nil

}

func (m *Media) videoFileThumbBytes() (io.Reader) {

	cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-i", m.Filepath, "-ss", "00:00:02.000", "-frames:v", "1", "-")
	stdout, _ := cmd.StdoutPipe()
	err := cmd.Run()

	if err != nil {
		panic(err)
	}

	return stdout

}

func (m *Media) ReadFullres() ([]byte) {
	var readable io.Reader
	if m.MediaType.IsRaw {
		readable, _ = m.rawImageReader()
	} else {
		var err error
		readable, err = os.Open(m.Filepath)
		if err != nil {
			panic(err)
		}
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(readable)
	if err != nil {
		util.Debug.Panicf("could not open full-res file: %s", m.Filepath)
	}

	return buf.Bytes()
}

func (m *Media) ReadFile() (image.Image) {
	var readable io.Reader
	if m.MediaType.IsRaw {
		readable, _ = m.rawImageReader()
	} else if m.MediaType.IsVideo {
		readable = m.videoFileThumbBytes()
	} else {
		var err error
		readable, err = os.Open(m.Filepath)
		if err != nil {
			panic(err)
		}
	}

	i, err := imaging.Decode(readable, imaging.AutoOrientation(true))
	if err != nil {
		panic(err)
	}

	h := sha256.New()
	_, err = io.Copy(h, readable)
	if err != nil {
		panic(err)
	}

	h.Write([]byte(m.Filepath)) // Make exact same files in different locations have unique id's

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
		util.Error.Fatal(err)
	}

	thumbBytesBuf := new(bytes.Buffer)
	webp.Encode(thumbBytesBuf, thumb, options)

	m.Thumbnail64 = base64.StdEncoding.EncodeToString(thumbBytesBuf.Bytes())

	return thumb

}