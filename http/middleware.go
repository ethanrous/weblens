package http

import (
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/metrics"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func ParseUserLogin(authHeader string, authService models.AccessService) (*models.User, error) {
	if len(authHeader) == 0 {
		return nil, werror.ErrNoAuth
	}
	// authParts := strings.Split(authHeader, "=")

	// if len(authParts) < 2 || authParts[0] != "Bearer" {
	// 	// Bad auth header format
	// 	return nil, werror.ErrBadAuthScheme
	// }

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

	key, err := pack.AccessService.GetApiKey(models.WeblensApiKey(authParts[1]))
	if err != nil {
		return nil, nil, err
	}

	if key.RemoteUsing != "" {
		i := pack.InstanceService.Get(key.RemoteUsing)
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

func WeblensAuth(requireAdmin, allowBadAuth bool, pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.RequestsCounter.Inc()

		// If we are still starting, allow all unauthenticated requests,
		// but everyone is the public user
		if !pack.Loaded.Load() {
			c.Set("user", pack.UserService.GetPublicUser())
			c.Next()
			return
		}

		sessionToken, err := c.Cookie("weblens-session-token")

		if len(sessionToken) != 0 && err == nil {
			usr, err := ParseUserLogin(sessionToken, pack.AccessService)
			if err != nil {
				if allowBadAuth {
					c.Next()
					return
				}
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			if requireAdmin && (usr == nil || !usr.IsAdmin()) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

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

		// apiKey := c.Request.Header["X-Api-Key"][0]
		// if len(apiKey) != 0 {
		// 	usr, _, err := ParseApiKeyLogin(apiKey, pack)
		// 	if err != nil {
		// 		log.ShowErr(err)
		// 		c.AbortWithStatus(http.StatusUnauthorized)
		// 		return
		// 	}
		//
		// 	c.Set("user", usr)
		// 	c.Next()
		// 	return
		// }
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
				log.Error.Println("Verified key-only login, but did not get server")
				c.AbortWithStatus(http.StatusInternalServerError)
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
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range",
		)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
