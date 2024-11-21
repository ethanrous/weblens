package proxy

import (
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
)

var _ models.MediaService = (*ProxyMediaService)(nil)

type ProxyMediaService struct {
	Core *models.Instance
}

func (pms *ProxyMediaService) Size() int {
	panic("implement me")
}

func (pms *ProxyMediaService) Add(media *models.Media) error {
	panic("implement me")
}

func (pms *ProxyMediaService) Get(id models.ContentId) *models.Media {
	panic("implement me")
}

func (pms *ProxyMediaService) GetAll() []*models.Media {
	panic("implement me")
}

func (pms *ProxyMediaService) Del(id models.ContentId) error {
	panic("implement me")
}

func (pms *ProxyMediaService) HideMedia(m *models.Media, hidden bool) error {
	panic("implement me")
}

func (pms *ProxyMediaService) LoadMediaFromFile(m *models.Media, file *fileTree.WeblensFileImpl) error {
	panic("implement me")
}

func (pms *ProxyMediaService) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {
	panic("implement me")
}

func (pms *ProxyMediaService) Cleanup() error {

	panic("implement me")
}

func (pms *ProxyMediaService) GetMediaType(m *models.Media) models.MediaType {
	panic("implement me")
}

func (pms *ProxyMediaService) GetProminentColors(media *models.Media) (prom []string, err error) {
	panic("implement me")
}

func (pms *ProxyMediaService) GetMediaTypes() models.MediaTypeService {
	panic("implement me")
}

func (pms *ProxyMediaService) IsFileDisplayable(file *fileTree.WeblensFileImpl) bool {
	panic("implement me")
}

func (pms *ProxyMediaService) IsCached(m *models.Media) bool {
	panic("implement me")
}

func (pms *ProxyMediaService) FetchCacheImg(m *models.Media, quality models.MediaQuality, pageNum int) ([]byte, error) {
	panic("implement me")
}

func (pms *ProxyMediaService) StreamVideo(m *models.Media, u *models.User, share *models.FileShare) (
	*models.VideoStreamer, error,
) {
	panic("implement me")
}

func (pms *ProxyMediaService) StreamCacheVideo(m *models.Media, startByte, endByte int) ([]byte, error) {
	panic("implement me")
}

func (pms *ProxyMediaService) NukeCache() error {
	panic("implement me")
}

func (pms *ProxyMediaService) GetFilteredMedia(
	requester *models.User, sort string, sortDirection int, excludeIds []models.ContentId, raw bool, hidden bool,
) ([]*models.Media, error) {
	panic("implement me")
}

func (pms *ProxyMediaService) RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []*models.Media {
	panic("implement me")
}

func (pms *ProxyMediaService) SetMediaLiked(mediaId models.ContentId, liked bool, username models.Username) error {
	panic("implement me")
}

func (pms *ProxyMediaService) AdjustMediaDates(
	anchor *models.Media, newTime time.Time, extraMedias []*models.Media,
) error {
	panic("implement me")
}

func (pms *ProxyMediaService) AddFileToMedia(m *models.Media, f *fileTree.WeblensFileImpl) error {
	panic("implement me")
}
