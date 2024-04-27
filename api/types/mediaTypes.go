package types

import "time"

type Media interface {
	Id() ContentId
	IsImported() bool
	IsFilledOut() (bool, string)
	IsHidden() bool
	GetOwner() User

	SetOwner(User)
	SetImported(bool)

	GetMediaType() MediaType
	GetCreateDate() time.Time

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
}
