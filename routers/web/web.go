package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers/router"
	context_service "github.com/ethanrous/weblens/services/context"
)

func CacheMiddleware(addGzip bool) router.PassthroughHandler {
	return func(next router.Handler) router.Handler {
		return router.HandlerFunc(func(ctx context_service.RequestContext) {
			ctx.W.Header().Set("Cache-Control", "public, max-age=3600")

			if addGzip {
				ctx.W.Header().Set("Content-Encoding", "gzip")
			}

			next.ServeHTTP(ctx)
		})
	}
}

func NewMemFs(ctx context_service.AppContext, cnf config.ConfigProvider) *InMemoryFS {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal), routesMu: &sync.RWMutex{}, proxyAddress: cnf.ProxyAddress, ctx: ctx}
	memFs.loadIndex(cnf.UIPath)

	return memFs
}

func UiRoutes(memFs *InMemoryFS) *router.Router {
	r := router.NewRouter()

	r.Handle("/assets/*", CacheMiddleware(true), http.FileServer(memFs))
	r.Get("/static/{filename}", CacheMiddleware(false), serveStaticContent)
	r.Get("/docs", func(ctx context_service.RequestContext) {
		http.Redirect(ctx.W, ctx.Req, "/docs/", http.StatusMovedPermanently)
	})

	r.NotFound(
		func(ctx context_service.RequestContext) {
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

var staticDir = ""

func serveStaticContent(ctx context_service.RequestContext) {
	filename := ctx.Path("filename")

	cnf := config.GetConfig()

	if staticDir == "" {
		testDir := cnf.StaticContentPath
		_, err := os.Stat(testDir)

		if err != nil {
			panic(err)
		}

		staticDir = testDir
	}

	fullPath := filepath.Join(staticDir, filename)
	ctx.Log().Debug().Msgf("Serving static content: %s", fullPath)

	http.ServeFile(ctx.W, ctx.Req, fullPath)

	// f, err := os.Open(fullPath)
	// if err != nil {
	// 	ctx.Error(http.StatusNotFound, err)
	//
	// 	return
	// }
	// defer f.Close()
	//
	// _, err = io.Copy(ctx.W, f)
	// if err != nil {
	// 	ctx.Error(http.StatusInternalServerError, err)
	//
	// 	return
	// }
	//
	// ctx.Status(http.StatusOK)
}
