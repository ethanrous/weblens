package routes

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func WeblensAuth(websocket, allowEmptyAuth, requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authString string

		if !websocket {
			authHeader := c.Request.Header["Authorization"]
			if len(authHeader) == 0 || len(authHeader[0]) == 0 {
				if !allowEmptyAuth {
					util.Info.Printf("Rejecting authorization for unknown user due to empty auth header")
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				c.Next()
				return
			}
			authString = authHeader[0]
		} else {
			authString = c.Query("Authorization")
			if len(authString) == 0 {
				c.Next()
				return
			}
		}

		authList := strings.Split(authString, " ")
		if len(authList) < 2 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		scheme := authList[0]
		cred := authList[1]

		if scheme == "Bearer" {
			if dataStore.CheckApiKey(authList[1]) {
				c.Next()
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			return
		} else if scheme == "Basic" {
			credB, err := base64.StdEncoding.DecodeString(cred)
			if err != nil {
				util.ErrTrace(err)
				c.AbortWithStatus(http.StatusBadRequest)
			}
			userAndToken := strings.Split(string(credB), ":")

			if len(userAndToken) != 2 {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			user := dataStore.GetUser(types.Username(userAndToken[0]))
			if user == nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}

			if !dataStore.CheckUserToken(user, userAndToken[1]) { // {user, token}
				util.Info.Printf("Rejecting authorization for %s due to invalid token", userAndToken[0])
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			if requireAdmin && !user.IsAdmin() {
				util.Info.Printf("Rejecting authorization for %s due to insufficient permissions on a privileged request", userAndToken[0])
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("username", userAndToken[0])
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
		}

	}
}
