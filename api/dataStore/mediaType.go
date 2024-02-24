package dataStore

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/util"
)

type mediaType struct {
	MimeType         string
	FriendlyName     string
	FileExtension    []string
	IsDisplayable    bool
	IsRaw            bool
	IsVideo          bool
	SupportsImgRecog bool
	RawThumbExifKey  string
}

var mediaTypeMap = map[string]mediaType{}
var displayableMap = map[string]mediaType{}

func InitMediaTypeMaps() error {
	typeJson, err := os.Open(filepath.Join(util.GetConfigDir(), "mediaType.json"))
	if err != nil {
		return err
	}
	defer typeJson.Close()

	typesBytes, err := io.ReadAll(typeJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(typesBytes, &mediaTypeMap)
	if err != nil {
		return err
	}

	initDisplayMap()

	return nil
}

func initDisplayMap() {
	for mime, mediaType := range mediaTypeMap {
		mediaType.MimeType = mime
		mediaTypeMap[mime] = mediaType
		for _, ext := range mediaType.FileExtension {
			displayableMap[ext] = mediaType
		}
	}
}

// Get a pointer to the weblens media type of a file given the mimeType
func ParseMimeType(mimeType string) *mediaType {
	mType, ok := mediaTypeMap[mimeType]
	if !ok {
		mType = mediaTypeMap["generic"]
		return &mType
	}
	return &mType
}

// Get a pointer to the weblens media type of a file given the file extension
func ParseExtType(ext string) *mediaType {
	var mType mediaType
	if ext == "" || displayableMap[ext].FriendlyName == "" {
		mType = mediaTypeMap["generic"]
	} else {
		mType = displayableMap[ext]
	}
	return &mType
}

func (f *WeblensFile) GetMediaType() (*mediaType, error) {
	if f.IsDir() {
		return nil, ErrDirNotAllowed
	}

	if f.media != nil && f.media.MediaType != nil {
		return f.media.MediaType, nil
	}

	m, err := f.GetMedia()
	if err != nil && err != ErrNoMedia {
		util.DisplayError(err)
		return nil, err
	}

	if m != nil && m.MediaType != nil {
		return m.MediaType, nil
	}
	err = ErrNoMedia

	mType := ParseExtType(f.Filename()[strings.Index(f.Filename(), ".")+1:])
	return mType, err
}

func (f *WeblensFile) IsDisplayable() (bool, error) {
	mType, err := f.GetMediaType()
	if mType == nil {
		return false, err
	}

	return mType.IsDisplayable, err
}
