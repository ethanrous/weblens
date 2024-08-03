package proxy

import (
	"context"
	"fmt"

	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
)

func (p *ProxyStore) GetAllMedia() ([]types.Media, error) {
	ret, err := p.CallHome("GET", "/api/core/media", nil)
	if err != nil {
		return nil, err
	}

	ms, err := ReadResponseBody[[]*media.Media](ret)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Media](ms), nil
}

func (p *ProxyStore) CreateMedia(m types.Media) error {
	return types.ErrNotImplemented("CreateMedia proxy")
}

func (p *ProxyStore) AddFileToMedia(mId types.ContentId, fId types.FileId) error {
	return types.ErrNotImplemented("AddFileToMedia proxy")
}

func (p *ProxyStore) RemoveFileFromMedia(mId types.ContentId, fId types.FileId) error {
	return types.ErrNotImplemented("RemoveFileFromMedia proxy")
}

func (p *ProxyStore) DeleteMedia(id types.ContentId) error {
	return types.ErrNotImplemented("DeleteMedia proxy")
}

func (p *ProxyStore) SetMediaHidden(id types.ContentId, hidden bool) error {
	return types.ErrNotImplemented("SetMediaHidden proxy")
}

func (p *ProxyStore) DeleteAllMedia() error {
	return types.ErrNotImplemented("DeleteAllMedia proxy")
}

func (p *ProxyStore) GetFetchMediaCacheImage(ctx context.Context) ([]byte, error) {
	wlog.Debug.Println("Cache miss")
	ret, err := p.CallHome(
		"GET", fmt.Sprintf(
			"/api/core/media/%s/content?quality=%s&page=%d",
			ctx.Value("media").(types.Media).ID(), ctx.Value("quality").(types.Quality), ctx.Value("pageNum").(int),
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	bs, err := ReadResponseBody[[]byte](ret)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
