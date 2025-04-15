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

var mimeMap = map[string]MediaType{}
var extMap = map[string]MediaType{}

func init() {
	var marshMap map[string]MediaType
	json.Unmarshal([]byte(MediaTypeJson), &marshMap)

	for k, t := range marshMap {
		t.Mime = k
		mimeMap[k] = t
	}

	for _, mt := range mimeMap {
		for _, ext := range mt.Extensions {
			extMap[ext] = mt
		}
	}
}

type MediaType struct {
	Mime            string   `json:"mime"`
	Name            string   `json:"FriendlyName"`
	RawThumbExifKey string   `json:"RawThumbExifKey"`
	Extensions      []string `json:"FileExtension"`
	Displayable     bool     `json:"IsDisplayable"`
	Raw             bool     `json:"IsRaw"`
	IsVideo         bool     `json:"IsVideo"`
	ImgRecog        bool     `json:"SupportsImgRecog"`
	MultiPage       bool     `json:"MultiPage"`
} // @name MediaType

// ParseExtension Get a pointer to the weblens Media type of a file given the file extension
func ParseExtension(ext string) MediaType {
	if len(ext) == 0 {
		return mimeMap["generic"]
	}

	if ext[0] == '.' {
		ext = ext[1:]
	}
	mt, ok := extMap[ext]
	if !ok {
		return mimeMap["generic"]
	}
	return mt
}

func ParseMime(mime string) MediaType {
	return mimeMap[mime]
}

func GetMaps() (map[string]MediaType, map[string]MediaType) {
	return mimeMap, extMap
}

func Generic() MediaType {
	return mimeMap["generic"]
}

func Size() int {
	return len(mimeMap)
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
