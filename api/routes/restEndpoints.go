package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
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

	api := router.Group("/api")
	api.Use(WeblensAuth(false, types.SERV.UserService))

	admin := router.Group("/api")
	admin.Use(WeblensAuth(false, types.SERV.UserService))

	core := router.Group("/api/core")
	core.Use(KeyOnlyAuth)

	local := types.SERV.InstanceService.GetLocal()
	if local.ServerRole() == types.Initialization {
		util.Debug.Println("Weblens not initialized, only adding initialization routes...")
		init := router.Group("/api")
		init.Use(WeblensAuth(true, types.SERV.UserService))
		AddInitializationRoutes(init)
		util.Info.Println("Ignoring requests from public IPs until weblens is initialized")
		router.Use(initSafety)
	} else {
		AddSharedRoutes(api)
		if local.ServerRole() == types.Core {
			AddApiRoutes(api)
			AddAdminRoutes(admin)
			AddCoreRoutes(core)
		} else if local.ServerRole() == types.Backup {
			AddBackupRoutes(api, admin)

			addr, err := local.GetCoreAddress()
			util.ShowErr(err)
			router.Use(ReverseProxyToCore(addr))
		}
		if util.IsDevMode() {
			AttachProfiler()
		}
	}

	if !util.DetachUi() {
		AddUiRoutes()
	}

	srv = &http.Server{
		Addr:    util.GetRouterIp() + ":" + util.GetRouterPort(),
		Handler: router,
	}

	util.Debug.Println("Starting router...")
	srvMu.Unlock()
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func AddInitializationRoutes(api *gin.RouterGroup) {
	api.POST("/initialize", initializeServer)
	api.GET("/info", getServerInfo)
	api.GET("/users", getUsers)
	api.GET("/user", getUserInfo)
}

func AddSharedRoutes(api *gin.RouterGroup) {
	router.GET("/ping", ping)
	api.GET("/info", getServerInfo)

	api.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	api.GET("/media/:mediaId/fullres", getMediaFullres)

	api.GET("/file/:fileId/history", getFileHistory)
	api.GET("/history/:folderId", getPastFolderInfo)
	api.POST("/history/restore", restorePastFiles)
}

func AddApiRoutes(api *gin.RouterGroup) {

	/* Public */

	api.GET("/media/types", getMediaTypes)
	api.GET("/media/random", getRandomMedias)

	api.GET("/file/:fileId", getFile)
	api.GET("/file/share/:shareId", getFileShare)
	api.GET("/download", downloadFile)

	api.PUT("/upload/:uploadId/file/:fileId", handleUploadChunk)

	api.POST("/login", loginUser)
	api.POST("/user", createUser)
	api.POST("/upload/:uploadId", newFileUpload)
	api.POST("/takeout", createTakeout)

	/* Api */

	api.GET("/user", getUserInfo)
	api.GET("/users/search", searchUsers)
	api.PATCH("/user/:username/password", updateUserPassword)

	api.GET("/media", getMediaBatch)
	api.PATCH("/media/hide", hideMedia)
	api.PATCH("/media/date", adjustMediaDate)

	api.GET("/file/:fileId/shares", getFilesShares)
	api.PATCH("/file/:fileId", updateFile)
	api.PATCH("/file/share/:shareId", updateFileShare)

	api.GET("/files/:folderId/stats", getFolderStats)
	api.GET("/files/shared", getSharedFiles)
	api.PATCH("/files", moveFiles)
	api.PATCH("/files/trash", trashFiles)
	api.PATCH("/files/untrash", unTrashFiles)
	api.DELETE("/files", deleteFiles)

	api.GET("/folder/:folderId", getFolder)
	api.GET("/folder/:folderId/search", searchFolder)
	api.POST("/folder", createFolder)

	api.GET("/folders/media", getFoldersMedia)

	api.POST("/upload", newUploadTask)

	api.DELETE("/share/:shareId", deleteShare)
	api.POST("/share/files", createFileShare)

	api.GET("/albums", getAlbums)
	api.POST("/album", createAlbum)
	api.GET("/album/:albumId", getAlbum)
	api.GET("/album/:albumId/preview", albumPreviewMedia)
	api.PATCH("/album/:albumId", updateAlbum)
	api.POST("/album/:albumId/leave", unshareMeAlbum)
	api.DELETE("/album/:albumId", deleteAlbum)

	/* Websocket */

	ws := router.Group("/api")
	ws.GET("/ws", wsConnect)
}

func AddAdminRoutes(admin *gin.RouterGroup) {
	admin.GET("/files/external", getExternalDirs)
	admin.GET("/files/external/:folderId", getExternalFolderInfo)

	admin.GET("/users", getUsers)
	admin.PATCH("/user/:username/activate", activateUser)
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
	util.Debug.Println("Proxy-ing not found routes to", coreAddress)
	index := strings.Index(coreAddress, "//")
	coreHost := coreAddress[index+2:]
	return func(c *gin.Context) {
		director := func(req *http.Request) {
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}

			req.URL.Scheme = scheme
			req.URL.Host = coreHost
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
			// using the real path here makes gin redirect to /, which creates an infinite loop
			// ctx.Writer.Header().Set("Content-Encoding", "gzip")
			ctx.FileFromFS("/index", memFs)
		}
	})
}

func snapshotHeap(ctx *gin.Context) {
	file, err := os.Create("heap.out")
	// file, err := os.Create(filepath.Join(util.GetMediaRoot(), "heap.out"))
	util.FailOnError(err, "")
	err = pprof.Lookup("heap").WriteTo(file, 0)
	if err != nil {
		util.ShowErr(err)
	}
}

func AttachProfiler() {
	debug := router.Group("/debug")

	debug.GET("/heap", snapshotHeap)
}
