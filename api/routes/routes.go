package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func uiRedirect(ctx *gin.Context) {
	ctx.Redirect(301, "/ui/")
}

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
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

func AddApiRoutes(r *gin.Engine) {
	r.Use(CORSMiddleware())
	api := r.Group("/api")
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

	api.POST("/login", func(ctx *gin.Context) { loginUser(ctx) })
	//api.POST("/item", func(ctx *gin.Context) { uploadItem(ctx) })
	//api.POST("/scan", func(ctx *gin.Context) { scan(ctx) })

	api.GET("/ws", func (ctx *gin.Context) { wsConnect(ctx) })
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