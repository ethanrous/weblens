package dataStore

import "fmt"

type mediaType struct {
	FriendlyName string
	FileExtension []string
	IsRaw bool
	IsVideo bool
}

var mediaTypeMap = map[string]mediaType {
	"image/x-sony-arw": {"Sony ARW", []string{"ARW"}, true, false},
	"image/x-nikon-nef": {"Nikon Raw", []string{"NEF"}, true, false},
	"image/jpeg": {"Jpeg", []string{"jpeg", "jpg"}, false, false},
	"image/png": {"Png", []string{"png"}, false, false},
	"image/gif": {"Gif", []string{"gif"}, false, false},
	"video/mp4": {"MP4", []string{"MP4"}, false, true},
	"generic": {"File", []string{}, false, false},
}

func ParseMediaType(mimeType string) (mediaType, error) {
	mediaType, ok := mediaTypeMap[mimeType]
	if !ok {
		return mediaTypeMap["generic"], fmt.Errorf("unsupported filetype: %s, falling back to generic type", mimeType)
	}
	return mediaType, nil
}