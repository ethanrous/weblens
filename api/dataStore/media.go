package dataStore

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	// mongoId			string					`bson:"_id"`
	FileHash		string					`bson:"fileHash" json:"fileHash"`
	ParentFolder 	string 					`bson:"parentFolderId" json:"parentFolder"`
	Filename	 	string 					`bson:"filename" json:"filename"`
	MediaType		mediaType				`bson:"mediaType" json:"mediaType"`
	BlurHash 		string 					`bson:"blurHash" json:"blurHash"`
	Thumbnail64 	string		 			`bson:"thumbnail" json:"thumbnail64"`
	MediaWidth 		int						`bson:"width" json:"mediaWidth"`
	MediaHeight 	int 					`bson:"height" json:"mediaHeight"`
	ThumbWidth 		int						`bson:"thumbWidth" json:"thumbWidth"`
	ThumbHeight 	int 					`bson:"thumbHeight" json:"thumbHeight"`
	CreateDate		time.Time				`bson:"createDate" json:"createDate"`
	Owner			string					`bson:"owner" json:"owner"`
	SharedWith		[]primitive.ObjectID	`bson:"sharedWith" json:"sharedWith"`

	// rotation		int
	image 			image.Image
	rawExif			map[string]any
	thumbBytes		[]byte
}

func (m Media) MarshalBinary() ([]byte, error) {
    return json.Marshal(m)
}

func (m *Media) IsFilledOut(skipThumbnail bool) (bool, string) {
	if m.FileHash == "" {
		return false, "filehash"
	}
	if m.ParentFolder == "" {
		return false, "parent folder"
	}
	if m.Filename == "" {
		return false, "filename"
	}
	if m.MediaType.FriendlyName == "" {
		return false, "friendly name"
	}

	if (m.Owner == "") {
		return false, "owner"
	}

	// Visual media specific properties
	if m.MediaType.IsDisplayable {

		if m.BlurHash == "" {
			return false, "blurhash"
		}
		if !skipThumbnail && m.Thumbnail64 == "" {
			return false, "thumbnail"
		}
		if m.MediaWidth == 0 {
			return false, "media width"
		}
		if m.MediaHeight == 0 {
			return false, "media height"
		}
		if m.ThumbWidth == 0 {
			return false, "thumb width"
		}
		if m.ThumbHeight == 0 {
			return false, "thumb height"
		}
	}

	if m.CreateDate.IsZero() {
		return false, "create date"
	}

	return true, ""

}

func (m *Media) extractExif() error {
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"))
	if err != nil {
		return err
	}
	defer et.Close()
	mFile := m.GetBackingFile()
	if mFile.Err() != nil {
		return mFile.Err()
	}
	fileInfos := et.ExtractMetadata(mFile.String())
	if fileInfos[0].Err != nil {
		return fileInfos[0].Err
	}

	m.rawExif = fileInfos[0].Fields
	return nil
}

func (m *Media) ComputeExif() (error) {
	if m.rawExif == nil {
		util.Warning.Println("Spawning lone exiftool for", m.Filename)
		err := m.extractExif()
		if err != nil {
			return err
		}
	}

	var err error
	r, ok := m.rawExif["SubSecCreateDate"]
	if !ok {
		r, ok = m.rawExif["MediaCreateDate"]
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
		return err
	}

	mimeType, ok := m.rawExif["MIMEType"].(string)
	if !ok {
		mimeType = "generic"
	}
	m.MediaType, _ = ParseMediaType(mimeType)

	if !m.MediaType.IsDisplayable {
		return nil
	}

	ok = true
	var dimentionsStr string
	if m.MediaType.IsVideo {
		d, k := m.rawExif["VideoSize"]
		ok = k
		if k {
			dimentionsStr = d.(string)
		}
	}
	if !m.MediaType.IsVideo || !ok {
		d, ok := m.rawExif["ImageSize"]
		if ok {
			dimentionsStr = d.(string)
		} else {
			return fmt.Errorf("did not find expected image dimentions from exif data")
		}
	}
	dimentionsList := strings.Split(dimentionsStr, "x")
	m.MediaHeight, err = strconv.Atoi(dimentionsList[0])
	if err != nil {
		return err
	}
	m.MediaWidth, err = strconv.Atoi(dimentionsList[1])
	if err != nil {
		return err
	}

	return nil
}

