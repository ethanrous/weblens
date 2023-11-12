package routes

import (
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}

func WeblensAuth(websocket, requireAdmin bool) gin.HandlerFunc {
    return func(c *gin.Context) {
		db := dataStore.NewDB("SYS")
		var authString string

		if !websocket {
			authHeader := c.Request.Header["Authorization"]
			if len(authHeader) == 0 {
				util.Info.Printf("Rejecting authorization for unknown user due to empty auth header")
				c.AbortWithStatus(http.StatusUnauthorized)
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

func AddApiRoutes(r *gin.Engine) {
	r.Use(CORSMiddleware())

	public := r.Group("/api")
	public.POST("/login", loginUser)

	api := r.Group("/api")
	api.Use(WeblensAuth(false, false))

	api.GET("/media", getPagedMedia)
	api.GET("/item/:filehash", getMediaItem)
	api.PUT("/items", updateMediaItems)
	api.GET("/stream/:filehash", streamVideo)

	api.GET("/directory", getDirInfo)
	api.POST("/directory", makeDir)

	api.GET("/file", getFile)
	api.POST("/file", uploadFile)
	api.DELETE("/file", moveFileToTrash)
	api.PUT("/file", updateFile)

	api.GET("/takeout/:takeoutId", getTakeout)
	api.POST("/takeout", createTakeout)

	api.GET("/user", getUserInfo)
	api.POST("/user", createUser)

	admin := r.Group("/api")
	admin.Use(WeblensAuth(false, true))

	admin.GET("/users", getUsers)

	websocket := r.Group("/api")
	websocket.Use(WeblensAuth(true, false))

	websocket.GET("/ws", wsConnect)
}

func AddUiRoutes(r *gin.Engine) {
	r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	r.NoRoute(func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
			ctx.File("../ui/build/index.html")
		}
		//default 404 page not found
	})
	//r.GET("/", uiRedirect)
	//r.StaticFS("/ui/", http.Dir("../ui/build"))
}