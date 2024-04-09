package routes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func parseAuthHeader(authHeaderParts []string, allowEmptyAuth bool) (string, string, error) {

	if len(authHeaderParts) == 0 || len(authHeaderParts[0]) == 0 {
		if !allowEmptyAuth {
			// Empty auth header on non-public endpoint. Not allowed
			return "", "", ErrEmptyAuth
		}
		// Public login
		return "", "", nil
	}
	authString := authHeaderParts[0]
	authList := strings.Split(authString, " ")
	if len(authList) < 2 {
		// Bad auth header format
		return "", "", ErrBadAuthScheme
	} else if authList[0] != "Basic" && authList[0] != "Bearer" {
		// Bad auth header scheme
		return "", "", ErrBadAuthScheme
	}

	// Pass
	return authList[0], authList[1], nil
}

func validateBasicAuth(c *gin.Context, cred string) (types.User, error) {
	credB, err := base64.StdEncoding.DecodeString(cred)
	if err != nil {
		util.ErrTrace(err)
		c.AbortWithStatus(http.StatusBadRequest)
	}
	userAndToken := strings.Split(string(credB), ":")

	if len(userAndToken) != 2 {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, ErrBasicAuthFormat
	}

	user := dataStore.GetUser(types.Username(userAndToken[0]))
	if user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return nil, dataStore.ErrNoUser
	}

	if !dataStore.CheckUserToken(user, userAndToken[1]) { // {user, token}
		util.Info.Printf("Rejecting authorization for %s due to invalid token", userAndToken[0])
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, ErrBasicAuthFormat
	}

	return user, nil
}

func WebsocketAuth(c *gin.Context, authHeader []string) (types.User, error) {
	scheme, cred, err := parseAuthHeader(authHeader, true)
	if err != nil {
		return nil, err
	}

	if scheme != "Basic" {
		return nil, ErrBadAuthScheme
	}

	user, err := validateBasicAuth(c, cred)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func WeblensAuth(allowEmptyAuth, requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["Authorization"]
		scheme, cred, err := parseAuthHeader(authHeader, allowEmptyAuth)
		if err != nil {
			util.ShowErr(err)
			c.Status(http.StatusBadRequest)
			return
		}

		if scheme == "Bearer" {
			if dataStore.CheckApiKey(types.WeblensApiKey(cred)) {
				c.Next()
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		} else if scheme == "Basic" {
			user, err := validateBasicAuth(c, cred)
			if err != nil {
				return
			}

			if requireAdmin && !user.IsAdmin() {
				util.Info.Printf("Rejecting authorization for %s due to insufficient permissions on a privileged request", user.GetUsername())
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("user", user)
			c.Next()
		} else {
			return
		}

	}
}

func KeyOnlyAuth(c *gin.Context) {
	authHeader := c.Request.Header["Authorization"]
	scheme, cred, err := parseAuthHeader(authHeader, false)
	if err != nil {
		util.ShowErr(err)
		c.Status(http.StatusBadRequest)
		return
	}
	if scheme != "Bearer" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if dataStore.CheckApiKey(types.WeblensApiKey(cred)) {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func colorStatus(status int) string {
	if status < 400 {
		return fmt.Sprintf("\u001b[32m%d\u001B[0m", status)
	} else if status >= 400 && status < 500 {
		return fmt.Sprintf("\u001b[33m%d\u001B[0m", status)
	} else if status >= 500 {
		return fmt.Sprintf("\u001b[31m%d\u001B[0m", status)
	}
	return "Not reached"
}

func colorTime(dur time.Duration) string {
	if dur < time.Millisecond*200 {
		return dur.String()
	} else if dur < time.Second {
		return fmt.Sprintf("\u001b[33m%s\u001B[0m", dur.String())
	} else {
		return fmt.Sprintf("\u001b[31m%s\u001B[0m", dur.String())
	}
}

func WeblensLogger(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	// raw := c.Request.URL.RawQuery

	c.Next()

	timeTotal := time.Since(start)
	remote := c.ClientIP()
	status := c.Writer.Status()
	method := c.Request.Method

	fmt.Printf("\u001B[0m[API] %s | %s | %12s | %s %s %s\n", start.Format("Jan 02 15:04:05"), remote, colorTime(timeTotal), colorStatus(status), method, path)
}
