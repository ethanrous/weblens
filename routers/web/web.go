package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog/log"
)

func CacheMiddleware(next router.Handler) router.Handler {
	return router.HandlerFunc(func(ctx context.RequestContext) {
		ctx.W.Header().Set("Cache-Control", "public, max-age=3600")
		ctx.W.Header().Set("Content-Encoding", "gzip")

		log.Debug().Msg("CacheMiddleware: Setting Cache-Control header")

		next.ServeHTTP(ctx)
	})
}

func NewMemFs(ctx context.AppContext, cnf config.ConfigProvider) *InMemoryFS {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}, proxyAddress: cnf.ProxyAddress, ctx: ctx}
	memFs.loadIndex(cnf.UIPath)

	return memFs
}

func UiRoutes(memFs *InMemoryFS) func() *router.Router {
	return func() *router.Router {
		r := router.NewRouter()

		r.Use(CacheMiddleware)

		r.Handle("/assets/*", http.FileServer(memFs))
		r.Get("/static/{filename}", serveStaticContent)

		r.NotFound(
			func(ctx context.RequestContext) {
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

var staticDir = ""

func serveStaticContent(ctx context.RequestContext) {
	filename := ctx.Path("filename")

	cnf := config.GetConfig()

	if staticDir == "" {
		testDir := filepath.Join(cnf.StaticContentPath, "/static")
		_, err := os.Stat(testDir)

		if err != nil {
			panic(err)
		}

		staticDir = testDir
	}

	fullPath := filepath.Join(staticDir, filename)

	f, err := os.Open(fullPath)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}
	defer f.Close()

	_, err = io.Copy(ctx.W, f)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}
