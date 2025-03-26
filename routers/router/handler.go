package router

import (
	"net/http"

	"github.com/ethanrous/weblens/services/context"
)

type HandlerFunc func(ctx context.RequestContext)

func handlerWrapper(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.GetFromHTTP(r)
		h(ctx)
	}
}

func middlewareWrapper(h HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.GetFromHTTP(r)
			h(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// func wrapHandlerProvider[T http.Handler](hp func(next http.Handler) T, funcInfo *routing.FuncInfo) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		h := hp(next) // this handle could be dynamically generated, so we can't use it for debug info
// 		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
// 			defer routing.RecordFuncInfo(req.Context(), funcInfo)()
// 			h.ServeHTTP(resp, req)
// 		})
// 	}
// }
//
// func toHandlerProvider(handler any) func(next http.Handler) http.Handler {
// 	funcInfo := routing.GetFuncInfo(handler)
// 	fn := reflect.ValueOf(handler)
// 	if fn.Type().Kind() != reflect.Func {
// 		panic(fmt.Sprintf("handler must be a function, but got %s", fn.Type()))
// 	}
//
// 	if hp, ok := handler.(func(next http.Handler) http.Handler); ok {
// 		return wrapHandlerProvider(hp, funcInfo)
// 	} else if hp, ok := handler.(func(http.Handler) http.HandlerFunc); ok {
// 		return wrapHandlerProvider(hp, funcInfo)
// 	}
//
// 	provider := func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(respOrig http.ResponseWriter, req *http.Request) {
// 			// wrap the response writer to check whether the response has been written
// 			resp := respOrig
// 			if _, ok := resp.(types.ResponseStatusProvider); !ok {
// 				resp = &responseWriter{respWriter: resp}
// 			}
//
// 			// prepare the arguments for the handler and do pre-check
// 			argsIn := prepareHandleArgsIn(resp, req, fn, funcInfo)
// 			if req == nil {
// 				preCheckHandler(fn, argsIn)
// 				return // it's doing pre-check, just return
// 			}
//
// 			defer routing.RecordFuncInfo(req.Context(), funcInfo)()
// 			ret := fn.Call(argsIn)
//
// 			// handle the return value (no-op at the moment)
// 			handleResponse(fn, ret)
//
// 			// if the response has not been written, call the next handler
// 			if next != nil && !hasResponseBeenWritten(argsIn) {
// 				next.ServeHTTP(resp, req)
// 			}
// 		})
// 	}
//
// 	provider(nil).ServeHTTP(nil, nil) // do a pre-check to make sure all arguments and return values are supported
// 	return provider
// }
