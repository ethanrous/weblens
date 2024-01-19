package dataStore

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type mediaType struct {
	FriendlyName    string
	FileExtension   []string
	IsDisplayable   bool
	IsRaw           bool
	IsVideo         bool
	RawThumbExifKey string
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
	for _, mediaType := range mediaTypeMap {
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

// type DirNotAllowed error
// type NoMedia error

var DirNotAllowed = errors.New("directory not allowed")
var NoMedia = errors.New("no media found")

func (f *WeblensFile) IsDisplayable() (bool, error) {
	if f.IsDir() {
		return false, DirNotAllowed
	}

	m, err := f.GetMedia()
	if err != nil && err != mongo.ErrNoDocuments {
		util.DisplayError(err)
		return false, err
	}

	if m != nil && m.MediaType != nil {
		return m.MediaType.IsDisplayable, nil
	}
	err = NoMedia

	mType := ParseExtType(f.Filename()[strings.Index(f.Filename(), ".")+1:])
	return mType.IsDisplayable, err
}
