package proxy

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
)

var _ models.MediaService = (*ProxyMediaService)(nil)

type ProxyMediaService struct {
	Core *models.Instance
}

func (pms *ProxyMediaService) Size() int {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) Add(media *models.Media) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) Get(id models.ContentId) *models.Media {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) GetAll() []*models.Media {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) Del(id models.ContentId) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) HideMedia(m *models.Media, hidden bool) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) LoadMediaFromFile(m *models.Media, file *fileTree.WeblensFileImpl) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) GetMediaType(m *models.Media) models.MediaType {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) GetProminentColors(media *models.Media) (prom []string, err error) {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) GetMediaTypes() models.MediaTypeService {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) IsFileDisplayable(file *fileTree.WeblensFileImpl) bool {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) IsCached(m *models.Media) bool {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) FetchCacheImg(m *models.Media, quality models.MediaQuality, pageNum int) ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) StreamVideo(m *models.Media, u *models.User, share *models.FileShare) (
	*models.VideoStreamer, error,
) {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) StreamCacheVideo(m *models.Media, startByte, endByte int) ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) NukeCache() error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) GetFilteredMedia(
	requester *models.User, sort string, sortDirection int, excludeIds []models.ContentId, raw bool, hidden bool,
) ([]*models.Media, error) {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []models.ContentId {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) SetMediaLiked(mediaId models.ContentId, liked bool, username models.Username) error {
	// TODO implement me
	panic("implement me")
}

func (pms *ProxyMediaService) AdjustMediaDates(
	anchor *models.Media, newTime time.Time, extraMedias []*models.Media,
) error {
	// TODO implement me
	panic("implement me")
}