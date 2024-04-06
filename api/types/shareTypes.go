package types

type Share interface {
	GetShareId() ShareId
	GetShareType() ShareType
	GetContentId() string
	SetContentId(string)
	IsPublic() bool
	SetPublic(bool)
	IsEnabled() bool
	SetEnabled(bool)
	GetAccessors() []Username
	SetAccessors([]Username)
	GetOwner() Username
}

type AccessMeta interface {
	AddShare(Share) AccessMeta
	AddShareId(ShareId, ShareType) AccessMeta
	SetRequestMode(RequestMode) AccessMeta

	User() User
	Shares() []Share
	RequestMode() RequestMode

	UsingShare() Share
}

type RequestMode string
type ShareType string
type ShareId string

func (sId ShareId) String() string {
	return string(sId)
}
