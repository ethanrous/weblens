package interfaces

import "fmt"

type mediaType struct {
	FriendlyName string
	FileExtension []string
	IsRaw bool
	IsVideo bool
}

var mediaTypeMap = map[string]mediaType {
	"image/x-sony-arw": {"Sony ARW", []string{"ARW"}, true, false},
	"image/jpeg": {"Jpeg", []string{"jpeg", "jpg"}, false, false},
	"image/png": {"Png", []string{"png"}, false, false},
	"video/mp4": {"MP4", []string{"MP4"}, false, true},
}

func ParseMediaType(mimeType string) (mediaType, error) {
	mediaType, ok := mediaTypeMap[mimeType]
	if !ok {
		return mediaType, fmt.Errorf("unsupported mimeTYPE: %s", mimeType)
	}
	return mediaType, nil
}