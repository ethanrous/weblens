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
	authParts := strings.Split(authHeader, " ")

	if len(authParts) < 2 || authParts[0] != "Bearer" {
		// Bad auth header format
		return nil, werror.ErrBadAuthScheme
	}

	return authService.GetUserFromToken(authParts[1])
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

	if len(authParts) < 2 || authParts[0] != "X-Api-Key" {
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

func WeblensAuth(requireAdmin bool, pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.RequestsCounter.Inc()
		authHeader := c.Request.Header["Authorization"]
		if len(authHeader) != 0 {
			usr, err := ParseUserLogin(authHeader[0], pack.AccessService)
			if err != nil {
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

		authHeader = c.Request.Header["X-Api-Key"]
		if len(authHeader) != 0 {
			usr, _, err := ParseApiKeyLogin(authHeader[0], pack)
			if err != nil {
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("user", usr)
			c.Next()
			return
		}
	}
}

func KeyOnlyAuth(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["X-Api-Key"]
		if len(authHeader) != 0 {
			usr, _, err := ParseApiKeyLogin(authHeader[0], pack)
			if err != nil {
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("user", usr)
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
