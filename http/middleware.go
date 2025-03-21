package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
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

func parseUserLogin(authHeader string, authService models.AccessService) (*models.User, error) {
	if len(authHeader) == 0 {
		return nil, werror.ErrNoAuth
	}

	return authService.GetUserFromToken(authHeader)
}

func parseApiKeyLogin(authHeader string, pack *models.ServicePack) (
	*models.User,
	error,
) {
	if len(authHeader) == 0 {
		return nil, werror.ErrNoAuth
	}
	authParts := strings.Split(authHeader, " ")

	if len(authParts) < 2 || authParts[0] != "Bearer" {
		// Bad auth header format
		return nil, werror.ErrBadAuth
	}

	key, err := pack.AccessService.GetApiKey(authParts[1])
	if err != nil {
		return nil, err
	}

	usr := pack.UserService.Get(key.Owner)
	return usr, nil
}

func WithServices(pack *models.ServicePack) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ServicesKey, pack))
			next.ServeHTTP(w, r)
		})
	}
}

func AllowPublic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), AllowPublicKey, true))
		next.ServeHTTP(w, r)
	})
}

func RequireSignIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := getUserFromCtx(r, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		if u.IsPublic() {
			writeError(w, werror.ErrNotAuthenticated.Code(), werror.ErrNotAuthenticated)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := getInstanceFromCtx(r)
		if server != nil {
			next.ServeHTTP(w, r)
			return
		}

		u, err := getUserFromCtx(r, true)
		if SafeErrorAndExit(err, w) {
			return
		}

		if u != nil && u.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		SafeErrorAndExit(werror.ErrNotAdmin, w)
	})
}

func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pack := getServices(r)

		server := getInstanceFromCtx(r)
		if server != nil {

			key, err := pack.AccessService.GetApiKey(server.GetUsingKey())
			if err != nil {
				SafeErrorAndExit(err, w)
				return
			}

			owner := pack.UserService.Get(key.Owner)
			if owner == nil || !owner.IsOwner() {
				SafeErrorAndExit(werror.ErrNotOwner, w)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		u, err := getUserFromCtx(r, true)
		if SafeErrorAndExit(err, w) {
			return
		}

		if u != nil && u.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		SafeErrorAndExit(werror.ErrNotOwner, w)
	})
}

func WeblensAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pack := getServices(r)

		// If we are still starting, allow all unauthenticated requests,
		// but everyone is the public user
		if !pack.Loaded.Load() || pack.InstanceService.GetLocal().GetRole() == models.InitServerRole {
			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, pack.UserService.GetPublicUser()))
			log.Trace().Msg("Allowing unauthenticated request")
			next.ServeHTTP(w, r)
			return
		}

		sessionCookie, err := r.Cookie(SessionTokenCookie)

		if sessionCookie != nil && len(sessionCookie.Value) != 0 && err == nil {
			pack.Log.Debug().Msg("Session cookie found")

			usr, err := parseUserLogin(sessionCookie.Value, pack.AccessService)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				if errors.Is(err, werror.ErrTokenExpired) {
					cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
					w.Header().Set("Set-Cookie", cookie)
				}
				writeError(w, http.StatusUnauthorized, errors.Wrap(err, "failed to validate sesion token"))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, usr))

			hlog.FromRequest(r).UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(string(UserContextKey), usr.GetUsername())
			})
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header["Authorization"]
		if len(authHeader) != 0 {
			usr, err := parseApiKeyLogin(authHeader[0], pack)
			if SafeErrorAndExit(err, w, pack.Log) {
				return
			}

			serverId := r.Header.Get("Wl-Server-Id")
			pack.Log.Debug().Msgf("Server ID: %s", serverId)

			if serverId != "" {
				server := pack.InstanceService.GetByInstanceId(serverId)
				if server != nil {
					r = r.WithContext(context.WithValue(r.Context(), ServerKey, server))
				}
			}

			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, usr))
			hlog.FromRequest(r).UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(string(UserContextKey), usr.GetUsername())
			})

			next.ServeHTTP(w, r)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserContextKey, pack.UserService.GetPublicUser()))
		next.ServeHTTP(w, r)
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

func Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				log := hlog.FromRequest(r)

				err, ok := rvr.(error)
				if ok {
					err = errors.WithStack(err)
					log.Error().Stack().Err(err).Msg("Recovered from panic in request handler")
				} else {
					err = errors.Errorf("Unknown panic in request handler: %v", rvr)
					log.Error().Stack().Err(err).Msg("")
				}

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
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
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}),
		URLGroupHandler("url_group"),
		QueryParamHandler("query"),
		HeaderHandler("header"),
		hlog.RemoteIPHandler("ip"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("req_id", "Request-Id"),
	}
}