func (m *Media) DumpRawExif(rawExif map[string]any) {
	m.rawExif = rawExif
}

func (m *Media) DumpThumbBytes(thumbBytes []byte) {
	m.thumbBytes = thumbBytes
}

func (m *Media) rawImageReader() (io.Reader, error) {
	mFile := m.GetBackingFile()
	if mFile.Err() != nil {
		return nil, mFile.Err()
	}
	escapedPath := strings.ReplaceAll(mFile.String(), " ", "\\ ")
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

	r := bytes.NewReader(out.Bytes())

	// i, err := imaging.Decode(r)
	// util.FailOnError(err, "Failed to read exiftool jpeg output into image")

	// buf := new(bytes.Buffer)
	// err = jpeg.Encode(buf, i, nil)
	// util.FailOnError(err, "Failed to convert image to jpeg")

	// return buf, nil
	return r, nil

}

func (m *Media) videoThumbnailReader() (io.Reader, error) {
	mFile := m.GetBackingFile()
	if mFile.Err() != nil {
		return nil, mFile.Err()
	}
	absolutePath := mFile.String()
	cmd := exec.Command("ffmpeg", "-ss", "00:00:02.000", "-i", absolutePath, "-frames:v", "1", "-f", "mjpeg", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	var bytesRead int64
	bytesRead, err = buf.ReadFrom(stdout)
	if err != nil {
		return nil, err
	}

	cmd.Wait()
	if bytesRead == 0 {
		return nil, fmt.Errorf("did not read anything from ffmpeg stdout")
	}
	return buf, nil
}

func (m *Media) ReadFullres(db *Weblensdb) ([]byte, error) {
	defer util.RecoverPanic()

	redisKey := "Fullres " + m.FileHash
	if util.ShouldUseRedis() {
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
		mFile:= m.GetBackingFile()
		if mFile.Err() != nil {
			return nil, mFile.Err()
		}
		readable, err = mFile.Read()
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
	if m.thumbBytes != nil {
		readable = bytes.NewReader(m.thumbBytes)
	} else if m.MediaType.IsRaw {
		var err error
		readable, err = m.rawImageReader()
		util.FailOnError(err, "Failed to get readable raw proxy image")
	} else if m.MediaType.IsVideo {
		var err error
		readable, err = m.videoThumbnailReader()
		util.FailOnError(err, "Failed to get readable video proxy image")
	} else {
		var err error
		mFile := m.GetBackingFile()
		util.FailOnError(mFile.Err(), "Failed to get generic media backing file")
		readable, err = mFile.Read()
		util.FailOnError(err, "Failed to get generic media readable")
	}

	return readable
}

func (m *Media) readFileIntoImage() (image.Image) {
	readable := m.getReadable()

	i, err := imaging.Decode(readable)

	util.FailOnError(err, "Failed to decode readable proxy image buffer")
	if m.rawExif["Orientation"] == "Rotate 270 CW" {
		m.MediaHeight, m.MediaWidth = m.MediaWidth, m.MediaHeight
		i = imaging.Rotate90(i)
	}

	return i
}

func (m *Media) GenerateFileHash(mFile *WeblensFileDescriptor) {
	readable := m.getReadable()

	h := sha256.New()
	if mFile.IsDisplayable() {
		_, err := io.Copy(h, readable)
		util.FailOnError(err, "Failed to copy readable image into hash")
	}

	h.Write([]byte(mFile.String())) // Make exact same files in different locations have unique id's

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
	thumb := imaging.Thumbnail(i, m.ThumbWidth, m.ThumbHeight, imaging.Lanczos)

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		util.Error.Fatal(err)
	}
	thumbBytesBuf := new(bytes.Buffer)
	webp.Encode(thumbBytesBuf, thumb, options)
	m.Thumbnail64 = base64.StdEncoding.EncodeToString(thumbBytesBuf.Bytes())

	return thumb
}

func (m *Media) GetBackingFile() *WeblensFileDescriptor {
	file := GetWFD(m.ParentFolder, m.Filename)
	return file
}

func (m *Media) GetImage() image.Image {
	if m.image == nil {
		m.image = m.readFileIntoImage()
	}

	return m.image
}