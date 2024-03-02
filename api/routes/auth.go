package routes

import (
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func WeblensAuth(websocket, allowEmptyAuth, requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := dataStore.NewDB()
		var authString string

		if !websocket {
			authHeader := c.Request.Header["Authorization"]
			if len(authHeader) == 0 {
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
		}

		authList := strings.Split(authString, ",")

		if len(authList) < 2 || !db.CheckToken(authList[0], authList[1]) { // {user, token}
			util.Info.Printf("Rejecting authorization for %s due to invalid token", authList[0])
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, _ := db.GetUser(authList[0])
		if requireAdmin && !user.Admin {
			util.Info.Printf("Rejecting authorization for %s due to insufficient permissions on a privileged request", authList[0])
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("username", authList[0])

		c.Next()
	}
}

func CanUserAccessFile(fileId, username, shareId string) bool {
	// There is no such thing as a public, non-shared file. Also, file id is required
	if (username == "" && shareId == "") || fileId == "" {
		return false
	}

	if shareId != "" {
		s, err := dataStore.GetShare(shareId, dataStore.FileShare)
		if err != nil {
			return false
		}

		// Accessing a public share
		if s.IsPublic() && s.GetContentId() == fileId {
			return true
		}
	}

	// Share is not public, so user must be logged in to access
	if username == "" {
		return false
	}

	if fileId != "" {
		f := dataStore.FsTreeGet(fileId)
		if f == nil {
			return false
		}

		if f.Owner() == username {
			return true
		}
	}

	return false
}
