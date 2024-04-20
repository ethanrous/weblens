package routes

import (
	"context"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"

	// "github.com/gin-contrib/pprof"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var router *gin.Engine
var srv *http.Server
var srvMu = &sync.Mutex{}

func DoRoutes() {
	srvMu.Lock()
	if srv != nil {
		defer srvMu.Unlock()

		// Wait for request to finish before shutting down router
		time.Sleep(time.Millisecond * 100)
		err := srv.Shutdown(context.TODO())
		util.ErrTrace(err)
		srv = nil
		return
	}
	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(WeblensLogger)
	router.Use(CORSMiddleware())

	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil {
		util.Debug.Println("Weblens not initialized, only adding initialization routes...")
		AddInitializationRoutes()
		util.Info.Println("Ignoring requests from public IPs until weblens is initialized")
		router.Use(initSafety)
	} else {
		api := router.Group("/api")
		api.Use(WeblensAuth(false, false))

		public := router.Group("/api/public")
		public.Use(WeblensAuth(true, false))

		admin := router.Group("/api")
		admin.Use(WeblensAuth(false, true))

		core := router.Group("/api/core")
		core.Use(KeyOnlyAuth)

		AddSharedRoutes(api, public)
		if srvInfo.IsCore() {
			AddApiRoutes(api, public)
			AddAdminRoutes(admin)
			AddCoreRoutes(core)
		} else {
			AddBackupRoutes(api, admin)

			addr, err := srvInfo.GetCoreAddress()
			util.ShowErr(err)
			router.Use(ReverseProxyToCore(addr))
		}
		if !util.DetachUi() {
			AddUiRoutes()
		}
		if util.IsDevMode() {
			AttachProfiler()
		}
	}

	srv = &http.Server{
		Addr:    util.GetRouterIp() + ":" + util.GetRouterPort(),
		Handler: router,
	}

	util.Debug.Println("Starting router...")
	srvMu.Unlock()
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		util.ShowErr(err)
	}
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Content-Range")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func AddInitializationRoutes() {
	api := router.Group("/api")

	api.POST("/initialize", initializeServer)
	api.GET("/public/info", getServerInfo)
	api.GET("/users", getUsers)
	api.GET("/user", getUserInfo)
}

func AddSharedRoutes(api, public *gin.RouterGroup) {
	router.GET("/ping", ping)
	public.GET("/info", getServerInfo)

	api.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	api.GET("/media/:mediaId/fullres", getMediaFullres)

	api.GET("/file/:fileId/history", getFileHistory)
	api.GET("/history/:folderId", getPastFolderInfo)
	api.POST("/history/restore", restorePastFiles)
}

func AddApiRoutes(api, public *gin.RouterGroup) {
	public.GET("/media/types", getMediaTypes)
	public.GET("/media/random", getRandomMedias)
	public.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	public.GET("/file/:fileId", getFile)
	public.GET("/file/share/:shareId", getFileShare)
	public.GET("/share/:shareId", getFileShare)
	public.GET("/download", downloadFile)

	public.PUT("/upload/:uploadId/file/:fileId", handleUploadChunk)

	public.POST("/login", loginUser)
	public.POST("/user", createUser)
	public.POST("/folder", pubMakeDir)
	public.POST("/upload", newSharedUploadTask)
	public.POST("/upload/:uploadId", newFileUpload)
	public.POST("/takeout", createTakeout)

	api.GET("/user", getUserInfo)
	api.GET("/users/search", searchUsers)
	api.PATCH("/user/:username/password", updateUserPassword)

	api.GET("/media", getMediaBatch)

	// api.GET("/media/:mediaId/meta", getMediaMeta)

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

	websocket := router.Group("/api")
	websocket.GET("/ws", wsConnect)
}

func AddAdminRoutes(admin *gin.RouterGroup) {
	admin.GET("/files/external", getExternalDirs)
	admin.GET("/files/external/:folderId", getExternalFolderInfo)

	admin.GET("/users", getUsers)
	admin.POST("/user", activateUser)
	admin.PATCH("/user/:username/admin", setUserAdmin)
	admin.DELETE("/user/:username", deleteUser)

	admin.GET("/apiKeys", getApiKeys)
	admin.GET("/remotes", getRemotes)
	admin.POST("/scan", recursiveScanDir)
	admin.POST("/cache", clearCache)
	admin.POST("/apiKey", newApiKey)
	admin.DELETE("/apiKey", deleteApiKey)
	admin.DELETE("/remote", removeRemote)
}

func AddBackupRoutes(api, admin *gin.RouterGroup) {
	api.GET("/snapshots", getSnapshots)

	admin.GET("/remotes", getRemotes)
	admin.POST("/backup", launchBackup)
}

func AddCoreRoutes(core *gin.RouterGroup) {
	core.POST("/remote", attachRemote)

	// Get all users
	core.GET("/users", getUsersArchive)
	core.GET("/snapshot", getBackupSnapshot)

	core.GET("/files", getFilesMeta)
	core.GET("/file/:fileId/content", getFileBytes)
}

func ReverseProxyToCore(coreAddress string) gin.HandlerFunc {
	util.Debug.Println("Proxying not found routes to", coreAddress)
	index := strings.Index(coreAddress, "//")
	coreHost := coreAddress[index+2:]
	return func(c *gin.Context) {
		director := func(req *http.Request) {
			// r := c.Request
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}

			req.URL.Scheme = scheme
			req.URL.Host = coreHost

			// req.Header["my-header"] = []string{r.Header.Get("my-header")}
			// Golang camelcases headers
			// delete(req.Header, "My-Header")
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func AddUiRoutes() {
	memFs := InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}}
	// indexAbsPath :=
	memFs.loadIndex()
	// r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	serveFunc := static.Serve("/", memFs)
	router.Use(func(ctx *gin.Context) {
		strings.Index(ctx.Request.RequestURI, "/assets")

		if strings.HasPrefix(ctx.Request.RequestURI, "/assets") {
			ctx.Writer.Header().Set("Content-Encoding", "gzip")
		}
		serveFunc(ctx)
	})
	// r.Use(serveUiFile)
	router.NoRoute(func(ctx *gin.Context) {
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

func AttachProfiler() {
	debug := router.Group("/debug")

	debug.GET("/heap", snapshotHeap)
}
