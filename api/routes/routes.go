package routes

import "github.com/gin-gonic/gin"

func addRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/photos", func(ctx *gin.Context) { listPhotos(ctx) })
	api.GET("/thumbnail", func(ctx *gin.Context) { getPhotoThumb(ctx) })
	api.POST("/upload", func(ctx *gin.Context) { uploadPhoto(ctx) })
	api.POST("/scan", func(ctx *gin.Context) { scan(ctx) })
}