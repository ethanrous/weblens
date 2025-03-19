package mock

import (
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
)

var _ models.MediaService = (*MockMediaService)(nil)

type MockMediaService struct{}

func (ms *MockMediaService) Size() int {

	panic("implement me")
}

func (ms *MockMediaService) Add(media *models.Media) error {

	panic("implement me")
}

func (ms *MockMediaService) Get(id models.ContentId) *models.Media {
	return nil
}

func (ms *MockMediaService) GetAll() []*models.Media {

	panic("implement me")
}

func (ms *MockMediaService) Del(id models.ContentId) error {

	panic("implement me")
}

func (ms *MockMediaService) HideMedia(m *models.Media, hidden bool) error {

	panic("implement me")
}

func (ms *MockMediaService) LoadMediaFromFile(m *models.Media, file *fileTree.WeblensFileImpl) error {

	panic("implement me")
}

func (ms *MockMediaService) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {

	panic("implement me")
}

func (ms *MockMediaService) Cleanup() error {

	panic("implement me")
}

func (ms *MockMediaService) Drop() error {

	panic("implement me")
}

func (ms *MockMediaService) GetMediaType(m *models.Media) models.MediaType {

	panic("implement me")
}

func (ms *MockMediaService) GetProminentColors(media *models.Media) (prom []string, err error) {

	panic("implement me")
}

func (ms *MockMediaService) GetMediaTypes() models.MediaTypeService {

	panic("implement me")
}

func (ms *MockMediaService) IsFileDisplayable(file *fileTree.WeblensFileImpl) bool {

	panic("implement me")
}

func (ms *MockMediaService) IsCached(m *models.Media) bool {

	panic("implement me")
}

func (ms *MockMediaService) FetchCacheImg(m *models.Media, quality models.MediaQuality, pageNum int) ([]byte, error) {

	panic("implement me")
}

func (ms *MockMediaService) StreamVideo(m *models.Media, u *models.User, share *models.FileShare) (
	*models.VideoStreamer, error,
) {

	panic("implement me")
}

func (ms *MockMediaService) StreamCacheVideo(m *models.Media, startByte, endByte int) ([]byte, error) {

	panic("implement me")
}

func (ms *MockMediaService) NukeCache() error {

	panic("implement me")
}

func (ms *MockMediaService) GetFilteredMedia(
	requester *models.User, sort string, sortDirection int, excludeIds []models.ContentId, raw bool, hidden bool, search string,
) ([]*models.Media, error) {

	panic("implement me")
}

func (ms *MockMediaService) RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []*models.Media {

	panic("implement me")
}

func (ms *MockMediaService) SetMediaLiked(mediaId models.ContentId, liked bool, username models.Username) error {

	panic("implement me")
}

func (ms *MockMediaService) AdjustMediaDates(
	anchor *models.Media, newTime time.Time, extraMedias []*models.Media,
) error {

	panic("implement me")
}

func (mms *MockMediaService) AddFileToMedia(m *models.Media, f *fileTree.WeblensFileImpl) error {
	panic("implement me")
}

func (mms *MockMediaService) GetMediaConverted(m *models.Media, format string) ([]byte, error) {
	panic("implement me")
}
