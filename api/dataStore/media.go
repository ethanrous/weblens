package dataStore

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func (m *Media) IsFilledOut(skipThumbnail bool) (bool, string) {
	if m.MediaId == "" {
		return false, "filehash"
	}
	if m.FileId == "" {
		return false, "file id"
	}
	if m.Owner == "" {
		return false, "owner"
	}

	// Visual media specific properties
	if m.MediaType != nil && m.MediaType.IsDisplayable {

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

func (m *Media) LoadFileBytes() error {
	if m.thumbBytes != nil {
		return nil
	}

	mFile := m.GetBackingFile()
	if mFile == nil {
		return fmt.Errorf("failed to get file to read")
	}

	osFile, err := mFile.Read()
	if err != nil {
		return err
	}
	defer osFile.Close()

	size, _ := mFile.Size()
	bytes, err := util.OracleReader(osFile, size)
	if err != nil {
		return err
	}
	m.thumbBytes = bytes

	return nil
}

func (m *Media) extractExif() error {
	if gexift == nil {
		err := errors.New("exiftool not initialized")
		return err
	}
	mFile := m.GetBackingFile()
	if mFile.Err() != nil {
		return mFile.Err()
	}
	fileInfos := gexift.ExtractMetadata(mFile.String())
	if fileInfos[0].Err != nil {
		return fileInfos[0].Err
	}

	m.rawExif = fileInfos[0].Fields
	return nil
}

func (m *Media) ComputeExif() error {

	if m.CreateDate.Unix() == 0 && m.MediaType != nil && m.thumbBytes != nil {
		return nil
	}

	// We don't need the exif data once we leave this method.
	defer func() { m.rawExif = nil }()

	if m.rawExif == nil {
		// util.Warning.Println("Spawning lone exiftool for", FsTreeGet(m.FileId).String())
		err := m.extractExif()
		if err != nil {
			return err
		}
	}

	var err error
	if m.CreateDate.Unix() <= 0 {
		r, ok := m.rawExif["SubSecCreateDate"]
		if !ok {
			r, ok = m.rawExif["MediaCreateDate"]
		}
		if ok {
			m.CreateDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
			if err != nil {
				m.CreateDate, err = time.Parse("2006:01:02 15:04:05.00-07:00", r.(string))
			}
			if err != nil {
				m.CreateDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
			}
		} else {
			stat, err := os.Stat(m.GetBackingFile().String())
			if err != nil {
				return err
			}

			m.CreateDate = stat.ModTime()
		}
		if err != nil {
			return err
		}
	}

	if m.MediaType == nil {
		mimeType, ok := m.rawExif["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MediaType = ParseMimeType(mimeType)
	}

	if m.rotate == "" {
		rotate := m.rawExif["Orientation"]
		if rotate != nil {
			m.rotate = rotate.(string)
		}
	}

	if !m.MediaType.IsDisplayable || m.MediaType.RawThumbExifKey == "" {
		return nil
	}

	thumb64 := m.rawExif[m.MediaType.RawThumbExifKey].(string)
	thumb64 = thumb64[strings.Index(thumb64, ":")+1:]

	thumbBytes, err := base64.StdEncoding.DecodeString(thumb64)
	if err != nil {
		return err
	}
	m.thumbBytes = thumbBytes

	return nil
}

func (m *Media) DumpRawExif(rawExif map[string]any) {
	m.rawExif = rawExif
}

// func (m *Media) videoThumbnailReader() (io.Reader, error) {
// 	mFile := m.GetBackingFile()
// 	if mFile.Err() != nil {
// 		return nil, mFile.Err()
// 	}
// 	absolutePath := fmt.Sprintf("'%s'", mFile.String())
// 	cmd := exec.Command("ffmpeg", "-ss", "00:00:02.000", "-i", absolutePath, "-frames:v", "1", "-f", "mjpeg", "pipe:1")
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = cmd.Start()
// 	if err != nil {
// 		return nil, err
// 	}
// 	buf := new(bytes.Buffer)
// 	var bytesRead int64
// 	bytesRead, err = buf.ReadFrom(stdout)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cmd.Wait()
// 	if bytesRead == 0 {
// 		return nil, fmt.Errorf("did not read anything from ffmpeg stdout")
// 	}
// 	return buf, nil
// }

func (m *Media) ReadFullres() ([]byte, error) {
	sw := util.NewStopwatch("Read Fullres")

	var redisKey string
	if util.ShouldUseRedis() {
		redisKey = "Fullres " + m.MediaId
		fullres64, err := fddb.RedisCacheGet(redisKey)
		sw.Lap("Got from redis cache")
		if err == nil {
			fullresBytes, err := base64.StdEncoding.DecodeString(fullres64)
			sw.Lap("Decoded")
			sw.Stop()
			sw.PrintResults()
			return fullresBytes, err
		}
	}

	if m.image == nil {
		_, err := m.readFileIntoImage()
		if err != nil {
			return nil, err
		}
		sw.Lap("Read file into image")
	}

	buf := new(bytes.Buffer)
	webp.Encode(buf, m.image, nil)
	sw.Lap("Encoded webp image")
	fullresBytes := buf.Bytes()
	sw.Lap("Got Bytes")

	if util.ShouldUseRedis() {
		fullres64 := base64.StdEncoding.EncodeToString(fullresBytes)
		sw.Lap("Base64 encoded image")
		fddb.RedisCacheSet(redisKey, fullres64)
		sw.Lap("Set redis cache")
	}
	sw.Stop()
	sw.PrintResults()

	return fullresBytes, nil
}

func (m *Media) getReadable() (readable io.Reader, bytesLen int64, err error) {
	if m.thumbBytes == nil {
		util.Debug.Println("Getting readable")

		err = m.ComputeExif()
		if err != nil {
			return
		}

		if !m.MediaType.IsDisplayable {
			err = fmt.Errorf("cannot get readable of media that is not displayable")
			return
		}

		// When the media is not raw, the exifdata does not load
		// the thumb bytes into the struct, so we read the file manually
		if m.thumbBytes == nil {
			err = m.LoadFileBytes()
			if err != nil {
				return
			}
		}
	}

	readable = bytes.NewReader(m.thumbBytes)
	bytesLen = int64(len(m.thumbBytes))
	return
}

func (m *Media) readFileIntoImage() (i image.Image, err error) {
	readable, _, err := m.getReadable()
	if err != nil {
		return
	}

	i, err = imaging.Decode(readable)
	if err != nil {
		return
	}

	switch m.rotate {
	case "Rotate 270 CW":
		i = imaging.Rotate90(i)
	case "Rotate 90 CW":
		i = imaging.Rotate270(i)
	}

	m.image = i

	return
}

func (m *Media) generateFileHash() (err error) {
	readable, _, err := m.getReadable()
	if err != nil {
		return
	}
	if readable == nil {
		return fmt.Errorf("nil readable")
	}

	h := sha256.New()

	_, err = io.Copy(h, readable)
	if err != nil {
		util.DisplayError(err)
		return
	}

	m.MediaId = base64.URLEncoding.EncodeToString(h.Sum(nil))[:8]

	return
}

func (m *Media) Id() string {
	if m.MediaId == "" {
		if m.FileId == "" {
			err := errors.New("trying to generate mediaId for media with no FileId")
			util.DisplayError(err)
			return ""
		}
		err := m.generateFileHash()
		util.DisplayError(err)
	}

	return m.MediaId
}

func (m *Media) calculateThumbSize(i image.Image) {
	dimentions := i.Bounds()
	width, height := dimentions.Dx(), dimentions.Dy()
	m.MediaHeight = height
	m.MediaWidth = width

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

func (m *Media) GenerateThumbnail() (thumb *image.NRGBA, err error) {
	i, err := m.readFileIntoImage()
	if err != nil {
		return
	}

	m.calculateThumbSize(i)
	thumb = imaging.Thumbnail(i, m.ThumbWidth, m.ThumbHeight, imaging.Lanczos)

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		return
	}

	thumbBytesBuf := new(bytes.Buffer)
	webp.Encode(thumbBytesBuf, thumb, options)
	m.Thumbnail64 = base64.StdEncoding.EncodeToString(thumbBytesBuf.Bytes())

	return thumb, nil
}

func (m *Media) GetBackingFile() *WeblensFile {
	file := FsTreeGet(m.FileId)
	return file
}

func (m *Media) GetImage() image.Image {
	if m.image == nil {
		var err error
		m.image, err = m.readFileIntoImage()
		if err != nil {
			util.DisplayError(err)
		}
	}

	return m.image
}

// Toss the cached data for the media generated when parsing a file.
// This will drastically reduce memory usage if used properly
func (m *Media) Clean() {
	m.thumbBytes = nil
	m.rawExif = nil
	m.image = nil
}

func (m *Media) SetImported() {
	m.imported = true
}

func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}
	return m.imported
}
