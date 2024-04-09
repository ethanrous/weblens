package types

import "time"

type Media interface {
	Id() MediaId
	IsImported() bool
	IsFilledOut() (bool, string)
	SetImported(bool)
	GetMediaType() MediaType
	GetCreateDate() time.Time
	Clean()
	LoadFromFile(WeblensFile, Task) (Media, error)
	Save() error
	AddFile(WeblensFile)
	RemoveFile(FileId)
	ReadDisplayable(Quality, ...int) ([]byte, error)
	PageCount() int
}

type MediaId string
type Quality string

type MediaType interface {
	IsRaw() bool
	IsDisplayable() bool
	FriendlyName() string
}
