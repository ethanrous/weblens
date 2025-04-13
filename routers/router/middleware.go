package router

import (
	"net/http"
	"strings"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/log"
	auth_service "github.com/ethanrous/weblens/services/auth"
	"github.com/ethanrous/weblens/services/context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const SessionTokenCookie = "weblens-session-token"

type ContextKey string

const (
	UserContextKey ContextKey = "user"
	ServerKey      ContextKey = "server"
	AllowPublicKey ContextKey = "allow_public"
	ServicesKey    ContextKey = "services"
	FuncNameKey    ContextKey = "func_name"
)

// func parseUserLogin(authHeader string, authService models.AccessService) (*user_model.User, error) {
// 	if len(authHeader) == 0 {
// 		return nil, werror.ErrNoAuth
// 	}
//
// 	return authService.GetUserFromToken(authHeader)
// }

// func parseApiKeyLogin(authHeader string, pack *models.ServicePack) (
// 	*user_model.User,
// 	error,
// ) {
// 	if len(authHeader) == 0 {
// 		return nil, werror.ErrNoAuth
// 	}
// 	authParts := strings.Split(authHeader, " ")
//
// 	if len(authParts) < 2 || authParts[0] != "Bearer" {
// 		// Bad auth header format
// 		return nil, werror.ErrBadAuth
// 	}
//
// 	key, err := pack.AccessService.GetApiKey(authParts[1])
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	usr := pack.UserService.Get(key.Owner)
// 	return usr, nil
// }

var ErrNotAuthenticated = errors.New("not authenticated")

func RequireSignIn(ctx *context.RequestContext) {
	if !ctx.IsLoggedIn {
		ctx.Error(http.StatusUnauthorized, ErrNotAuthenticated)
		return
	}
}

func RequireAdmin() func(context.RequestContext) {
	return func(ctx context.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsAdmin() {
			ctx.Error(http.StatusUnauthorized, ErrNotAuthenticated)
			return
		}
	}
}

func RequireOwner() func(context.RequestContext) {
	return func(ctx context.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsOwner() {
			ctx.Error(http.StatusUnauthorized, ErrNotAuthenticated)
			return
		}
	}
}

func WeblensAuth(next Handler) Handler {
	return HandlerFunc(func(ctx *context.RequestContext) {
		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get local instance"))
			return
		}

		if local.Role == tower_model.InitTowerRole {
			next.ServeHTTP(ctx)
			return
		}

		sessionCookie, err := ctx.GetCookie(SessionTokenCookie)
		if err == nil {
			// ctx.Log().Info().Msgf("Session cookie found: %s", sessionCookie)
			user, err := auth_service.GetUserFromJWT(ctx, sessionCookie)
			if err != nil {
				ctx.ExpireCookie()
				ctx.Error(http.StatusUnauthorized, errors.Wrap(err, "failed to validate sesion token"))
				return
			}
			ctx.Requester = user
			ctx.IsLoggedIn = true
		} else {
			// ctx.Log().Error().Err(err).Msg("Failed to get session cookie")
		}
		next.ServeHTTP(ctx)
		//
		//
		// 	// If we are still starting, allow all unauthenticated requests,
		// 	// but everyone is the public user
		// 	if !pack.Loaded.Load() || pack.InstanceService.GetLocal().GetRole() == models.InitServerRole {
		// 		r = r.WithContext(context.WithValue(r.Context(), UserContextKey, pack.UserService.GetPublicUser()))
		// 		log.Trace().Msg("Allowing unauthenticated request")
		// 		next.ServeHTTP(w, r)
		// 		return
		// 	}
		//
		// 	sessionCookie, err := ctx.GetCookie(SessionTokenCookie)
		//
		// 	if sessionCookie != nil && len(sessionCookie.Value) != 0 && err == nil {
		// 		pack.Log.Debug().Msg("Session cookie found")
		//
		// 		usr, err := parseUserLogin(sessionCookie.Value, pack.AccessService)
		// 		if err != nil {
		// 			log.Error().Stack().Err(err).Msg("")
		// 			if errors.Is(err, werror.ErrTokenExpired) {
		// 				ctx.ExpireCookie()
		// 			}
		// 			writeError(w, http.StatusUnauthorized, errors.Wrap(err, "failed to validate sesion token"))
		// 			return
		// 		}
		//
		// 		r = r.WithContext(context.WithValue(r.Context(), UserContextKey, usr))
		//
		// 		hlog.FromRequest(r).UpdateContext(func(c zerolog.Context) zerolog.Context {
		// 			return c.Str(string(UserContextKey), usr.GetUsername())
		// 		})
		// 		next.ServeHTTP(w, r)
		// 		return
		// 	}
		//
		// 	authHeader := r.Header["Authorization"]
		// 	if len(authHeader) != 0 {
		// 		usr, err := parseApiKeyLogin(authHeader[0], pack)
		// 		if SafeErrorAndExit(err, w, pack.Log) {
		// 			return
		// 		}
		//
		// 		serverId := r.Header.Get("Wl-Server-Id")
		// 		pack.Log.Debug().Msgf("Server ID: %s", serverId)
		//
		// 		if serverId != "" {
		// 			server := pack.InstanceService.GetByInstanceId(serverId)
		// 			if server != nil {
		// 				r = r.WithContext(context.WithValue(r.Context(), ServerKey, server))
		// 			}
		// 		}
		//
		// 		r = r.WithContext(context.WithValue(r.Context(), UserContextKey, usr))
		// 		hlog.FromRequest(r).UpdateContext(func(c zerolog.Context) zerolog.Context {
		// 			return c.Str(string(UserContextKey), usr.GetUsername())
		// 		})
		//
		// 		next.ServeHTTP(w, r)
		// 		return
		// 	}
		//
		// 	r = r.WithContext(context.WithValue(r.Context(), UserContextKey, pack.UserService.GetPublicUser()))
		// 	next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(proxyAddress string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", proxyAddress)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range, Cookie",
			)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Recoverer(next Handler) Handler {
	return HandlerFunc(func(ctx *context.RequestContext) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}
				log := hlog.FromRequest(ctx.Req)
				err, ok := rvr.(error)
				if !ok {
					err = errors.Errorf("Non-error panic in request handler: %v", rvr)
				}
				err = errors.WithStack(err)
				log.Error().Stack().Err(err).Msg("Recovered from panic in request handler")
				if ctx.Header("Connection") != "Upgrade" {
					ctx.Error(http.StatusInternalServerError, err)
				}
			}
		}()
		next.ServeHTTP(ctx)
	})
}

