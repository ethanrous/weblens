package types

import "time"

type Media interface {
	Id() ContentId
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

	Clean()
	Save() error
	AddFile(WeblensFile)
	RemoveFile(FileId)
	GetFiles() []FileId

	LoadFromFile(WeblensFile, []byte, Task) (Media, error)

	ReadDisplayable(Quality, ...int) ([]byte, error)
	GetPageCount() int
}

type ContentId string
type Quality string

type MediaType interface {
	IsRaw() bool
	IsDisplayable() bool
	FriendlyName() string
	IsMime(string) bool
}
