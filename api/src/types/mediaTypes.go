package types

import (
	"time"

	"github.com/barasher/go-exiftool"
)

const (
	Thumbnail Quality = "thumbnail"
	Fullres   Quality = "fullres"
	Video     Quality = "video"
)

type MediaRepo interface {
	WeblensService[ContentId, Media, MediaStore]
	TypeService() MediaTypeService
	FetchCacheImg(m Media, q Quality, pageNum int) ([]byte, error)
	StreamCacheVideo(m Media, startByte, endByte int) ([]byte, error)
	GetFilteredMedia(
		requester User, sort string, sortDirection int, albumFilter []AlbumId, raw bool, hidden bool,
	) ([]Media, error)
	RunExif(path string) ([]exiftool.FileMetadata, error)
	GetAll() []Media
	NukeCache() error
	SetMediaLiked(mediaId ContentId, liked bool, username Username) error
}

type Media interface {
	ID() ContentId
	IsImported() bool
	IsCached() bool
	IsFilledOut() (bool, string)
	IsHidden() bool
	IsEnabled() bool
	GetOwner() User

	SetOwner(User)
	SetImported(bool)
	SetEnabled(bool)
	SetContentId(id ContentId)

	GetMediaType() MediaType
	GetCreateDate() time.Time
	SetCreateDate(time.Time) error

	GetProminentColors() (prom []string, err error)

	Hide(bool) error

	Clean()
	AddFile(WeblensFile) error
	RemoveFile(file WeblensFile) error
	GetFiles() []FileId

	LoadFromFile(WeblensFile, []byte, Task) (Media, error)

	ReadDisplayable(Quality, int) ([]byte, error)
	GetPageCount() int
	GetVideoLength() int

	GetCacheFile(q Quality, generateIfMissing bool, pageNum int) (WeblensFile, error)
	// MarshalJSON SetPageCount(int)
	MarshalJSON() ([]byte, error)
}

type ContentId string
type Quality string

type MediaType interface {
	IsRaw() bool
	IsVideo() bool
	IsDisplayable() bool
	FriendlyName() string
	GetMime() string
	IsMime(string) bool
	IsMultiPage() bool
	GetThumbExifKey() string
	SupportsImgRecog() bool
	IsSupported() bool
}

type MediaTypeService interface {
	ParseExtension(ext string) MediaType
	ParseMime(mime string) MediaType
	Generic() MediaType
	Size() int
}

// Error

var ErrNoMedia = NewWeblensError("no media found")
var ErrNoExiftool = NewWeblensError("exiftool not initialized")
