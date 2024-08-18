package routes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
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
		return "", "", ErrBadAuthScheme
	} else if authList[0] != "Basic" && authList[0] != "Bearer" {
		// Bad auth header scheme
		return "", "", ErrBadAuthScheme
	}

	// Pass
	return authList[0], authList[1], nil
}

func validateBasicAuth(cred string) (types.User, error) {
	credB, err := base64.StdEncoding.DecodeString(cred)
	if err != nil {
		wlog.ErrTrace(err)
		return nil, err
	}
	userAndToken := strings.Split(string(credB), ":")

	if len(userAndToken) != 2 {
		return nil, ErrBasicAuthFormat
	}

	u := types.SERV.UserService.Get(types.Username(userAndToken[0]))
	if u == nil {
		if types.SERV.UserService.Size() == 0 {
			// c.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return nil, dataStore.ErrNoUser
		}
		// c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "u not found"})
		return nil, dataStore.ErrNoUser
	}

	if u.GetToken() != userAndToken[1] { // {u, token}
		wlog.Info.Printf("Rejecting authorization for %s due to invalid token", userAndToken[0])
		// c.AbortWithStatus(http.StatusUnauthorized)
		return nil, ErrBasicAuthFormat
	}

	return u, nil
}

func WebsocketAuth(c *gin.Context, authHeader []string) (types.User, types.Instance, error) {
	scheme, cred, err := parseAuthHeader(authHeader)
	if err != nil {
		return nil, nil, err
	}

	var user types.User
	var instance types.Instance
	if scheme == "" {
		user = types.SERV.UserService.GetPublicUser()
	} else if scheme == "Basic" {
		user, err = validateBasicAuth(cred)
		if err != nil {
			var weblensError types.WeblensError
			switch {
			case errors.As(err, &weblensError):
				return nil, nil, err
			default:
				c.AbortWithStatus(http.StatusInternalServerError)
				return nil, nil, err
			}
		}
	} else if scheme == "Bearer" {
		key := types.SERV.AccessService.Get(types.WeblensApiKey(cred))
		if key.RemoteUsing == "" {
			return nil, nil, types.WeblensErrorMsg("Bad bearer token in websocket auth")
		}
		instance = types.SERV.InstanceService.Get(key.RemoteUsing)
		if instance == nil {
			return nil, nil, types.WeblensErrorMsg("No remote using key in websocket auth")
		}
	} else {
		return nil, nil, ErrBadAuthScheme
	}

	return user, instance, nil
}

func WeblensAuth(requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header["Authorization"]
		scheme, cred, err := parseAuthHeader(authHeader)
		if err != nil {
			wlog.ShowErr(err)
			c.Status(http.StatusBadRequest)
			return
		}

		if scheme == "Bearer" {
			if types.SERV.AccessService.Get(types.WeblensApiKey(cred)).Key != "" {
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
				wlog.Info.Printf(
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
		wlog.ShowErr(err)
		c.Status(http.StatusBadRequest)
		return
	}
	if scheme != "Bearer" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if types.SERV.AccessService.Get(types.WeblensApiKey(cred)).Key != "" {
		c.Next()
	} else {
		wlog.Debug.Println("Failed to find key with this id:", cred)
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
	path := c.Request.RequestURI

	handler := runtime.FuncForPC(reflect.ValueOf(c.Handler()).Pointer()).Name()
	handler = handler[strings.LastIndex(handler, ".")+1:]

	util.LabelThread(
		func(_ context.Context) {
			c.Next()
		}, "Req Path", path, "Handler Func", handler,
	)

	status := c.Writer.Status()
	if !util.IsDevMode() && status < 400 {
		return
	}

	timeTotal := time.Since(start)
	remote := c.ClientIP()
	method := c.Request.Method

	fmt.Printf(
		"\u001B[0m[API] %s | %s | %12s | [%s] %s %s %s\n", start.Format("Jan 02 15:04:05"), remote,
		colorTime(timeTotal),
		handler, colorStatus(status), method, path,
	)
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
