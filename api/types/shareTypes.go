package types

import "time"

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetContentId() string
	SetContentId(string)
	IsPublic() bool
	SetPublic(bool)
	IsEnabled() bool
	SetEnabled(bool)
	GetAccessors() []User
	SetAccessors([]Username)
	GetOwner() User
}

type AccessMeta interface {
	AddShare(Share) error
	AddShareId(ShareId, ShareType) AccessMeta
	SetRequestMode(RequestMode) AccessMeta
	SetTime(t time.Time) AccessMeta

	User() User
	Shares() []Share
	RequestMode() RequestMode
	GetTime() time.Time

	UsingShare() Share
	SetUsingShare(Share)
}

type RequestMode string
type ShareType string
type ShareId string

func (sId ShareId) String() string {
	return string(sId)
}
