package router

import (
	"net/http"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/services/context"
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

func (r *Router) Mount(prefix string, h ...any) {
	subRouter := h[len(h)-1].(http.Handler)

	r.chi.With(r.middlewares...).With(parseMiddlewares(h[:len(h)-1]...)...).Mount(r.prefix+prefix, subRouter)
}

const requestContextKey = "requestContext"

func (r *Router) Method(method, path string, h ...any) {
	r.chi.With(r.middlewares...).With(parseMiddlewares(h[:len(h)-1]...)...).Method(method, r.prefix+path, parseHandlerFunc(h[len(h)-1]))
}

func (r *Router) Get(path string, h ...any) {
	r.Method(http.MethodGet, path, h...)
}
func (r *Router) Post(path string, h ...any) {
	r.Method(http.MethodPost, path, h...)
}
func (r *Router) Put(path string, h ...any) {
	r.Method(http.MethodPut, path, h...)
}
func (r *Router) Patch(path string, h ...any) {
	r.Method(http.MethodPatch, path, h...)
}
func (r *Router) Head(path string, h ...any) {
	r.Method(http.MethodHead, path, h...)
}
func (r *Router) Delete(path string, h ...any) {
	r.Method(http.MethodDelete, path, h...)
}

func (r *Router) Route(pattern string, fn func(r *Router)) *Router {
	r.chi.Route(r.prefix+pattern, func(r chi.Router) {
		fn(&Router{chi: r})
	})

	return r
}

func (r *Router) Group(path string, fn func(), middlewares ...any) {
	previousGroupPrefix := r.prefix
	previousMiddlewares := r.middlewares

	r.middlewares = append(r.middlewares, parseMiddlewares(middlewares...)...)
	r.prefix += path

	fn()

	r.prefix = previousGroupPrefix
	r.middlewares = previousMiddlewares
}

func (r *Router) Handle(prefix string, h ...any) {
	finalHandler := h[len(h)-1].(http.Handler)

	r.chi.With(r.middlewares...).With(parseMiddlewares(h[:len(h)-1]...)...).Handle(r.prefix+prefix, finalHandler)
}

func (r *Router) NotFound(h HandlerFunc) {
	r.chi.NotFound(toStdHandlerFunc(h))
}

func (r *Router) Use(middlewares ...any) {
	r.middlewares = append(r.middlewares, parseMiddlewares(middlewares)...)

	// for _, m := range middlewares {
	// 	if m != nil {
	// 		switch m := m.(type) {
	// 		case func(http.Handler) http.Handler:
	// 			r.middlewares = append(r.middlewares, m)
	// 		case PassthroughHandler:
	// 			r.middlewares = append(r.middlewares, mdlwToStd(m))
	// 		case func(Handler) Handler:
	// 			r.middlewares = append(r.middlewares, mdlwToStd(m))
	// 		case []func(http.Handler) http.Handler:
	// 			for _, mw := range m {
	// 				r.middlewares = append(r.middlewares, mw)
	// 			}
	// 		default:
	// 			panic(errors.Errorf("middleware must be a function or a PassthroughHandler but got %T", m))
	// 		}
	// 	}
	// }
	// log.Debug().Msgf("Registering %d middlewares", len(r.middlewares))
}

func parseHandlerFunc(h any) http.HandlerFunc {
	switch h := h.(type) {
	case http.HandlerFunc:
		return h
	case func(context.RequestContext):
		return toStdHandlerFunc(HandlerFunc(h))
	default:
		panic(errors.Errorf("handler is not a valid function: %T", h))
	}
}

func parseMiddlewares(middlewares ...any) []func(http.Handler) http.Handler {
	parsedMiddlewares := make([]func(http.Handler) http.Handler, 0, len(middlewares))
	for _, mw := range middlewares {
		switch mw := mw.(type) {
		case func(http.Handler) http.Handler:
			parsedMiddlewares = append(parsedMiddlewares, mw)
		case PassthroughHandler:
			parsedMiddlewares = append(parsedMiddlewares, mdlwToStd(mw))
		case func(Handler) Handler:
			parsedMiddlewares = append(parsedMiddlewares, mdlwToStd(mw))
		case []func(http.Handler) http.Handler:
			for _, m := range mw {
				parsedMiddlewares = append(parsedMiddlewares, m)
			}
		case []any:
			parsedMiddlewares = append(parsedMiddlewares, parseMiddlewares(mw...)...)
		case func(context.RequestContext):
			parsedMiddlewares = append(parsedMiddlewares, middlewareWrapper(HandlerFunc(mw)))
		default:
			panic(errors.Errorf("middleware is not a valid function: %T", mw))
		}
	}
	return parsedMiddlewares
}
