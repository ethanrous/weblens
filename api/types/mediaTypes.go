package types

import "time"

type MediaRepo interface {
	Get(ContentId) Media
	Add(m Media) error
	TypeService() MediaTypeService
	Del(Media, FileTree)
	Size() int
}

type Media interface {
	ID() ContentId
	IsImported() bool
	IsCached(FileTree) bool
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

	Clean()
	Save() error
	AddFile(WeblensFile)
	RemoveFile(file WeblensFile)
	GetFiles() []FileId

	LoadFromFile(WeblensFile, []byte, Task) (Media, error)

	ReadDisplayable(Quality, FileTree, ...int) ([]byte, error)
	GetPageCount() int
	// SetPageCount(int)
}

type ContentId string
type Quality string

type MediaType interface {
	IsRaw() bool
	IsDisplayable() bool
	FriendlyName() string
	GetMime() string
	IsMime(string) bool
	IsMultiPage() bool
	GetThumbExifKey() string
	SupportsImgRecog() bool
}

type MediaTypeService interface {
	ParseExtension(ext string) MediaType
	ParseMime(mime string) MediaType
	Generic() MediaType
	Size() int
}
