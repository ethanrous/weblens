package proxy

import (
	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) GetAllAlbums() ([]types.Album, error) {
	wlog.Debug.Println("implement me")
	return []types.Album{}, nil
}

func (p *ProxyStoreImpl) CreateAlbum(album types.Album) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) SetAlbumCover(id types.AlbumId, s string, s2 string, id2 weblens.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) GetAlbumsByMedia(id weblens.ContentId) ([]types.Album, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) AddMediaToAlbum(aId types.AlbumId, mIds []weblens.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) RemoveMediaFromAlbum(id types.AlbumId, id2 weblens.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) AddUsersToAlbum(aId types.AlbumId, us []types.User) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) DeleteAlbum(aId types.AlbumId) error { panic("implement me") }
