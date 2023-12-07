package dataStore

import (
	"fmt"
	"path/filepath"
)

type mediaType struct {
	FriendlyName string
	FileExtension []string
	IsDisplayable bool
	IsRaw bool
	IsVideo bool
}

var mediaTypeMap = map[string]mediaType {
	"image/x-sony-arw": 	{FriendlyName: "Sony ARW", FileExtension: []string{"ARW"}, IsDisplayable: true, IsRaw: true, IsVideo: false},
	"image/x-nikon-nef": 	{FriendlyName: "Nikon Raw", FileExtension: []string{"NEF"}, IsDisplayable: true, IsRaw: true, IsVideo: false},
	"image/jpeg": 			{FriendlyName: "Jpeg", FileExtension: []string{"jpeg", "jpg"}, IsDisplayable: true, IsRaw: false, IsVideo: false},
	"image/png": 			{FriendlyName: "Png", FileExtension: []string{"png"}, IsDisplayable: true, IsRaw: false, IsVideo: false},
	"image/gif": 			{FriendlyName: "Gif", FileExtension: []string{"gif"},IsDisplayable: true, IsRaw: false,IsVideo: false},
	"video/mp4": 			{FriendlyName: "MP4",FileExtension: []string{"MP4"},IsDisplayable: true,IsRaw: false,IsVideo: true},
	"application/zip": 		{FriendlyName: "Zip",FileExtension: []string{"zip"},IsDisplayable: false, IsRaw: false, IsVideo: false},
	"generic": 				{FriendlyName: "File", FileExtension: []string{}, IsDisplayable: false, IsRaw: false, IsVideo: false},
}

var displayableMap = map[string]mediaType{}

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
	if len(displayableMap) == 0 {
		initDisplayMap()
	}
	ext := filepath.Ext(f.Filename)
	var mType mediaType
	if ext == "" || displayableMap[ext[1:]].FriendlyName == "" {
		mType = mediaTypeMap["generic"]
	} else {
		mType = displayableMap[ext[1:]]
	}
	return mType
}

func (f *WeblensFileDescriptor) IsDisplayable() bool {
	return f.getMediaType().IsDisplayable
}