package comm

import (
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/dataStore"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
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

func validateBasicAuth(cred string) (*models.User, error) {
	credB, err := base64.StdEncoding.DecodeString(cred)
	if err != nil {
		log.ErrTrace(err)
		return nil, err
	}
	userAndToken := strings.Split(string(credB), ":")

	if len(userAndToken) != 2 {
		return nil, werror.ErrBasicAuthFormat
	}

	u := UserService.Get(models.Username(userAndToken[0]))
	if u == nil {
		if UserService.Size() == 0 {
			// c.JSON(comm.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return nil, dataStore.ErrNoUser
		}
		// c.AbortWithStatusJSON(comm.StatusNotFound, gin.H{"error": "u not found"})
		return nil, dataStore.ErrNoUser
	}

	if u.GetToken() != userAndToken[1] { // {u, token}
		log.Info.Printf("Rejecting authorization for %s due to invalid token", userAndToken[0])
		// c.AbortWithStatus(comm.StatusUnauthorized)
		return nil, werror.ErrBasicAuthFormat
	}

	return u, nil
}

func WebsocketAuth(c *gin.Context, authHeader []string) (*models.User, *models.Instance, error) {
	scheme, cred, err := parseAuthHeader(authHeader)
	if err != nil {
		return nil, nil, err
	}

	var user *models.User
	var instance *models.Instance
	if scheme == "" {
		user = UserService.GetPublicUser()
	} else if scheme == "Basic" {
		user, err = validateBasicAuth(cred)
		if err != nil {
			var weblensError error
			switch {
			case errors.As(err, &weblensError):
				return nil, nil, err
			default:
				c.AbortWithStatus(http.StatusInternalServerError)
				return nil, nil, err
			}
		}
	} else if scheme == "Bearer" {
		key, err := AccessService.GetApiKeyById(models.WeblensApiKey(cred))
		if err != nil {
			return nil, nil, err
		}
		if key.RemoteUsing == "" {
			return nil, nil, werror.New("Bad bearer token in websocket auth")
		}
		instance = InstanceService.Get(key.RemoteUsing)
		if instance == nil {
			return nil, nil, werror.New("No remote using key in websocket auth")
		}
	} else {
		return nil, nil, werror.ErrBadAuthScheme
	}

	return user, instance, nil
}

func WeblensAuth(requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["Authorization"]
		scheme, cred, err := parseAuthHeader(authHeader)
		if err != nil {
			log.ShowErr(err)
			c.Status(http.StatusBadRequest)
			return
		}

		if scheme == "Bearer" {
			key, err := AccessService.GetApiKeyById(models.WeblensApiKey(cred))
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			if key.Key != "" {
				c.Next()
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		} else if scheme == "Basic" {
			user, err := validateBasicAuth(cred)
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
			c.Next()
			return
		}

	}
}

func KeyOnlyAuth(c *gin.Context) {
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

	key, err := AccessService.GetApiKeyById(models.WeblensApiKey(cred))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if key.Key != "" {
		c.Next()
	} else {
		log.Debug.Println("Failed to find key with this id:", cred)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func initSafety(c *gin.Context) {
	ip := net.ParseIP(c.ClientIP())
	if !ip.IsPrivate() {
		c.Status(http.StatusNotFound)
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
