package interfaces

import "fmt"

type mediaType struct {
	FriendlyName string
	FileExtension []string
	IsRaw bool
}

var mediaTypeMap = map[string]mediaType {
	"image/x-sony-arw": {"Sony ARW", []string{"ARW"}, true},
	"image/jpeg": {"Jpeg", []string{"jpeg", "jpg"}, false},
}

func ParseMediaType(mimeType string) (mediaType) {
	mediaType, ok := mediaTypeMap[mimeType]
	if !ok {
		panic(fmt.Errorf("unsupported mimeTYPE: %s", mimeType))
	}
	return mediaType
}