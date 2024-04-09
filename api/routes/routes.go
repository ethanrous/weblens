package routes

import (
	"net/http"
	"os"
	"runtime/pprof"
	"sync"

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

	public := r.Group("/api/public")
	public.Use(WeblensAuth(true, false))

	r.GET("/ping", ping)

	public.GET("/info", getServerInfo)
	public.GET("/media/types", getMediaTypes)
	public.GET("/media/random", getRandomMedias)
	public.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	public.GET("/file/:fileId", getFile)
	public.GET("/file/share/:shareId", getFileShare)
	public.GET("/share/:shareId", getFileShare)
	public.GET("/download", downloadFile)
	public.GET("/users", publicGetUsers)

	public.PUT("/upload/:uploadId/file/:fileId", handleUploadChunk)

	public.POST("/initialize", initializeServer)

	public.POST("/login", loginUser)
	public.POST("/user", createUser)
	public.POST("/folder", pubMakeDir)
	public.POST("/upload", newSharedUploadTask)
	public.POST("/upload/:uploadId", newFileUpload)
	public.POST("/takeout", createTakeout)

	api := r.Group("/api")
	api.Use(WeblensAuth(false, false))

	api.GET("/user", getUserInfo)
	api.GET("/users", searchUsers)
	api.PATCH("/user/:username/password", updateUserPassword)

	api.GET("/media", getMediaBatch)
	api.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	api.GET("/media/:mediaId/fullres", getMediaFullres)
	api.GET("/media/:mediaId/meta", getMediaMeta)

	api.GET("/file/:fileId/shares", getFilesShares)
	api.PATCH("/file/:fileId", updateFile)
	api.PATCH("/file/share/:shareId", updateFileShare)

	api.GET("/files/:folderId/stats", getFolderStats)
	api.GET("/files/shared", getSharedFiles)
	api.PATCH("/files", updateFiles)
	api.PATCH("/files/trash", trashFiles)
	api.PATCH("/files/untrash", unTrashFiles)
	api.DELETE("/files", deleteFiles)

	api.GET("/folder/:folderId", getFolderInfo)
	api.GET("/folder/:folderId/search", searchFolder)
	api.GET("/folder/:folderId/media", getFolderMedia)
	api.POST("/folder", makeDir)

	api.POST("/upload", newUploadTask)

	api.DELETE("/share/:shareId", deleteShare)
	api.POST("/share/files", createFileShare)

	api.GET("/albums", getAlbums)
	api.GET("/album/:albumId", getAlbum)
	api.POST("/album", createAlbum)
	api.PATCH("/album/:albumId", updateAlbum)
	api.DELETE("/album/:albumId", deleteAlbum)

	admin := r.Group("/api/admin")
	admin.Use(WeblensAuth(false, true))

	admin.GET("/files/external", getExternalDirs)
	admin.GET("/files/external/:folderId", getExternalFolderInfo)

	admin.GET("/users", getUsers)
	admin.POST("/user", activateUser)
	admin.PATCH("/user/:username/admin", setUserAdmin)
	admin.DELETE("/user/:username", deleteUser)

	admin.GET("/apiKeys", getApiKeys)
	admin.POST("/scan", recursiveScanDir)
	admin.POST("/cleanup/medias", cleanupMedias)
	admin.POST("/cache", clearCache)
	admin.POST("/apiKey", newApiKey)

	admin.POST("/backup", getBackupSnapshot)

	websocket := r.Group("/api")
	websocket.GET("/ws", wsConnect)

	keyOnly := r.Group("/api")
	keyOnly.Use(KeyOnlyAuth)

	keyOnly.POST("/remote", attachRemote)
	keyOnly.GET("/snapshot", getBackupSnapshot)
}

func AddUiRoutes(r *gin.Engine) {
	memFs := InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}}
	// indexAbsPath :=
	memFs.loadIndex()
	// r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	serveFunc := static.Serve("/", memFs)
	r.Use(func(ctx *gin.Context) {
		strings.Index(ctx.Request.RequestURI, "/assets")

		if strings.HasPrefix(ctx.Request.RequestURI, "/assets") {
			ctx.Writer.Header().Set("Content-Encoding", "gzip")
		}
		serveFunc(ctx)
	})
	// r.Use(serveUiFile)
	r.NoRoute(func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
			// using the real path here makes gin redicrect to /, which creates an infinite loop
			// ctx.Writer.Header().Set("Content-Encoding", "gzip")
			ctx.FileFromFS("/index", memFs)
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
