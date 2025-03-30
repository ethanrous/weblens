package router

import (
	"net/http"

	"github.com/ethanrous/weblens/services/context"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

type Router struct {
	chi.Router
	prefix      string
	middlewares []func(HandlerFunc) HandlerFunc
}

// @title						Weblens API
// @version					1.0
// @description				Programmatic access to the Weblens server
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
// @host						localhost:8080
// @schemes					http https
// @BasePath					/api/
//
// @securityDefinitions.apikey	SessionAuth
// @in							cookie
// @name						weblens-session-token
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
//
// @scope.admin				Grants read and write access to privileged data
func NewRouter() *Router {
	return &Router{Router: chi.NewRouter()}
}

type Injection struct {
	DB  *mongo.Database
	Log *zerolog.Logger
}

func (r *Router) Inject(i Injection) {
	r.Router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.RequestContext{
					Context: r.Context(),
					Req:     r,
					W:       w,
					DB:      i.DB,
					Log:     i.Log,
				}

				r = r.WithContext(ctx)

				next.ServeHTTP(w, r)
			})
		},
	)
}

func (r *Router) Get(path string, h HandlerFunc) {
	r.Router.Get(r.prefix+path, handlerWrapper(h))
}
func (r *Router) Post(path string, h HandlerFunc) {
	r.Router.Post(r.prefix+path, handlerWrapper(h))
}
func (r *Router) Put(path string, h HandlerFunc) {
	r.Router.Put(r.prefix+path, handlerWrapper(h))
}
func (r *Router) Patch(path string, h HandlerFunc) {
	r.Router.Patch(r.prefix+path, handlerWrapper(h))
}
func (r *Router) Head(path string, h HandlerFunc) {
	r.Router.Head(r.prefix+path, handlerWrapper(h))
}
func (r *Router) Delete(path string, h HandlerFunc) {
	r.Router.Delete(r.prefix+path, handlerWrapper(h))
}

func (r *Router) Route(pattern string, fn func(r *Router)) *Router             { return r }
func (r *Router) Group(fn func(r *Router), middlewares ...HandlerFunc) *Router { return r }

// Use supports two middlewares
func (r *Router) Use(middlewares ...HandlerFunc) {
	for _, m := range middlewares {
		if m != nil {
			r.Router.Use(middlewareWrapper(m))
		}
	}
}

func (r *Router) Start() {
	for {
		if s.services.StartupChan == nil {
			return
		}

		s.router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
		})

		// Kinda hacky, but allows for docs to be served from /docs/ instead of /docs/index.html
		s.router.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/docs/" {
				r.RequestURI = "/docs/index.html"
			}
			httpSwagger.WrapHandler(w, r)
		})

		s.router.Mount("/api", s.UseApi())

		if !env.DetachUi() {
			s.router.Mount("/", s.UseUi())
		}

		s.RouterLock.Lock()
		go s.StartupFunc()

		startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		select {
		case <-s.services.StartupChan:
			cancel()
		case <-startupCtx.Done():
			cancel()
			s.services.Log.WithLevel(zerolog.FatalLevel).Msg("Server startup timed out, exiting")
			os.Exit(1)
		}
		s.services.Log.Debug().Msg("Startup function signaled to continue")

		s.stdServer = &http.Server{Addr: s.hostStr, Handler: s.router, ReadHeaderTimeout: 5 * time.Second}
		s.Running = true

		s.services.Log.Debug().Msgf("Starting router at %s", s.hostStr)
		s.RouterLock.Unlock()

		err := s.stdServer.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			s.services.Log.Fatal().Err(err).Msg("Error starting server")
		}
		s.RouterLock.Lock()
		s.Running = false
		s.stdServer = nil

		// s.router = gin.New()
		s.router = chi.NewRouter()
		s.RouterLock.Unlock()
	}
}
