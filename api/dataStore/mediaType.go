package dataStore

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

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

	marshMap := map[string]marshalableMediaType{}
	err = json.Unmarshal(typesBytes, &marshMap)
	if err != nil {
		return err
	}

	for _, key := range util.MapToKeys(marshMap) {
		typeEntry := marshalableToMediaType(marshMap[key])
		typeEntry.mimeType = key
		mediaTypeMap[key] = typeEntry

	}

	initDisplayMap()

	return nil
}

func GetMediaTypeMap() map[string]mediaType {
	return mediaTypeMap
}

func initDisplayMap() {
	for mime, mediaType := range mediaTypeMap {
		mediaType.mimeType = mime
		mediaTypeMap[mime] = mediaType
		for _, ext := range mediaType.fileExtension {
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
	if ext == "" || displayableMap[ext].friendlyName == "" {
		mType = mediaTypeMap["generic"]
	} else {
		mType = displayableMap[ext]
	}
	return &mType
}

func (mt mediaType) IsRaw() bool {
	return mt.isRaw
}

func (mt mediaType) IsMime(mime string) bool {
	return mt.mimeType == mime
}

func (mt *mediaType) IsDisplayable() bool {
	return mt.isDisplayable
}

func (mt *mediaType) FriendlyName() string {
	return mt.friendlyName
}

func (f *weblensFile) GetMediaType() (types.MediaType, error) {
	if f.IsDir() {
		return nil, ErrDirNotAllowed
	}
	m := MediaMapGet(f.GetContentId())
	if m != nil {
		mt := m.GetMediaType()
		if mt != nil {
			return mt, nil
		}
	}

	mType := ParseExtType(f.Filename()[strings.LastIndex(f.Filename(), ".")+1:])
	return mType, nil
}

func (f *weblensFile) IsDisplayable() bool {
	mType, _ := f.GetMediaType()
	if mType == nil {
		return false
	}

	return mType.IsDisplayable()
}

func (m *mediaType) toMarshalable() marshalableMediaType {
	return marshalableMediaType{
		MimeType:         m.mimeType,
		FriendlyName:     m.friendlyName,
		FileExtension:    m.fileExtension,
		IsDisplayable:    m.isDisplayable,
		IsRaw:            m.isRaw,
		IsVideo:          m.isVideo,
		SupportsImgRecog: m.supportsImgRecog,
		MultiPage:        m.multiPage,
		RawThumbExifKey:  m.rawThumbExifKey,
	}
}

func (m mediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.toMarshalable())
}

func marshalableToMediaType(m marshalableMediaType) mediaType {
	return mediaType{
		mimeType:         m.MimeType,
		friendlyName:     m.FriendlyName,
		fileExtension:    m.FileExtension,
		isDisplayable:    m.IsDisplayable,
		isRaw:            m.IsRaw,
		isVideo:          m.IsVideo,
		supportsImgRecog: m.SupportsImgRecog,
		multiPage:        m.MultiPage,
		rawThumbExifKey:  m.RawThumbExifKey,
	}
}
