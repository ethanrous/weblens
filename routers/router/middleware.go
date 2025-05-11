package router

import (
	"net/http"
	"strings"
	"time"

	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	auth_service "github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

var ErrNotAuthenticated = errors.New("not authenticated")
var ErrNotAuthorized = errors.New("not authorized")

func RequireSignIn(ctx context_service.RequestContext) {
	if !ctx.IsLoggedIn {
		ctx.Error(http.StatusUnauthorized, ErrNotAuthenticated)

		return
	}
}

func RequireAdmin(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsAdmin() {
			ctx.Error(http.StatusUnauthorized, errors.Wrap(ErrNotAuthorized, "not an admin"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

func RequireOwner(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsOwner() {
			ctx.Error(http.StatusUnauthorized, errors.Wrap(ErrNotAuthorized, "not an owner"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

func RequireCoreTower(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get local instance"))

			return
		}

		if local.Role != tower_model.RoleCore {
			ctx.Error(http.StatusUnauthorized, errors.New("endpoint is not allowed when not a core tower"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

func ShareInjector(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		shareId := ctx.Query("shareId")
		if shareId != "" {
			share, err := share_model.GetShareById(ctx, shareId)
			if err != nil {
				ctx.Error(http.StatusNotFound, errors.Wrap(err, "failed to get share"))

				return
			}

			ctx.Share = share
		}

		next.ServeHTTP(ctx)
	})
}

func WeblensAuth(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "failed to get local instance"))

			return
		}

		if local.Role == tower_model.RoleInit {
			next.ServeHTTP(ctx)

			return
		}

		ctx.Requester = user_model.GetPublicUser()

		remoteTowerId := ctx.Header(tower_service.TowerIdHeader)
		if remoteTowerId != "" {
			remote, err := tower_model.GetTowerById(ctx, remoteTowerId)
			if err != nil {
				ctx.Error(http.StatusNotFound, errors.Wrapf(err, "failed to get remote instance [%s]", remoteTowerId))

				return
			}

			ctx.Remote = remote
		}

		authHeader := ctx.Header("Authorization")
		if authHeader != "" {
			usr, err := auth_service.GetUserFromAuthHeader(ctx, authHeader)
			if err != nil {
				ctx.Error(http.StatusUnauthorized, errors.Wrap(err, "failed to validate authorization header"))

				return
			}

			ctx.Requester = usr
			ctx.IsLoggedIn = true

			next.ServeHTTP(ctx)

			return
		}

		if ctx.Remote.TowerId != "" {
			ctx.Error(http.StatusUnauthorized, errors.Wrap(ErrNotAuthenticated, "towers must authenticate with a token"))

			return
		}

		sessionCookie, err := ctx.GetCookie(SessionTokenCookie)
		if err == nil {
			usr, err := auth_service.GetUserFromJWT(ctx, sessionCookie)
			if err != nil {
				ctx.ExpireCookie()
				ctx.Error(http.StatusUnauthorized, errors.Wrap(err, "failed to validate sesion token"))

				return
			}

			ctx.Requester = usr
			ctx.IsLoggedIn = true

			log.FromContext(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("requester", usr.Username)
			})

			// ctx = ctx.WithContext(logCtx).(context_service.RequestContext)

			next.ServeHTTP(ctx)

			return
		}

		// Do public
		next.ServeHTTP(ctx)
	})
}

func CORSMiddleware(next Handler) Handler {
	proxyAddress := config.GetConfig().ProxyAddress

	return HandlerFunc(func(ctx context_service.RequestContext) {
		ctx.SetHeader("Access-Control-Allow-Origin", proxyAddress)
		ctx.SetHeader("Access-Control-Allow-Credentials", "true")
		ctx.SetHeader(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range, Cookie",
		)
		ctx.SetHeader("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if ctx.Req.Method == http.MethodOptions {
			ctx.Status(http.StatusNoContent)

			return
		}

		next.ServeHTTP(ctx)
	})
}

func Recoverer(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				err, ok := rvr.(error)
				if !ok {
					err = errors.Errorf("Non-error panic in request handler: %v", rvr)
				}

				err = errors.WithStack(err)
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
				l := hlog.FromRequest(r).With().Logger()
				r = r.WithContext(log.WithContext(r.Context(), &l))

				next.ServeHTTP(ww, r)

				status := ww.Status()
				if status == 0 && r.Header.Get("Upgrade") == "websocket" {
					status = http.StatusSwitchingProtocols
				}

				remote := r.RemoteAddr
				method := r.Method
				timeTotal := time.Since(start)

				route := log.RouteColor(r)

				l.Info().Msgf("\u001B[0m[%s][%7s][%s %s][%s]", remote, log.ColorTime(timeTotal), method, route, log.ColorStatus(status))
			})
		}}
}
