package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/metrics"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
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

func withServices(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("services", pack)
		c.Next()
	}
}

func AllowPublic() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("allow_public", true)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {

		userI, ok := c.Get("user")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, ok := userI.(*models.User)
		if !ok {
			log.Error.Println(werror.Errorf("Could not assert user from context in RequireAdmin"))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !user.IsAdmin() {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

const SessionTokenCookie = "weblens-session-token"

func WeblensAuth(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.RequestsCounter.Inc()

		// If we are still starting, allow all unauthenticated requests,
		// but everyone is the public user
		if !pack.Loaded.Load() || pack.InstanceService.GetLocal().GetRole() == models.InitServer {
			c.Set("user", pack.UserService.GetPublicUser())
			log.Trace.Println("Allowing unauthenticated request")
			c.Next()
			return
		}

		sessionToken, err := c.Cookie(SessionTokenCookie)

		if len(sessionToken) != 0 && err == nil {
			usr, err := ParseUserLogin(sessionToken, pack.AccessService)
			if err != nil {
				log.ShowErr(err)
				if errors.Is(err, werror.ErrTokenExpired) {
					cookie := fmt.Sprintf("%s=;Path=/;expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
					c.Header("Set-Cookie", cookie)
				}
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			log.Trace.Printf("User [%s] authenticated", usr.GetUsername())

			c.Set("user", usr)
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if len(authHeader) != 0 {
			usr, server, err := ParseApiKeyLogin(authHeader, pack)
			if err != nil {
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			if server != nil {
				c.Set("server", server)
			} else {
				c.Set("user", usr)
			}
			c.Next()
			return
		}

		if pack.InstanceService.GetLocal().GetRole() == models.BackupServer {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !c.GetBool("allow_public") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", pack.UserService.GetPublicUser())
		c.Next()
	}
}

func KeyOnlyAuth(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["Authorization"]
		if len(authHeader) != 0 {
			_, server, err := ParseApiKeyLogin(authHeader[0], pack)
			if err != nil {
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if server == nil {
				log.Warning.Println(werror.Errorf("Got nil server in KeyOnlyAuth"))
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("server", server)
			c.Next()
			return
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	host := env.GetProxyAddress()
	// host = "http://local.weblens.io:8080"
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", host)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range, Cookie",
		)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
