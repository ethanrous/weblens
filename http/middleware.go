package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
)

const SessionTokenCookie = "weblens-session-token"

type ContextKey string

const (
	UserKey        ContextKey = "user"
	ServerKey      ContextKey = "server"
	AllowPublicKey ContextKey = "allow_public"
	ServicesKey    ContextKey = "services"
	FuncNameKey    ContextKey = "func_name"
)

func ParseUserLogin(authHeader string, authService models.AccessService) (*models.User, error) {
	if len(authHeader) == 0 {
		return nil, werror.ErrNoAuth
	}

	return authService.GetUserFromToken(authHeader)
}

func ParseApiKeyLogin(authHeader string, pack *models.ServicePack) (
	*models.User,
	*models.Instance,
	error,
) {
	if len(authHeader) == 0 {
		return nil, nil, werror.ErrNoAuth
	}
	authParts := strings.Split(authHeader, " ")

	if len(authParts) < 2 || authParts[0] != "Bearer" {
		// Bad auth header format
		return nil, nil, werror.ErrBadAuthScheme
	}

	key, err := pack.AccessService.GetApiKey(authParts[1])
	if err != nil {
		return nil, nil, err
	}

	if key.RemoteUsing != "" {
		i := pack.InstanceService.GetByInstanceId(key.RemoteUsing)
		return nil, i, nil
	}

	usr := pack.UserService.Get(key.Owner)
	return usr, nil, nil
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

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := getInstanceFromCtx(r)
		if server != nil {
			next.ServeHTTP(w, r)
			return
		}

		u, err := getUserFromCtx(w, r)
		if SafeErrorAndExit(err, w) {
			return
		}

		if u != nil && u.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		log.Error.Println("Unauthorized request")
		w.WriteHeader(http.StatusForbidden)
	})
}

func WithFuncName(next http.Handler) http.Handler {
	if log.GetLogLevel() >= log.DEBUG {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug.Println("Setting func name")
			funcName := runtime.FuncForPC(reflect.ValueOf(next).Pointer()).Name()
			funcName = funcName[strings.LastIndex(funcName, "/")+1 : strings.LastIndex(funcName, ".")]
			r = r.WithContext(context.WithValue(r.Context(), FuncNameKey, funcName))
			next.ServeHTTP(w, r)
		})
	}
	return next
}

func WeblensAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pack := getServices(r)

		// If we are still starting, allow all unauthenticated requests,
		// but everyone is the public user
		if !pack.Loaded.Load() || pack.InstanceService.GetLocal().GetRole() == models.InitServerRole {
			r = r.WithContext(context.WithValue(r.Context(), UserKey, pack.UserService.GetPublicUser()))
			log.Trace.Println("Allowing unauthenticated request")
			next.ServeHTTP(w, r)
			return
		}

		sessionCookie, err := r.Cookie(SessionTokenCookie)

		if sessionCookie != nil && len(sessionCookie.Value) != 0 && err == nil {
			usr, err := ParseUserLogin(sessionCookie.Value, pack.AccessService)
			if err != nil {
				log.ShowErr(err)
				if errors.Is(err, werror.ErrTokenExpired) {
					cookie := fmt.Sprintf("%s=;Path=/;expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
					w.Header().Set("Set-Cookie", cookie)
				}
				SafeErrorAndExit(err, w)
				return
			}

			log.Trace.Printf("User [%s] authenticated", usr.GetUsername())

			r = r.WithContext(context.WithValue(r.Context(), UserKey, usr))
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header["Authorization"]
		if len(authHeader) != 0 {
			usr, server, err := ParseApiKeyLogin(authHeader[0], pack)
			if err != nil {
				log.ShowErr(err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if server != nil {
				r = r.WithContext(context.WithValue(r.Context(), ServerKey, server))

			} else {
				r = r.WithContext(context.WithValue(r.Context(), UserKey, usr))
			}
			next.ServeHTTP(w, r)
			return
		}

		if pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// allowPublic, ok := r.Context().Value("allow_public").(bool)
		// if !ok || !allowPublic {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		r = r.WithContext(context.WithValue(r.Context(), UserKey, pack.UserService.GetPublicUser()))
		next.ServeHTTP(w, r)
	})
}

func KeyOnlyAuth(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pack, ok := r.Context().Value(ServicesKey).(*models.ServicePack)
		if pack == nil || !ok {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error.Println(werror.Errorf("Could not assert services from context in WeblensAuth"))
			return
		}

		authHeader := r.Header["Authorization"]
		if len(authHeader) != 0 {
			_, server, err := ParseApiKeyLogin(authHeader[0], pack)
			if err != nil {
				log.ShowErr(err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if server == nil {
				log.Warning.Println(werror.Errorf("Got nil server in KeyOnlyAuth"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), ServerKey, server))
			next.ServeHTTP(w, r)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	host := env.GetProxyAddress()
	// host = "http://local.weblens.io:8080"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", host)
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
