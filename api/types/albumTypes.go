package types

type AlbumId string

type Album interface {
	ID() AlbumId
}

type AlbumService interface {
	Add(a Album) error
}
