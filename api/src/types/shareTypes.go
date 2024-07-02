package types

import "time"

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetItemId() string
	IsPublic() bool
	IsEnabled() bool
	GetAccessors() []User
	GetOwner() User
	AddUsers(newUsers []User) error

	SetItemId(string)
	SetPublic(bool)
	SetEnabled(bool) error
	// SetAccessors(newUsers []User, c ...BroadcasterAgent)
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

	GetSharedWithUser(u User) ([]Share, error)
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

var ErrUserNotAuthorized = NewWeblensError("user does not have access the requested resource")
