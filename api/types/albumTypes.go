package types

type AlbumId string

type Album interface {
	ID() AlbumId
	GetName() string
	GetMedias() []Media
	GetCover() Media
	GetOwner() User
	GetUsers() []User
	GetPrimaryColor() string

	AddMedia(...Media) error
	SetCover(Media) error
	Rename(newName string) error
	AddUsers(...User) error

	RemoveMedia(...ContentId) error
	RemoveUsers(...Username) error
}

type AlbumService interface {
	BaseService[AlbumId, Album]
	GetAllByUser(u User) []Album
	RemoveMediaFromAny(ContentId) error
}
