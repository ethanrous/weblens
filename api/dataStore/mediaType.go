package dataStore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ethrousseau/weblens/api/util"
)

type mediaType struct {
	FriendlyName string
	FileExtension []string
	IsDisplayable bool
	IsRaw bool
	IsVideo bool
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

func ParseMediaType(mimeType string) (mediaType, error) {
	mediaType, ok := mediaTypeMap[mimeType]
	if !ok {
		return mediaTypeMap["generic"], fmt.Errorf("unsupported filetype: %s, falling back to generic type", mimeType)
	}
	return mediaType, nil
}

func (f *WeblensFileDescriptor) getMediaType() mediaType {
	if f.media != nil && f.media.MediaType.FriendlyName != "" {
		return f.media.MediaType
	}

	if len(displayableMap) == 0 {
		initDisplayMap()
	}

	ext := filepath.Ext(f.Filename())
	var mType mediaType
	if ext == "" || displayableMap[ext[1:]].FriendlyName == "" {
		mType = mediaTypeMap["generic"]
	} else {
		mType = displayableMap[ext[1:]]
	}
	if f.media != nil {
		f.media.MediaType = mType
	}
	return mType
}

func (f *WeblensFileDescriptor) IsDisplayable() bool {
	// s, err := json.Marshal(mediaTypeMap)
	// util.DisplayError(err)
	// if s != nil {
	// 	util.Debug.Println(string(s))
	// }
	return f.getMediaType().IsDisplayable
}