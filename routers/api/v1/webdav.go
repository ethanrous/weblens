package v1

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/models/featureflags"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/services"
	"github.com/ethanrous/weblens/services/ctxservice"
	"golang.org/x/net/webdav"
)

// WebDAVRouter returns an http.Handler that serves WebDAV requests.
// It checks the WebDAV feature flag, authenticates via HTTP Basic Auth,
// and scopes file access to the authenticated user's home directory.
func WebDAVRouter(appCtx ctxservice.AppContext) router.HandlerFunc {
	davFS := &services.WebdavFs{FileService: appCtx.FileService}

	davHandler := &webdav.Handler{
		Prefix:     "",
		FileSystem: davFS,
		LockSystem: webdav.NewMemLS(),
		Logger: func(_ *http.Request, err error) {
			if err != nil {
				appCtx.Log().Error().Err(err).Msg("WebDAV request error")
			}
		},
	}

	return router.HandlerFunc(getWebDAVHandlerFunc(davHandler))
}

// FIXME: This is AWFUL. I cannot imagine the security implications of this, but I cannot think of
// another way to stop the webdav from waiting ~2s EVERY request while it bcrypts the password.
var userCache = make(map[string]*user_model.User)
var userCacheLock = sync.RWMutex{}

func getWebDAVHandlerFunc(davHandler *webdav.Handler) func(ctxservice.RequestContext) {
	return func(ctx ctxservice.RequestContext) {
		// Check feature flag
		flags, err := featureflags.GetFlags(ctx)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		if !flags.EnableWebDAV {
			ctx.JSON(http.StatusServiceUnavailable, "WebDAV is disabled")

			return
		}

		// HTTP Basic Auth
		username, password, ok := ctx.Req.BasicAuth()
		if !ok {
			ctx.SetHeader("WWW-Authenticate", `Basic realm="Weblens WebDAV"`)
			ctx.Error(http.StatusUnauthorized, wlerrors.Errorf("Unauthorized"))

			return
		}

		var user *user_model.User

		var found bool

		userCacheLock.Lock()

		// FIXME: BAD VERY BAD SO VERY BAD. RAINBOW TABLE ATTACK IMMINENT. But better than being slow, I guess??
		cacheKey := username + password + strings.Split(ctx.Req.RemoteAddr, ":")[0] // Include IP to mitigate attacks
		if user, found = userCache[cacheKey]; !found {
			user, err = user_model.GetUserByUsername(ctx, username)
			if err != nil {
				time.Sleep(2 * time.Second) // Mitigate rainbow table attacks by adding a delay on failed lookups
			}

			if err != nil || !user.CheckLogin(password) {
				ctx.SetHeader("WWW-Authenticate", `Basic realm="Weblens WebDAV"`)
				ctx.Error(http.StatusUnauthorized, wlerrors.Errorf("Unauthorized"))

				return
			}

			ctx.Log().Info().Str("username", username).Msg("Authenticated WebDAV user")

			userCache[cacheKey] = user
		}

		userCacheLock.Unlock()

		// Inject user into context for the filesystem adapter
		newCtx := services.WithWebDAVUser(ctx, user)

		ctx.Req.URL.Path = strings.TrimPrefix(ctx.Req.URL.Path, "/webdav")
		davHandler.ServeHTTP(ctx.W, ctx.Req.WithContext(newCtx))
	}
}
