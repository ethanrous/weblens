package router

import (
	"net/http"
	"strings"
	"time"

	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
	auth_service "github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SessionTokenCookie defines the cookie name used for storing session authentication tokens.
const SessionTokenCookie = "weblens-session-token"

// ContextKey defines the type for context keys used in request contexts.
type ContextKey string

const (
	// UserContextKey is the context key for storing user information.
	UserContextKey ContextKey = "user"
	// ServerKey is the context key for storing server information.
	ServerKey ContextKey = "server"
	// AllowPublicKey is the context key for indicating if public access is allowed.
	AllowPublicKey ContextKey = "allow_public"
	// ServicesKey is the context key for storing service references.
	ServicesKey ContextKey = "services"
	// FuncNameKey is the context key for storing the function name being executed.
	FuncNameKey ContextKey = "func_name"
)

// ErrNotAuthenticated indicates that the request lacks valid authentication credentials.
var ErrNotAuthenticated = wlerrors.New("not authenticated")

// ErrNotAuthorized indicates that the authenticated user lacks permission for the requested action.
var ErrNotAuthorized = wlerrors.New("not authorized")

// RequireSignIn returns a middleware that ensures the requester is authenticated before proceeding.
func RequireSignIn(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		if !ctx.IsLoggedIn {
			ctx.Log().Trace().Msg("Expected authenticated user, but none found, returning 401")

			ctx.Error(http.StatusUnauthorized, ErrNotAuthenticated)

			return
		}

		next.ServeHTTP(ctx)
	})
}

// RequireAdmin returns a middleware that ensures the requester has admin privileges before proceeding.
func RequireAdmin(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsAdmin() {
			ctx.Error(http.StatusUnauthorized, wlerrors.Wrap(ErrNotAuthorized, "not an admin"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

// RequireOwner returns a middleware that ensures the requester is the owner before proceeding.
func RequireOwner(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		if ctx.Requester == nil || !ctx.Requester.IsOwner() {
			ctx.Error(http.StatusUnauthorized, wlerrors.Wrap(ErrNotAuthorized, "not an owner"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

// RequireCoreTower returns a middleware that ensures the local tower is in core role before proceeding.
func RequireCoreTower(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get local instance"))

			return
		}

		if local.Role != tower_model.RoleCore {
			ctx.Error(http.StatusUnauthorized, wlerrors.New("endpoint is not allowed when not a core tower"))

			return
		}

		next.ServeHTTP(ctx)
	})
}

// ShareInjector returns a middleware that loads share information from the shareID query parameter into the request context.
func ShareInjector(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		shareIDStr := ctx.Query("shareID")

		shareID, err := primitive.ObjectIDFromHex(shareIDStr)
		if err == nil && !shareID.IsZero() {
			share, err := share_model.GetShareByID(ctx, shareID)
			if err != nil {
				ctx.Error(http.StatusNotFound, wlerrors.Wrap(err, "failed to get share"))

				return
			}

			ctx.Log().Debug().Msgf("Share found: %s", shareIDStr)

			ctx.Share = share
		}

		next.ServeHTTP(ctx)
	})
}

// WeblensAuth returns a middleware that handles authentication for Weblens requests using auth headers, session tokens, or tower credentials.
func WeblensAuth(next Handler) Handler {
	return HandlerFunc(func(ctx context_service.RequestContext) {
		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to get local instance"))

			return
		}

		if local.Role == tower_model.RoleUninitialized {
			next.ServeHTTP(ctx)

			return
		}

		ctx = ctx.WithRequester(user_model.GetPublicUser())

		remoteTowerID := ctx.Header(tower_service.TowerIDHeader)
		if remoteTowerID != "" {
			remote, err := tower_model.GetTowerByID(ctx, remoteTowerID)
			if err != nil {
				ctx.Error(http.StatusNotFound, wlerrors.Wrapf(err, "failed to get remote instance [%s]", remoteTowerID))

				return
			}

			ctx.Remote = remote
		}

		authHeader := ctx.Header("Authorization")
		if authHeader != "" {
			ctx.Log().Trace().Msg("Authorization header found, attempting to authenticate via header")

			usr, err := auth_service.GetUserFromAuthHeader(ctx, authHeader)
			if err != nil {
				ctx.Error(http.StatusUnauthorized, wlerrors.Wrap(err, "failed to validate authorization header"))

				return
			}

			ctx.Log().Trace().Msgf("Authenticated user via auth header: %s", usr.Username)

			ctx = ctx.WithRequester(usr)
		} else if ctx.Remote.TowerID != "" {
			ctx.Error(http.StatusUnauthorized, wlerrors.Wrap(ErrNotAuthenticated, "towers must authenticate with a token"))

			return
		}

		// Regular session token
		if !ctx.IsLoggedIn {
			sessionCookie, err := ctx.GetCookie(SessionTokenCookie)
			if err == nil {
				usr, err := auth_service.GetUserFromJWT(ctx, sessionCookie)
				if err != nil {
					ctx.ExpireCookie()
					ctx.Error(http.StatusUnauthorized, wlerrors.WrapStatus(http.StatusUnauthorized, wlerrors.Wrap(err, "failed to validate sesion token")))

					return
				}

				ctx.Log().Trace().Msgf("Authenticated user via session token: %s", usr.Username)

				ctx = ctx.WithRequester(usr)
			}
		}

		log.FromContext(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("requester", ctx.Requester.Username)
		})

		next.ServeHTTP(ctx)
	})
}

// CORSMiddleware returns a middleware that sets CORS headers for cross-origin requests.
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

// Recoverer returns a middleware that recovers from panics and returns appropriate error responses.
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
					err = wlerrors.Errorf("Non-error panic in request handler: %v", rvr)
				}

				err = wlerrors.WithStack(err)
				if ctx.Header("Connection") != "Upgrade" {
					ctx.Error(http.StatusInternalServerError, err)
				}
			}
		}()

		next.ServeHTTP(ctx)
	})
}

