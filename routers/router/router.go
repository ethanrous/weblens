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
	middlewares []func(http.Handler) http.Handler
}

// @title						Weblens API
// @version					1.0
// @description				Programmatic access to the Weblens server
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
// @host						localhost:8080
// @schemes					http https
// @BasePath					/api/v1/
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

				reqContext.Req = r

				next.ServeHTTP(w, r)
			})
		},
	)
}

func (r *Router) Get(path string, h ...HandlerFunc) {
	r.chi.With(r.middlewares...).With(wrapManyHandlers(h[:len(h)-1]...)...).Get(r.prefix+path, toStdHandlerFunc(h[len(h)-1]))
}
func (r *Router) Post(path string, h ...HandlerFunc) {
	r.chi.With(r.middlewares...).With(wrapManyHandlers(h[:len(h)-1]...)...).Post(r.prefix+path, toStdHandlerFunc(h[len(h)-1]))
}
func (r *Router) Put(path string, h HandlerFunc) {
	r.chi.Put(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Patch(path string, h ...HandlerFunc) {
	r.chi.With(r.middlewares...).With(wrapManyHandlers(h[:len(h)-1]...)...).Patch(r.prefix+path, toStdHandlerFunc(h[len(h)-1]))
}
func (r *Router) Head(path string, h HandlerFunc) {
	r.chi.Head(r.prefix+path, toStdHandlerFunc(h))
}
func (r *Router) Delete(path string, h HandlerFunc) {
	r.chi.Delete(r.prefix+path, toStdHandlerFunc(h))
}

func (r *Router) Route(pattern string, fn func(r *Router)) *Router {
	r.chi.Route(r.prefix+pattern, func(r chi.Router) {
		fn(&Router{chi: r})
	})
	return r
}

func (r *Router) Group(path string, fn func(), middlewares ...HandlerFunc) {
	previousGroupPrefix := r.prefix
	previousMiddlewares := r.middlewares

	for _, m := range middlewares {
		r.middlewares = append(r.middlewares, middlewareWrapper(m))
	}
	r.prefix += path

	fn()

	r.prefix = previousGroupPrefix
	r.middlewares = previousMiddlewares
}

func (r *Router) Handle(pattern string, h http.Handler) { r.chi.Handle(pattern, h) }

func (r *Router) NotFound(h HandlerFunc) {
	r.chi.NotFound(toStdHandlerFunc(h))
}

func (r *Router) Use(middlewares ...PassthroughHandler) {
	for _, m := range middlewares {
		if m != nil {
			r.chi.Use(mdlwToStd(m))
		}
	}
}
