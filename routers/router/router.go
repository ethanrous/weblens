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
