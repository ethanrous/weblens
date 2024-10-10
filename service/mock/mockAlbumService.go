package mock

import (
	"iter"

	"github.com/ethanrous/weblens/models"
)

var _ models.AlbumService = (*MockAlbumService)(nil)

type MockAlbumService struct{}

func (m *MockAlbumService) Init() error {
	
	return nil
}

func (m *MockAlbumService) Size() int {
	
	return 0
}

func (m *MockAlbumService) Get(id models.AlbumId) *models.Album {
	
	return nil
}

func (m *MockAlbumService) Add(album *models.Album) error {
	
	return nil
}

func (m *MockAlbumService) Del(id models.AlbumId) error {
	
	return nil
}

func (m *MockAlbumService) GetAllByUser(u *models.User) ([]*models.Album, error) {
	
	return nil, nil
}

func (m *MockAlbumService) GetAlbumMedias(album *models.Album) iter.Seq[*models.Media] {
	
	return nil
}

func (m *MockAlbumService) RenameAlbum(album *models.Album, newName string) error {
	
	return nil
}

func (m *MockAlbumService) SetAlbumCover(albumId models.AlbumId, cover *models.Media) error {
	
	return nil
}

func (m *MockAlbumService) AddMediaToAlbum(album *models.Album, media ...*models.Media) error {
	
	return nil
}

func (m *MockAlbumService) RemoveMediaFromAlbum(album *models.Album, mediaIds ...models.ContentId) error {
	
	return nil
}

func (m *MockAlbumService) RemoveMediaFromAny(id models.ContentId) error {
	
	return nil
}