// URLHandler adds the requested URL as a field to the context's logger
// using fieldKey as field key.
func URLGroupHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			next.ServeHTTP(w, r)
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				route := chi.RouteContext(r.Context()).RoutePattern()
				return c.Str(fieldKey, route)
			})
		})
	}
}

func QueryParamHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			next.ServeHTTP(w, r)
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				// Get the query parameters from the request
				queryParams := r.URL.Query()
				// Convert the query parameters to a string representation
				for key, values := range queryParams {
					value := strings.Join(values, ",")
					c = c.Str(fieldKey+"_"+key, value)

				}
				return c
			})
		})
	}
}

func HeaderHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			next.ServeHTTP(w, r)

			if log.GetLevel() > zerolog.TraceLevel {
				return
			}

			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				// Get the query parameters from the request
				headers := r.Header
				// Convert the query parameters to a string representation
				for key, values := range headers {
					value := strings.Join(values, ",")
					c = c.Str(fieldKey+"_"+key, value)

				}
				return c
			})
		})
	}
}

func LoggerMiddlewares(logger zerolog.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		hlog.NewHandler(logger),
		// hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		// 	hlog.FromRequest(r).Info().
		// 		Str("method", r.Method).
		// 		Stringer("url", r.URL).
		// 		Int("status", status).
		// 		Int("size", size).
		// 		Dur("duration_ms", duration).
		// 		Msg("")
		// }),
		URLGroupHandler("url_group"),
		QueryParamHandler("query"),
		HeaderHandler("header"),
		hlog.RemoteIPHandler("ip"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("req_id", "Request-Id"),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

				next.ServeHTTP(ww, r)

				status := ww.Status()
				if status == 0 && r.Header.Get("Upgrade") == "websocket" {
					status = 101
				}

				remote := r.RemoteAddr
				method := r.Method
				timeTotal := time.Since(start)

				route := chi.RouteContext(r.Context()).RoutePattern()

				l := hlog.FromRequest(r)
				l.Info().Msgf("\u001B[0m[%s][%7s][%s %s][%s]", remote, log.ColorTime(timeTotal), method, route, log.ColorStatus(status))
			})
		}}
}