// URLGroupHandler adds the route pattern as a field to the context's logger using fieldKey as field key.
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

// QueryParamHandler adds query parameters as fields to the context's logger using fieldKey as a prefix.
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

// HeaderHandler adds request headers as fields to the context's logger when trace logging is enabled.
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

// NewHandler injects log into request contexts.
// Each request gets its own copy of the logger to avoid data races
// when UpdateContext is called by downstream middleware.
func NewHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appCtx, ok := context_service.FromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)

				return
			}

			// Create a copy of the logger for this request to avoid races
			// when UpdateContext is called concurrently by multiple requests
			logger := appCtx.Log().With().Logger()
			r = r.WithContext(logger.WithContext(r.Context()))

			next.ServeHTTP(w, r)
		})
	}
}

// LoggerMiddlewares returns a collection of middleware functions that handle request logging.
func LoggerMiddlewares() []func(http.Handler) http.Handler {
	doDevLog := config.GetConfig().LogFormat == "dev"

	return []func(http.Handler) http.Handler{
		NewHandler(),
		URLGroupHandler("url_group"),
		QueryParamHandler("query"),
		HeaderHandler("header"),
		hlog.RemoteIPHandler("ip"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("req_id", "Request-ID"),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
				l := zerolog.Ctx(r.Context()).With().Logger()
				r = r.WithContext(log.WithContext(r.Context(), &l))

				next.ServeHTTP(ww, r)

				status := ww.Status()
				if status == 0 && r.Header.Get("Upgrade") == "websocket" {
					status = http.StatusSwitchingProtocols
				}

				remote := r.RemoteAddr
				method := r.Method
				timeTotal := time.Since(start)

				if doDevLog {
					route := log.RouteColor(r)

					l.Info().Msgf("\u001B[0m[%s][%7s][%s %s][%s]", remote, log.ColorTime(timeTotal), method, route, log.ColorStatus(status))

					return
				}

				l.Info().
					Str("method", r.Method).
					Str("route_group", chi.RouteContext(r.Context()).RoutePattern()).
					Str("ip_addr", r.RemoteAddr).
					Int("status", status).
					Dur("req_duration", timeTotal).
					Msg(r.URL.String())
			})
		}}
}
