package dataStore

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
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
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

func (m *Media) IsFilledOut(skipThumbnail bool) (bool) {
	if m.FileHash == "" {
		return false
	}
	if m.Filepath == "" {
		return false
	}
	if m.MediaType.FriendlyName == "" {
		return false
	}

	// Visual media specific properties
	if m.MediaType.FriendlyName != "File" {

		if m.BlurHash == "" {
			return false
		}
		if !skipThumbnail && m.Thumbnail64 == "" {
			return false
		}
		if m.MediaWidth == 0 {
			return false
		}
		if m.MediaHeight == 0 {
			return false
		}
		if m.ThumbWidth == 0 {
			return false
		}
		if m.ThumbHeight == 0 {
			return false
		}
	}

	if m.CreateDate.IsZero() {
		return false
	}

	return true

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
	m.MediaType, _ = ParseMediaType(mimeType)
	// if err != nil {
	// 	return err
	// }

	if m.MediaType.FriendlyName == "File" {
		return nil
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
	absolutePath := GuaranteeAbsolutePath(m.Filepath)

	escapedPath := strings.ReplaceAll(absolutePath, " ", "\\ ")

	//tmpFile, err := os.CreateTemp(util.GetTmpDir(), "*.jpeg")
	//util.FailOnError(err, "Could not create temp jpeg file")

	cmdString := fmt.Sprintf("exiftool -a -b -JpgFromRaw %s | exiftool -tagsfromfile %s -Orientation -", escapedPath, escapedPath)
	// cmdString := fmt.Sprintf("exiftool -a -b -JpgFromRaw %s", escapedPath)

	cmd := exec.Command("/bin/bash", "-c", cmdString)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		util.FailOnError(err, "Failed to run exiftool extract command")
	}

	// util.Debug.Printf("Buf size: %d", readBytes)
	r := bytes.NewReader(out.Bytes())

	i, err := imaging.Decode(r, imaging.AutoOrientation(true))
	util.FailOnError(err, "Failed to read exiftool jpeg output into image")

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, i, nil)
	util.FailOnError(err, "Failed to convert image to jpeg")

	return buf, nil

}

func (m *Media) videoThumbnailReader() (io.Reader, error) {

	cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-i", GuaranteeAbsolutePath(m.Filepath), "-ss", "00:00:02.000", "-frames:v", "1", "-f", "mjpeg", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	util.FailOnError(err, "Failed to get ffmpeg stdout pipe")

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	buf := new (bytes.Buffer)
	_, err = buf.ReadFrom(stdout)

	cmd.Wait()

	util.FailOnError(err, "Failed to run ffmpeg to get video thumbnail")

	return buf, nil

}

func (m *Media) ReadFullres(db Weblensdb) ([]byte, error) {
	defer func() any {
		if err := recover(); err != nil {
			util.Error.Printf("Could not read fullres image: %s", err)
			return err
		}
		return nil
	}()

	redisKey := "Fullres " + m.FileHash
	if util.ShouldUseRedis() {
		db := NewDB()
		fullres64, err := db.RedisCacheGet(redisKey)
		if err == nil {
			fullresBytes, err := base64.StdEncoding.DecodeString(fullres64)
			util.FailOnError(err, "Failed to decode fullres image base64 string")

			return fullresBytes, nil
		}
	}

	var readable io.Reader
	if m.MediaType.IsRaw {
		var err error
		readable, err = m.rawImageReader()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		readable, err = os.Open(GuaranteeAbsolutePath(m.Filepath))
		if err != nil {
			return nil, err
		}
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(readable)
	util.FailOnError(err, "Failed to read fullres image from buffer")

	fullresBytes := buf.Bytes()

	if util.ShouldUseRedis() {
		fullres64 := base64.StdEncoding.EncodeToString(fullresBytes)
		db.RedisCacheSet(redisKey, fullres64)
	}

	return fullresBytes, nil
}

func (m *Media) getReadable() (io.Reader) {
	var readable io.Reader
	if m.MediaType.IsRaw {
		var err error
		readable, err = m.rawImageReader()
		util.FailOnError(err, "Failed to get readable raw proxy image")
	} else if m.MediaType.IsVideo {
		var err error
		readable, err = m.videoThumbnailReader()
		util.FailOnError(err, "Failed to get readable video proxy image")
	} else {
		var err error
		readable, err = os.Open(GuaranteeAbsolutePath(m.Filepath))
		util.FailOnError(err, "Failed to open generic image file")
	}

	return readable
}

func (m *Media) readFileIntoImage() (image.Image) {
	readable := m.getReadable()

	i, err := imaging.Decode(readable, imaging.AutoOrientation(true))
	util.FailOnError(err, "Failed to decode readable proxy image buffer")

	return i
}

func (m *Media) GenerateFileHash() {
	readable := m.getReadable()

	h := sha256.New()
	_, err := io.Copy(h, readable)
	util.FailOnError(err, "Failed to copy readable image into hash")

	h.Write([]byte(m.Filepath)) // Make exact same files in different locations have unique id's

	m.FileHash = base64.URLEncoding.EncodeToString(h.Sum(nil))
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

func (m *Media) GenerateThumbnail() (*image.NRGBA) {
	i := m.readFileIntoImage()

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