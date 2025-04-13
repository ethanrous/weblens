package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/services/context"
)

func CacheMiddleware(next router.Handler) router.Handler {
	return router.HandlerFunc(func(ctx *context.RequestContext) {
		ctx.W.Header().Set("Cache-Control", "public, max-age=3600")
		ctx.W.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(ctx)
	})
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
					index := memFs.Index(ctx)
					_, err := ctx.W.Write(index.realFile.data)
					if err != nil {
						ctx.Status(http.StatusInternalServerError)
						return
					}
				} else {
					ctx.Status(http.StatusNotFound)
					return
				}
			},
		)
		return r
	}
}
