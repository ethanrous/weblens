package mock

import (
	"iter"

	"github.com/ethrousseau/weblens/models"
)

var _ models.AlbumService = (*MockAlbumService)(nil)

type MockAlbumService struct{}

func (m *MockAlbumService) Init() error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) Size() int {
	// TODO implement me
	return 0
}

func (m *MockAlbumService) Get(id models.AlbumId) *models.Album {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) Add(album *models.Album) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) Del(id models.AlbumId) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) GetAllByUser(u *models.User) ([]*models.Album, error) {
	// TODO implement me
	return nil, nil
}

func (m *MockAlbumService) GetAlbumMedias(album *models.Album) iter.Seq[*models.Media] {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) RenameAlbum(album *models.Album, newName string) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) SetAlbumCover(albumId models.AlbumId, cover *models.Media) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) AddMediaToAlbum(album *models.Album, media ...*models.Media) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) RemoveMediaFromAlbum(album *models.Album, mediaIds ...models.ContentId) error {
	// TODO implement me
	return nil
}

func (m *MockAlbumService) RemoveMediaFromAny(id models.ContentId) error {
	// TODO implement me
	return nil
}
