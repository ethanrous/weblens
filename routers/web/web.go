package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/services/context"
)

func CacheMiddleware(ctx *context.RequestContext) {
	ctx.W.Header().Set("Cache-Control", "public, max-age=3600")
	ctx.W.Header().Set("Content-Encoding", "gzip")
}

func NewMemFs(ctx *context.AppContext, cnf config.ConfigProvider) *InMemoryFS {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}, proxyAddress: cnf.ProxyAddress, ctx: ctx}
	memFs.loadIndex(cnf.UIPath)

	return memFs
}

func UiRoutes(memFs *InMemoryFS) func() *router.Router {
	return func() *router.Router {
		r := router.NewRouter()

		r.Route("/assets", func(r *router.Router) {
			r.Use(CacheMiddleware)
			r.Handle("/*", http.FileServer(memFs))
		})

		r.NotFound(
			func(ctx *context.RequestContext) {
				if !strings.HasPrefix(ctx.Req.RequestURI, "/api") {
					ctx.Log().Info().Msg("Writing index")
					index := memFs.Index(ctx, ctx.Req.RequestURI)
					_, err := ctx.W.Write(index.realFile.data)
					if err != nil {
						ctx.Error(http.StatusInternalServerError, err)
						return
					}
				} else {
					ctx.Error(http.StatusNotFound, nil)
					return
				}
			},
		)
		return r
	}
}
