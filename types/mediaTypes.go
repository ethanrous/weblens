package types

import (
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal/werror"
)

type Media interface {
	ID() weblens.ContentId
	IsImported() bool
	IsCached() bool
	IsFilledOut() (bool, string)
	IsHidden() bool
	IsEnabled() bool
	GetOwner() User

	SetOwner(User)
	SetImported(bool)
	SetEnabled(bool)
	SetContentId(id weblens.ContentId)

	GetMediaType() MediaType
	GetCreateDate() time.Time
	SetCreateDate(time.Time) werror.WErr

	GetProminentColors() (prom []string, err werror.WErr)

	Hide(bool) werror.WErr

	Clean()
	AddFile(WeblensFile) werror.WErr
	RemoveFile(file WeblensFile) werror.WErr
	GetFiles() []FileId

	LoadFromFile(WeblensFile, []byte, Task) (Media, werror.WErr)

	ReadDisplayable(weblens.MediaQuality, int) ([]byte, werror.WErr)
	GetPageCount() int
	GetVideoLength() int

	GetCacheFile(q weblens.MediaQuality, generateIfMissing bool, pageNum int) (WeblensFile, werror.WErr)
	// MarshalJSON SetPageCount(int)
	MarshalJSON() ([]byte, werror.WErr)
}

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

