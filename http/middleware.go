package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

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

	var i *models.Instance
	if key.RemoteUsing != "" {
		i = pack.InstanceService.GetByInstanceId(key.RemoteUsing)
	}

	usr := pack.UserService.Get(key.Owner)
	return usr, i, nil
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

		u, err := getUserFromCtx(r)
		if SafeErrorAndExit(err, w) {
			return
		}

		if u != nil && u.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		SafeErrorAndExit(werror.ErrNotAdmin, w)
		return
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

		u, err := getUserFromCtx(r)
		if SafeErrorAndExit(err, w) {
			return
		}

		if u != nil && u.IsAdmin() {
			next.ServeHTTP(w, r)
			return
		}

		SafeErrorAndExit(werror.ErrNotOwner, w)
		return
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
					cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
					w.Header().Set("Set-Cookie", cookie)
				}
				SafeErrorAndExit(err, w)
				return
			}

			log.Trace.Func(func(l log.Logger) { l.Printf("User [%s] authenticated", usr.GetUsername()) })

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
			}
			r = r.WithContext(context.WithValue(r.Context(), UserKey, usr))
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

				err, ok := rvr.(error)
				if ok {
					log.ErrTrace(err)
				} else {
					log.Error.Println("HTTP PANIC\n", rvr)
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
