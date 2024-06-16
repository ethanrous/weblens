package media

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type mediaType struct {
	Mime            string
	Name            string
	Extensions      []string
	Displayable     bool
	Raw             bool
	IsVideo         bool
	ImgRecog        bool
	MultiPage       bool
	RawThumbExifKey string
}

type typeService struct {
	mimeMap map[string]types.MediaType
	extMap  map[string]types.MediaType
}

func NewTypeService() types.MediaTypeService {

	service := &typeService{
		mimeMap: make(map[string]types.MediaType),
		extMap:  make(map[string]types.MediaType),
	}

	// Only from config file, for now
	typeJson, err := os.Open(filepath.Join(util.GetConfigDir(), "mediaType.json"))
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
	marshMap := map[string]*mediaType{}
	err = json.Unmarshal(typesBytes, &marshMap)
	if err != nil {
		panic(err)
	}

	for k, t := range marshMap {
		t.Mime = k
		service.mimeMap[k] = t
	}

	for mime, mt := range service.mimeMap {
		realMt := mt.(*mediaType)
		realMt.Mime = mime
		service.mimeMap[mime] = mt
		for _, ext := range realMt.Extensions {
			service.extMap[ext] = mt
		}
	}

	return &typeService{}
}

// Get a pointer to the weblens Media type of a file given the file extension
func (ts *typeService) ParseExtension(ext string) types.MediaType {
	if ext == "" || ts.extMap[ext].FriendlyName() == "" {
		return ts.mimeMap["generic"]
	} else {
		return ts.extMap[ext]
	}
}

func (ts *typeService) ParseMime(mime string) types.MediaType {
	return ts.mimeMap[mime]
}

func (ts *typeService) Generic() types.MediaType {
	return ts.mimeMap["generic"]
}

func (ts *typeService) Size() int {
	return len(ts.mimeMap)
}

func (mt mediaType) IsRaw() bool {
	return mt.Raw
}

func (mt mediaType) GetMime() string {
	return mt.Mime
}

func (mt mediaType) IsMime(mime string) bool {
	return mt.Mime == mime
}

func (mt *mediaType) IsDisplayable() bool {
	return mt.Displayable
}

func (mt *mediaType) FriendlyName() string {
	return mt.Name
}

func (mt *mediaType) IsMultiPage() bool {
	return mt.MultiPage
}

func (mt *mediaType) GetThumbExifKey() string {
	return mt.RawThumbExifKey
}

func (mt *mediaType) SupportsImgRecog() bool {
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
