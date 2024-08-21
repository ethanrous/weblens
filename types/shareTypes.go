package types

import (
	"time"

	"github.com/ethrousseau/weblens/api"
)

type AccessMeta interface {
	AddShare(weblens.Share) error
	AddShareId(ShareId) AccessMeta
	SetRequestMode(RequestMode) AccessMeta
	SetTime(t time.Time) AccessMeta

	User() User
	Shares() []weblens.Share
	RequestMode() RequestMode
	GetTime() time.Time

	UsingShare() weblens.Share
	SetUsingShare(weblens.Share)

	CanAccessFile(WeblensFile) bool
	CanAccessShare(weblens.Share) bool
	CanAccessAlbum(Album) bool
}

type RequestMode string
type ShareId string

func (sId ShareId) String() string {
	return string(sId)
}

