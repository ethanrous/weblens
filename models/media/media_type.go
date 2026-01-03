package media

import "encoding/json"

// MediaTypeJSON contains the JSON definition of all supported media types.
const MediaTypeJSON = `{
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
        "FriendlyName": "QuickTime",
        "FileExtension": [
			"MOV",
			"mov"
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

var mimeMap = map[string]MType{}
var extMap = map[string]MType{}

func init() {
	var marshMap map[string]MType

	err := json.Unmarshal([]byte(MediaTypeJSON), &marshMap)
	if err != nil {
		panic(err)
	}

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

// MType represents the properties and capabilities of a specific media file type.
type MType struct {
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
func ParseExtension(ext string) MType {
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

// ParseMime returns the MediaType for a given MIME type string.
func ParseMime(mime string) MType {
	return mimeMap[mime]
}

// GetMaps returns both the MIME type map and the extension map for media types.
func GetMaps() (map[string]MType, map[string]MType) {
	return mimeMap, extMap
}

// Generic returns the generic MediaType used for unsupported file types.
func Generic() MType {
	return mimeMap["generic"]
}

// Size returns the total number of registered media types.
func Size() int {
	return len(mimeMap)
}

// IsMime checks if the MediaType has the specified MIME type.
func (mt MType) IsMime(mime string) bool {
	return mt.Mime == mime
}

// IsDisplayable checks if the MediaType can be displayed in a browser.
func (mt MType) IsDisplayable() bool {
	return mt.Displayable
}

// FriendlyName returns the human-readable name of the MediaType.
func (mt MType) FriendlyName() string {
	return mt.Name
}

// IsSupported checks if the MediaType is a recognized and supported type.
func (mt MType) IsSupported() bool {
	return mt.Mime != "generic" && mt.Mime != ""
}

// IsMultiPage checks if the MediaType supports multiple pages.
func (mt MType) IsMultiPage() bool {
	return mt.MultiPage
}

// GetThumbExifKey returns the EXIF key used to extract thumbnails from RAW images.
func (mt MType) GetThumbExifKey() string {
	return mt.RawThumbExifKey
}

// SupportsImgRecog checks if the MediaType supports image recognition processing.
func (mt MType) SupportsImgRecog() bool {
	return mt.ImgRecog
}

func rawMimes() []string {
	rawMimes := []string{}

	for mime, mediaType := range mimeMap {
		if mediaType.Raw {
			rawMimes = append(rawMimes, mime)
		}
	}

	return rawMimes
}
