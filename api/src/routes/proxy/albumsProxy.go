package proxy

import "github.com/ethrousseau/weblens/api/types"

func (p *ProxyStore) GetAllAlbums() ([]types.Album, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) CreateAlbum(album types.Album) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) SetAlbumCover(id types.AlbumId, s string, s2 string, id2 types.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) GetAlbumsByMedia(id types.ContentId) ([]types.Album, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) AddMediaToAlbum(aId types.AlbumId, mIds []types.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) RemoveMediaFromAlbum(id types.AlbumId, id2 types.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) AddUsersToAlbum(aId types.AlbumId, us []types.User) error {
	// TODO implement me
	panic("implement me")
}
