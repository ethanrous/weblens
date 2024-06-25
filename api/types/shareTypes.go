package types

import "time"

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetItemId() string
	SetItemId(string)
	IsPublic() bool
	SetPublic(bool)
	IsEnabled() bool
	SetEnabled(bool) error
	GetAccessors() []User
	SetAccessors(newUsers []User, c ...BroadcasterAgent)
	GetOwner() User
}

type AccessMeta interface {
	AddShare(Share) error
	AddShareId(ShareId) AccessMeta
	SetRequestMode(RequestMode) AccessMeta
	SetTime(t time.Time) AccessMeta

	User() User
	Shares() []Share
	RequestMode() RequestMode
	GetTime() time.Time

	UsingShare() Share
	SetUsingShare(Share)

	CanAccessFile(WeblensFile) bool
	CanAccessShare(Share) bool
	CanAccessAlbum(Album) bool
}

type ShareService interface {
	BaseService[ShareId, Share]
}

type RequestMode string
type ShareType string
type ShareId string

func (sId ShareId) String() string {
	return string(sId)
}

const (
	FileShare  ShareType = "file"
	AlbumShare ShareType = "album"
)
