package media

import "encoding/json"

const MediaTypeJson = `{
    "application/zip": {
        "FriendlyName": "Zip",
        "FileExtension": [
            "zip"
        ],
        "IsDisplayable": false,
        "IsRaw": false,
        "IsVideo": false
    },
    "generic": {
        "FriendlyName": "File",
        "FileExtension": [],
        "IsDisplayable": false,
        "IsRaw": false,
        "IsVideo": false,
        "SupportsImgRecog": false
    },
    "image/gif": {
        "FriendlyName": "Gif",
        "FileExtension": [
            "gif"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "SupportsImgRecog": false
    },
    "image/jpeg": {
        "FriendlyName": "Jpeg",
        "FileExtension": [
            "jpeg",
            "jpg",
            "JPG"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "SupportsImgRecog": true
    },
    "image/png": {
        "FriendlyName": "Png",
        "FileExtension": [
            "png"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "SupportsImgRecog": true
    },
    "image/x-nikon-nef": {
        "FriendlyName": "Nikon Raw",
        "FileExtension": [
            "NEF",
            "nef"
        ],
        "IsDisplayable": true,
        "IsRaw": true,
        "IsVideo": false,
        "RawThumbExifKey": "JpgFromRaw",
        "SupportsImgRecog": true
    },
    "image/x-sony-arw": {
        "FriendlyName": "Sony ARW",
        "FileExtension": [
            "ARW"
        ],
        "IsDisplayable": true,
        "IsRaw": true,
        "IsVideo": false,
        "RawThumbExifKey": "PreviewImage",
        "SupportsImgRecog": true
    },
    "image/x-canon-cr2": {
        "FriendlyName": "Cannon Raw",
        "FileExtension": [
            "CR2"
        ],
        "IsDisplayable": true,
        "IsRaw": true,
        "IsVideo": false,
        "RawThumbExifKey": "PreviewImage",
        "SupportsImgRecog": true
    },
    "image/heic": {
        "FriendlyName": "HEIC",
        "FileExtension": [
            "HEIC",
            "heic"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "RawThumbExifKey": "",
        "SupportsImgRecog": true
    },
    "image/heif": {
        "FriendlyName": "HEIF",
        "FileExtension": [
            "HEIF",
            "heif"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "RawThumbExifKey": "",
        "SupportsImgRecog": true
    },
    "image/webp": {
        "FriendlyName": "webp",
        "FileExtension": [
            "webp"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "RawThumbExifKey": "",
        "SupportsImgRecog": true
    },
    "application/pdf": {
        "FriendlyName": "PDF",
        "FileExtension": [
            "pdf"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": false,
        "RawThumbExifKey": "",
        "SupportsImgRecog": false,
        "MultiPage": true
    },
    "video/mp4": {
        "FriendlyName": "MP4",
        "FileExtension": [
            "MP4",
            "mp4",
            "MOV"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": true,
        "SupportsImgRecog": false
    },
    "video/quicktime": {
        "FriendlyName": "MP4",
        "FileExtension": [
            "MP4",
            "mp4"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": true,
        "SupportsImgRecog": false
    },
    "video/x-matroska": {
        "FriendlyName": "MKV",
        "FileExtension": [
            "MKV",
            "mkv"
        ],
        "IsDisplayable": true,
        "IsRaw": false,
        "IsVideo": true,
        "SupportsImgRecog": false
    }
}
`

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

func NewTypeService() MediaTypeService {
	var marshMap map[string]MediaType
	json.Unmarshal([]byte(MediaTypeJson), &marshMap)

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
