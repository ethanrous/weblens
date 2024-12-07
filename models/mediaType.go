package models

type MediaType struct {
	Mime            string   `json:"mime"`
	Name            string   `json:"FriendlyName"`
	RawThumbExifKey string   `json:"RawThumbExifKey"`
	Extensions      []string `json:"FileExtension"`
	Displayable     bool     `json:"IsDisplayable"`
	Raw             bool     `json:"IsRaw"`
	Video           bool     `json:"IsVideo"`
	ImgRecog        bool     `json:"SupportsImgRecog"`
	MultiPage       bool     `json:"MultiPage"`
} // @name MediaType

type typeService struct {
	mimeMap map[string]MediaType
	extMap  map[string]MediaType
}

func NewTypeService(marshMap map[string]MediaType) MediaTypeService {

	ts := &typeService{
		mimeMap: make(map[string]MediaType),
		extMap:  make(map[string]MediaType),
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

func (ts *typeService) GetMaps() (map[string]MediaType, map[string]MediaType) {
	return ts.mimeMap, ts.extMap
}

func (ts *typeService) Generic() MediaType {
	return ts.mimeMap["generic"]
}

func (ts *typeService) Size() int {
	return len(ts.mimeMap)
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
	return mt.Mime != "generic" && mt.Mime != ""
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

type MediaTypeService interface {
	GetMaps() (map[string]MediaType, map[string]MediaType)
	ParseExtension(ext string) MediaType
	ParseMime(mime string) MediaType
	Generic() MediaType
	Size() int
}
