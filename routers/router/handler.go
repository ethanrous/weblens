package router

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
)

type HandlerFunc func(ctx *context.RequestContext)

func getFromHTTP(r *http.Request) *context.RequestContext {
	ctx, _ := r.Context().Value(requestContextKey).(*context.RequestContext)
	return ctx
}

func toStdHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := getFromHTTP(r)
		if ctx == nil {
			panic(errors.New("request context is nil"))
		}
		h(ctx)
	}
}

func middlewareWrapper(h HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := getFromHTTP(r)
			h(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func wrapHandlerProvider[T http.Handler](hp func(next http.Handler) T) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := hp(next) // this handle could be dynamically generated, so we can't use it for debug info
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(resp, req)
		})
	}
}

// toHandlerProvider converts a handler to a handler provider
// A handler provider is a function that takes a "next" http.Handler, it can be used as a middleware
func toHandlerProvider(handler any) func(next http.Handler) http.Handler {
	fn := reflect.ValueOf(handler)
	if fn.Type().Kind() != reflect.Func {
		panic(fmt.Sprintf("handler must be a function, but got %s", fn.Type()))
	}

	if hp, ok := handler.(func(next http.Handler) http.Handler); ok {
		return wrapHandlerProvider(hp)
	} else if hp, ok := handler.(func(http.Handler) http.HandlerFunc); ok {
		return wrapHandlerProvider(hp)
	}

	provider := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(r http.ResponseWriter, req *http.Request) {
			// wrap the response writer to check whether the response has been written

			// prepare the arguments for the handler and do pre-check
			// argsIn := prepareHandleArgsIn(r, req, fn, funcInfo)
			// if req == nil {
			// 	preCheckHandler(fn, argsIn)
			// 	return // it's doing pre-check, just return
			// }
			//
			// defer routing.RecordFuncInfo(req.Context(), funcInfo)()
			// ret := fn.Call(argsIn)
			//
			// // handle the return value (no-op at the moment)
			// handleResponse(fn, ret)

			// if the response has not been written, call the next handler
			if next != nil {
				next.ServeHTTP(r, req)
			}
		})
	}

	provider(nil).ServeHTTP(nil, nil) // do a pre-check to make sure all arguments and return values are supported
	return provider
}
