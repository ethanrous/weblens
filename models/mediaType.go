package models

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/ethrousseau/weblens/internal"
)

type MediaType struct {
	Mime            string   `json:"mime"`
	Name            string   `json:"FriendlyName"`
	Extensions      []string `json:"FileExtension"`
	Displayable     bool     `json:"IsDisplayable"`
	Raw             bool     `json:"IsRaw"`
	Video           bool     `json:"IsVideo"`
	ImgRecog        bool     `json:"SupportsImgRecog"`
	MultiPage       bool     `json:"MultiPage"`
	RawThumbExifKey string   `json:"RawThumbExifKey"`
}

type typeService struct {
	mimeMap map[string]MediaType
	extMap  map[string]MediaType
}

func NewTypeService() MediaTypeService {

	ts := &typeService{
		mimeMap: make(map[string]MediaType),
		extMap:  make(map[string]MediaType),
	}

	// Only from config file, for now
	typeJson, err := os.Open(filepath.Join(internal.GetConfigDir(), "mediaType.json"))
	if err != nil {
		panic(err)
	}
	defer func(typeJson *os.File) {
		err := typeJson.Close()
		if err != nil {
			panic(err)
		}
	}(typeJson)

	typesBytes, err := io.ReadAll(typeJson)
	marshMap := map[string]MediaType{}
	err = json.Unmarshal(typesBytes, &marshMap)
	if err != nil {
		panic(err)
	}

	for k, t := range marshMap {
		t.Mime = k
		ts.mimeMap[k] = t
	}

	for _, mt := range ts.mimeMap {
		for _, ext := range mt.Extensions {
			ts.extMap[ext] = mt
		}
	}

	return ts
}

// ParseExtension Get a pointer to the weblens Media type of a file given the file extension
func (ts *typeService) ParseExtension(ext string) MediaType {
	if len(ext) == 0 {
		return ts.mimeMap["generic"]
	}

	if ext[0] == '.' {
		ext = ext[1:]
	}
	mt, ok := ts.extMap[ext]
	if !ok {
		return ts.mimeMap["generic"]
	}
	return mt
}

func (ts *typeService) ParseMime(mime string) MediaType {
	return ts.mimeMap[mime]
}

func (ts *typeService) Generic() MediaType {
	return ts.mimeMap["generic"]
}

func (ts *typeService) Size() int {
	return len(ts.mimeMap)
}

func (ts *typeService) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.mimeMap)
}

func (mt MediaType) IsRaw() bool {
	return mt.Raw
}

func (mt MediaType) IsVideo() bool {
	return mt.Video
}

func (mt MediaType) GetMime() string {
	return mt.Mime
}

func (mt MediaType) IsMime(mime string) bool {
	return mt.Mime == mime
}

func (mt MediaType) IsDisplayable() bool {
	return mt.Displayable
}

func (mt MediaType) FriendlyName() string {
	return mt.Name
}

func (mt MediaType) IsSupported() bool {
	return mt.Mime != "generic"
}

func (mt MediaType) IsMultiPage() bool {
	return mt.MultiPage
}

func (mt MediaType) GetThumbExifKey() string {
	return mt.RawThumbExifKey
}

func (mt MediaType) SupportsImgRecog() bool {
	return mt.ImgRecog
}

// func (m *mediaType) toMarshalable() dataStore.marshalableMediaType {
// 	return dataStore.marshalableMediaType{
// 		MimeType:         m.mimeType,
// 		FriendlyName:     m.friendlyName,
// 		Extensions:    m.fileExtension,
// 		IsDisplayable:    m.isDisplayable,
// 		IsRaw:            m.isRaw,
// 		IsVideo:          m.isVideo,
// 		SupportsImgRecog: m.supportsImgRecog,
// 		MultiPage:        m.multiPage,
// 		RawThumbExifKey:  m.rawThumbExifKey,
// 	}
// }
//
// func (m mediaType) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(m.toMarshalable())
// }
//
// func marshalableToMediaType(m dataStore.marshalableMediaType) mediaType {
// 	return mediaType{
// 		mimeType:         m.MimeType,
// 		friendlyName:     m.FriendlyName,
// 		fileExtension:    m.Extensions,
// 		isDisplayable:    m.IsDisplayable,
// 		isRaw:            m.IsRaw,
// 		isVideo:          m.IsVideo,
// 		supportsImgRecog: m.SupportsImgRecog,
// 		multiPage:        m.MultiPage,
// 		rawThumbExifKey:  m.RawThumbExifKey,
// 	}
// }

type MediaTypeService interface {
	ParseExtension(ext string) MediaType
	ParseMime(mime string) MediaType
	Generic() MediaType
	Size() int
}
