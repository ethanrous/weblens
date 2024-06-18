package media

import (
	"context"
	"fmt"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/ethrousseau/weblens/api/types"
)

var thumbnailCache = sturdyc.New[[]byte](500, 10, time.Hour, 10)

func getMediaCache(m types.Media, q types.Quality, pageNum int, ft types.FileTree) ([]byte, error) {
	cacheKey := string(m.ID()) + string(q)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "cacheKey", cacheKey)
	ctx = context.WithValue(ctx, "quality", q)
	ctx = context.WithValue(ctx, "pageNum", pageNum)
	ctx = context.WithValue(ctx, "Media", m)
	ctx = context.WithValue(ctx, "fileTree", ft)
	return thumbnailCache.GetFetch(ctx, cacheKey, memCacheMediaImage)
}

func memCacheMediaImage(ctx context.Context) (data []byte, err error) {
	m := ctx.Value("Media").(*Media)
	q := ctx.Value("quality").(types.Quality)
	ft := ctx.Value("fileTree").(types.FileTree)
	pageNum := ctx.Value("pageNum").(int)

	f, err := m.GetCacheFile(q, true, pageNum, ft)
	if err != nil {
		return
	} else if f == nil {
		return nil, types.ErrNoFile
	}

	data, err = f.ReadAll()
	if err != nil {
		return
	}
	if len(data) == 0 {
		err = fmt.Errorf("displayable bytes empty")
		return
	}

	return

}
