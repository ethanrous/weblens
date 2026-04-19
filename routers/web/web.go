package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers/router"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// cacheMiddleware creates a middleware that sets cache headers and optionally enables gzip encoding.
func cacheMiddleware(addGzip bool, cacheTime int) router.PassthroughHandler {
	cacheTimeStr := fmt.Sprintf("public, max-age=%d", cacheTime)

	return func(next router.Handler) router.Handler {
		return router.HandlerFunc(func(ctx context_service.RequestContext) {
			ctx.W.Header().Set("Cache-Control", cacheTimeStr)

			if addGzip && (strings.HasSuffix(ctx.Req.URL.Path, ".js") || strings.HasSuffix(ctx.Req.URL.Path, ".css")) {
				ctx.W.Header().Set("Content-Encoding", "gzip")
			}

			next.ServeHTTP(ctx)
		})
	}
}

// NewMemFs creates a new in-memory filesystem for serving UI assets.
func NewMemFs(ctx context_service.AppContext, cnf config.Provider) *InMemoryFS {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal), routesMu: &sync.RWMutex{}, proxyAddress: cnf.ProxyAddress, ctx: ctx}
	memFs.loadIndex(cnf.UIPath)

	return memFs
}

// UIRoutes configures and returns the router for serving the web UI.
func UIRoutes(memFs *InMemoryFS) *router.Router {
	r := router.NewRouter()

	r.Handle("/_nuxt/*", cacheMiddleware(true, int((time.Hour*24).Seconds())), http.FileServer(memFs))
	r.Get("/static/{filename}", cacheMiddleware(false, int(time.Hour.Seconds())), serveStaticContentFromCtx)
	r.Get("/docs", func(ctx context_service.RequestContext) {
		http.Redirect(ctx.W, ctx.Req, "/docs/", http.StatusMovedPermanently)
	})

	r.Get("/favicon.ico", func(ctx context_service.RequestContext) {
		serveStaticContent(ctx, "favicon.ico")
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

func serveStaticContentFromCtx(ctx context_service.RequestContext) {
	filename := ctx.Path("filename")

	serveStaticContent(ctx, filename)
}

func serveStaticContent(ctx context_service.RequestContext, filename string) {
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
}
