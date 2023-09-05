package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func uiRedirect(ctx *gin.Context) {
	ctx.Redirect(301, "/ui/")
}

func AddApiRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/media", func(ctx *gin.Context) { getPagedMedia(ctx) })
	api.GET("/item/:filehash", func(ctx *gin.Context) { getMediaItem(ctx) })
	api.POST("/item", func(ctx *gin.Context) { uploadItem(ctx) })

	api.POST("/scan", func(ctx *gin.Context) { scan(ctx) })
}

func AddUiRoutes(r *gin.Engine) {
	r.GET("/", func(ctx *gin.Context) { uiRedirect(ctx) })
	r.StaticFS("/ui/", http.Dir("../ui/build"))
}