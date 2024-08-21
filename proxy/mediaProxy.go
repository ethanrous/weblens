package proxy

import (
	"context"
	"fmt"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) GetAllMedia() ([]types.Media, error) {
	ret, err := p.CallHome("GET", "/api/core/media", nil)
	if err != nil {
		return nil, err
	}

	ms, err := ReadResponseBody[[]*weblens.Media](ret)
	if err != nil {
		return nil, err
	}

	return internal.SliceConvert[types.Media](ms), nil
}

func (p *ProxyStoreImpl) CreateMedia(m types.Media) error {
	return werror.NotImplemented("CreateMedia proxy")
}

func (p *ProxyStoreImpl) AddFileToMedia(mId weblens.ContentId, fId types.FileId) error {
	return werror.NotImplemented("AddFileToMedia proxy")
}

func (p *ProxyStoreImpl) RemoveFileFromMedia(mId weblens.ContentId, fId types.FileId) error {
	return werror.NotImplemented("RemoveFileFromMedia proxy")
}

func (p *ProxyStoreImpl) DeleteMedia(id weblens.ContentId) error {
	return werror.NotImplemented("DeleteMedia proxy")
}

func (p *ProxyStoreImpl) SetMediaHidden(id weblens.ContentId, hidden bool) error {
	return werror.NotImplemented("SetMediaHidden proxy")
}

func (p *ProxyStoreImpl) DeleteAllMedia() error {
	return werror.NotImplemented("DeleteAllMedia proxy")
}

func (p *ProxyStoreImpl) GetFetchMediaCacheImage(ctx context.Context) ([]byte, error) {
	wlog.Debug.Println("Cache miss")
	ret, err := p.CallHome(
		"GET", fmt.Sprintf(
			"/api/core/media/%s/content?quality=%s&page=%d",
			ctx.Value("media").(types.Media).ID(), ctx.Value("quality").(weblens.MediaQuality),
			ctx.Value("pageNum").(int),
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

func (p *ProxyStoreImpl) AddLikeToMedia(id weblens.ContentId, user types.Username, liked bool) error {
	return werror.NotImplemented("AddLikeToMedia proxy")
}
