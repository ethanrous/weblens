package routes

import (
	"net/http"
	"os"
	"runtime/pprof"

	"strings"

	"github.com/ethrousseau/weblens/api/util"

	// "github.com/gin-contrib/pprof"
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func AddApiRoutes(r *gin.Engine) {
	r.Use(CORSMiddleware())

	public := r.Group("/api")
	public.Use(WeblensAuth(false, true, false))
	public.POST("/login", loginUser)
	public.POST("/user", createUser)

	api := r.Group("/api")
	api.Use(WeblensAuth(false, false, false))

	api.GET("/media", getMediaBatch)
	api.GET("/media/:mediaId", getOneMedia)
	api.PUT("/media", updateMedias)
	// api.GET("/stream/:mediaId", streamVideo)

	api.GET("/folder/:folderId", getFolderInfo)
	api.POST("/folder", makeDir)

	api.GET("/folder/:folderId/media", getFolderMedia)

	// Allow publically creating folders
	public.POST("/public/folder", pubMakeDir)

	api.GET("/trash", getUserTrashInfo)

	// Regular file upload endpoint
	api.POST("/upload", newFileUpload)

	// Allow publically creating file uploads for wormholes
	public.POST("/public/upload", newSharedFileUpload)

	// Allow public chunk upload to support wormhole drops
	public.PUT("/upload/:uploadId", handleUploadChunk)

	api.GET("/file/:fileId", getFile)
	api.PATCH("/file", updateFile)
	api.DELETE("/files", trashFiles)

	api.GET("/file/:fileId/shares", getFileShare)
	api.PATCH("/file/share/:shareId", updateFileShare)

	api.GET("/share", getSharedFiles)
	api.PATCH("/files", updateFiles)
	api.DELETE("/share", deleteShare)
	public.GET("/share/:shareId", getShare)

	// api.GET("/share/files", getFileShare)
	api.POST("/share/files", createFileShare)

	public.GET("/download", downloadFile)

	api.POST("/takeout", createTakeout)

	api.GET("/user", getUserInfo)
	api.GET("/users", searchUsers)

	api.GET("/albums", getAlbums)

	api.GET("/album/:albumId", getAlbum)
	api.POST("/album", createAlbum)
	api.PATCH("/album/:albumId", updateAlbum)
	api.DELETE("/album/:albumId", deleteAlbum)

	admin := r.Group("/api/admin")
	admin.Use(WeblensAuth(false, false, true))

	public.GET("/fileTree", getFileTreeInfo)

	admin.GET("/users", getUsers)
	admin.POST("/user", updateUser)
	admin.DELETE("/user/:username", deleteUser)
	admin.POST("/cleanup/medias", cleanupMedias)
	admin.POST("/cache", clearCache)

	websocket := r.Group("/api")
	websocket.Use(WeblensAuth(true, false, false))

	websocket.GET("/ws", wsConnect)
}

func AddUiRoutes(r *gin.Engine) {
	r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	r.NoRoute(func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
			ctx.File("../ui/build/index.html")
		}
	})
}

func snapshotHeap(ctx *gin.Context) {
	file, err := os.Create("heap.out")
	// file, err := os.Create(filepath.Join(util.GetMediaRoot(), "heap.out"))
	util.FailOnError(err, "")
	pprof.Lookup("heap").WriteTo(file, 0)
}

func AttachProfiler(r *gin.Engine) {
	debug := r.Group("/debug")

	debug.GET("/heap", snapshotHeap)
}
