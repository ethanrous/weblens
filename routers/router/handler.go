package router

import (
	"net/http"

	context_service "github.com/ethanrous/weblens/services/context"
)

type HandlerFunc func(ctx context_service.RequestContext)

type PassthroughHandler func(Handler) Handler

type Handler interface {
	ServeHTTP(ctx context_service.RequestContext)
}

func (f HandlerFunc) ServeHTTP(ctx context_service.RequestContext) {
	f(ctx)
}

func getFromHTTP(w http.ResponseWriter, r *http.Request) context_service.RequestContext {
	reqCtx, ok := context_service.ReqFromContext(r.Context())
	if !ok {
		panic("request context not found in request")
	}

	reqCtx.ReqCtx = r.Context()
	r.WithContext(reqCtx)
	reqCtx.Req = r
	reqCtx.W = w

	return reqCtx
	// ctx, _ := r.Context().Value(requestContextKey).(context_service.RequestContext)
	// ctx.Req = r
	// ctx.W = w
	// return ctx
}

func toStdHandlerFunc(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := getFromHTTP(w, r)
		h.ServeHTTP(ctx)
	}
}

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
