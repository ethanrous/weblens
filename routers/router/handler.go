// Package router provides HTTP routing and handler functionality for the Weblens application.
package router

import (
	"net/http"

	context_service "github.com/ethanrous/weblens/services/context"
)

// HandlerFunc defines a function type that handles requests using the Weblens request context.
type HandlerFunc func(ctx context_service.RequestContext)

// PassthroughHandler defines a function type that wraps a handler and returns a new handler.
type PassthroughHandler func(Handler) Handler

// Handler defines the interface for types that can handle Weblens requests.
type Handler interface {
	ServeHTTP(ctx context_service.RequestContext)
}

// ServeHTTP implements the Handler interface by invoking the function.
func (f HandlerFunc) ServeHTTP(ctx context_service.RequestContext) {
	f(ctx)
}

func getFromHTTP(w http.ResponseWriter, r *http.Request) context_service.RequestContext {
	reqCtx, ok := context_service.ReqFromContext(r.Context())
	if !ok {
		panic("request context not found in request")
	}

	reqCtx = reqCtx.WithContext(r.Context()).(context_service.RequestContext)
	r.WithContext(reqCtx)
	reqCtx.Req = r
	reqCtx.W = w

	return reqCtx
}

func toStdHandlerFunc(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := getFromHTTP(w, r)
		h.ServeHTTP(ctx)
	}
}

// FromStdHandler converts a standard http.Handler to a Weblens HandlerFunc.
func FromStdHandler(h http.Handler) HandlerFunc {
	return func(ctx context_service.RequestContext) {
		r := ctx.Req
		r = r.WithContext(ctx)
		// ctx.Req = ctx.Req.WithContext(context.WithValue(ctx.Req.Context(), requestContextKey, ctx))
		h.ServeHTTP(ctx.W, r)
	}
}

func mdlwToStd(h PassthroughHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := getFromHTTP(w, r)
			h(FromStdHandler(next)).ServeHTTP(ctx)
		})
	}
}

// WrapHandlerProvider converts a standard middleware function to a PassthroughHandler.
func WrapHandlerProvider(hp func(http.Handler) http.Handler) PassthroughHandler {
	return func(next Handler) Handler {
		h := hp(toStdHandlerFunc(next))

		return FromStdHandler(h)
	}
}

func middlewareWrapper(h Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := getFromHTTP(w, r)

			h.ServeHTTP(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func wrapManyHandlers(hs ...HandlerFunc) []func(http.Handler) http.Handler {
	newHs := make([]func(http.Handler) http.Handler, len(hs))
	for i := range hs {
		newHs[i] = middlewareWrapper(hs[i])
	}

	return newHs
}
