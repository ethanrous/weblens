package router

import (
	"context"
	"net/http"

	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/go-chi/chi/v5"
)

var _ http.Handler = &Router{}

type Router struct {
	chi         chi.Router
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
	return &Router{chi: chi.NewRouter()}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

func (r *Router) Mount(prefix string, fn func() *Router) {
	r.prefix = prefix
	r.chi.Mount(prefix, fn())
}

const requestContextKey = "requestContext"

func (r *Router) WithAppContext(ctx context_service.AppContext) {
	r.chi.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqContext := &context_service.RequestContext{
					AppContext: ctx,
					Req:        r,
					W:          w,
				}

				r = r.WithContext(context.WithValue(r.Context(), requestContextKey, reqContext))

				next.ServeHTTP(w, r)
			})
		},
	)
}

func (r *Router) Get(path string, h HandlerFunc) {
	r.chi.Get(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Post(path string, h HandlerFunc) {
	r.chi.Post(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Put(path string, h HandlerFunc) {
	r.chi.Put(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Patch(path string, h HandlerFunc) {
	r.chi.Patch(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Head(path string, h HandlerFunc) {
	r.chi.Head(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Delete(path string, h HandlerFunc) {
	r.chi.Delete(r.prefix+path, toStdHandlerFunc(h))
}

func (r *Router) Route(pattern string, fn func(r *Router)) *Router             { return r }
func (r *Router) Group(fn func(r *Router), middlewares ...HandlerFunc) *Router { return r }
func (r *Router) Handle(pattern string, h http.Handler)                        { r.chi.Handle(pattern, h) }

func (r *Router) NotFound(h HandlerFunc) {
	r.chi.NotFound(toStdHandlerFunc(h))
}

// Use supports two middlewares
func (r *Router) Use(middlewares ...HandlerFunc) {
	for _, m := range middlewares {
		if m != nil {
			r.chi.Use(middlewareWrapper(m))
		}
	}
}

// func (r *Router) Start() {
// 	for {
//
// 		if !env.DetachUi() {
// 			s.router.Mount("/", s.UseUi())
// 		}
//
// 		s.RouterLock.Lock()
// 		go s.StartupFunc()
//
// 		startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//
// 		select {
// 		case <-s.services.StartupChan:
// 			cancel()
// 		case <-startupCtx.Done():
// 			cancel()
// 			s.services.Log.WithLevel(zerolog.FatalLevel).Msg("Server startup timed out, exiting")
// 			os.Exit(1)
// 		}
// 		s.services.Log.Debug().Msg("Startup function signaled to continue")
//
// 		s.stdServer = &http.Server{Addr: s.hostStr, Handler: s.router, ReadHeaderTimeout: 5 * time.Second}
// 		s.Running = true
//
// 		s.services.Log.Debug().Msgf("Starting router at %s", s.hostStr)
// 		s.RouterLock.Unlock()
//
// 		err := s.stdServer.ListenAndServe()
//
// 		if !errors.Is(err, http.ErrServerClosed) {
// 			s.services.Log.Fatal().Err(err).Msg("Error starting server")
// 		}
// 		s.RouterLock.Lock()
// 		s.Running = false
// 		s.stdServer = nil
//
// 		// s.router = gin.New()
// 		s.router = chi.NewRouter()
// 		s.RouterLock.Unlock()
// 	}
// }
