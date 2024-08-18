package types

type AlbumId string

type Album interface {
	ID() AlbumId
	GetName() string
	GetMedias() []Media
	GetCover() ContentId
	GetOwner() User
	GetSharedWith() []Username
	GetPrimaryColor() string

	AddMedia(...Media) error
	Rename(newName string) error
	AddUsers(...User) error

	RemoveMedia(...ContentId) error
	RemoveUsers(...Username) error
}

type AlbumService interface {
	WeblensService[AlbumId, Album, AlbumsStore]
	GetAllByUser(u User) []Album
	RemoveMediaFromAny(ContentId) error
	SetAlbumCover(albumId AlbumId, cover Media) error
}
