package http

import (
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/metrics"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func parseAuthHeader(authHeaderParts []string) (string, string, error) {
	if len(authHeaderParts) == 0 || len(authHeaderParts[0]) == 0 {
		// Public login
		return "", "", nil
	}
	authString := authHeaderParts[0]
	authList := strings.Split(authString, " ")
	if len(authList) < 2 {
		// Bad auth header format
		return "", "", werror.ErrBadAuthScheme
	} else if authList[0] != "Basic" && authList[0] != "Bearer" {
		// Bad auth header scheme
		return "", "", werror.ErrBadAuthScheme
	}

	// Pass
	return authList[0], authList[1], nil
}

func withServices(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("services", pack)
	}
}

func validateBasicAuth(cred string, userService models.UserService) (*models.User, error) {
	credB, err := base64.StdEncoding.DecodeString(cred)
	if err != nil {
		log.ErrTrace(err)
		return nil, err
	}
	userAndToken := strings.Split(string(credB), ":")

	if len(userAndToken) != 2 {
		return nil, werror.ErrBasicAuthFormat
	}

	u := userService.Get(models.Username(userAndToken[0]))
	if u == nil {
		if userService.Size() == 0 {
			return nil, werror.ErrUserNotFound
		}
		return nil, werror.ErrUserNotFound
	}

	if u.GetToken() != userAndToken[1] {
		log.Info.Printf("Rejecting authorization for %s due to invalid token", userAndToken[0])
		return nil, werror.ErrBasicAuthFormat
	}

	return u, nil
}

func WebsocketAuth(c *gin.Context, authHeader []string) (*models.User, *models.Instance, error) {
	scheme, cred, err := parseAuthHeader(authHeader)
	if err != nil {
		return nil, nil, err
	}

	s, ok := c.Get("services")
	if !ok {
		panic("No services in websocket auth")
	}
	service := s.(*models.ServicePack)

	var user *models.User
	var instance *models.Instance
	if scheme == "" {
		user = service.UserService.GetPublicUser()
	} else if scheme == "Basic" {
		user, err = validateBasicAuth(cred, service.UserService)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			c.JSON(code, safe)
			return nil, nil, err
		}
	} else if scheme == "Bearer" {
		key, err := service.AccessService.GetApiKey(models.WeblensApiKey(cred))
		if err != nil {
			return nil, nil, err
		}
		if key.RemoteUsing == "" {
			return nil, nil, errors.New("Bad bearer token in websocket auth")
		}
		instance = service.InstanceService.Get(key.RemoteUsing)
		if instance == nil {
			return nil, nil, errors.New("No remote using key in websocket auth")
		}
	} else {
		return nil, nil, werror.ErrBadAuthScheme
	}

	return user, instance, nil
}

func WeblensAuth(requireAdmin bool, pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.RequestsCounter.Inc()
		authHeader := c.Request.Header["Authorization"]
		scheme, cred, err := parseAuthHeader(authHeader)
		if err != nil {
			log.ShowErr(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if scheme == "Bearer" {
			var usr *models.User
			usr, err = pack.AccessService.GetUserFromToken(cred)
			if err != nil {
				log.ShowErr(err)
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			if requireAdmin && (usr == nil || !usr.IsAdmin()) {
				var usrname models.Username = "Public User"
				if usr != nil {
					usrname = usr.GetUsername()
				}
				log.Info.Printf(
					"Rejecting authorization for [%s] due to insufficient permissions on a privileged request",
					usrname,
				)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Set("user", usr)
			c.Next()

			// key, err := pack.AccessService.GetApiKey(models.WeblensApiKey(cred))
			// if err != nil {
			// 	c.AbortWithStatus(http.StatusNotFound)
			// 	return
			// }
			// if key.Key != "" {
			// 	c.Next()
			// } else {
			// 	c.AbortWithStatus(http.StatusUnauthorized)
			// 	return
			// }
		} else if scheme == "Basic" {
			user, err := validateBasicAuth(cred, pack.UserService)
			if err != nil {
				return
			}

			if requireAdmin && !user.IsAdmin() {
				log.Info.Printf(
					"Rejecting authorization for %s due to insufficient permissions on a privileged request",
					user.GetUsername(),
				)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("user", user)
			c.Next()
		} else {
			if requireAdmin {
				log.Warning.Printf("Request at admin endpoint from unauthorized source")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Next()
			return
		}

	}
}

func KeyOnlyAuth(pack *models.ServicePack) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["Authorization"]
		scheme, cred, err := parseAuthHeader(authHeader)
		if err != nil {
			log.ShowErr(err)
			c.Status(http.StatusBadRequest)
			return
		}
		if scheme != "Bearer" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		key, err := pack.AccessService.GetApiKey(models.WeblensApiKey(cred))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if key.Key == "" {
			log.Debug.Println("Failed to find key with this id:", cred)
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		instance := pack.InstanceService.Get(key.RemoteUsing)
		c.Set("instance", instance)
		c.Next()
	}
}

func initSafety(c *gin.Context) {
	ip := net.ParseIP(c.ClientIP())
	if !ip.IsPrivate() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
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
