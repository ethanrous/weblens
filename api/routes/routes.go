package routes

import (
	"net/http"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
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

func WeblensAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
		db := dataStore.NewDB()

		authHeader := c.Request.Header["Authorization"]
		if len(authHeader) == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authList := strings.Split(authHeader[0], ",")

		if len(authList) < 2 || !db.CheckToken(authList[0], authList[1]) { // {user, token}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

        c.Next()
    }
}

func AddApiRoutes(r *gin.Engine) {
	r.Use(CORSMiddleware())

	public := r.Group("/api")
	public.POST("/login", func(ctx *gin.Context) { loginUser(ctx) })

	api := r.Group("/api")
	api.Use(WeblensAuth())

	api.GET("/media", func(ctx *gin.Context) { getPagedMedia(ctx) })
	api.GET("/item/:filehash", func(ctx *gin.Context) { getMediaItem(ctx) })
	api.GET("/stream/:filehash", func(ctx *gin.Context) { streamVideo(ctx) })

	api.GET("/directory", func(ctx *gin.Context) { getDirInfo(ctx) })
	api.POST("/directory", func(ctx *gin.Context) { makeDir(ctx) })

	api.GET("/file", func(ctx *gin.Context) { getFile(ctx) })
	api.DELETE("/file", func(ctx *gin.Context) { moveFileToTrash(ctx) })
	api.PUT("/file", func(ctx *gin.Context) { updateFile(ctx) })

	api.GET("/takeout/:takeoutId", func(ctx *gin.Context) { getTakeout(ctx) })
	api.POST("/takeout", func(ctx *gin.Context) { createTakeout(ctx) })

	r.GET("/api/ws", func (ctx *gin.Context) { wsConnect(ctx) })
}

func AddUiRoutes(r *gin.Engine) {
	r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	r.NoRoute(func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
			ctx.File("../ui/build/index.html")
		}
		//default 404 page not found
	})
	//r.GET("/", func(ctx *gin.Context) { uiRedirect(ctx) })
	//r.StaticFS("/ui/", http.Dir("../ui/build"))
}