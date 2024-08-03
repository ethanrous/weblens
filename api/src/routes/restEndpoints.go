package routes

import (
	"context"
	"errors"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var router *gin.Engine
var srv *http.Server

func DoRoutes() {
	for types.SERV == nil || !types.SERV.MinimallyReady() {
		time.Sleep(100 * time.Millisecond)
	}

	for {
		types.SERV.RouterLock.Lock()

		if srv != nil {

			// Wait for request to finish before shutting down router
			time.Sleep(time.Millisecond * 100)
			err := srv.Shutdown(context.TODO())
			wlog.ErrTrace(err)
			srv = nil
			types.SERV.RouterLock.Unlock()
			continue
		}

		router = gin.New()
		router.Use(gin.Recovery())
		router.Use(WeblensLogger)
		router.Use(CORSMiddleware())

		api := router.Group("/api")
		api.Use(WeblensAuth(false))

		admin := router.Group("/api")
		admin.Use(WeblensAuth(true))

		core := router.Group("/api/core")
		core.Use(KeyOnlyAuth)

		if !types.SERV.InstanceService.IsLocalLoaded() {
			addLoadingRoutes(api)
		} else {
			local := types.SERV.InstanceService.GetLocal()
			if local.ServerRole() == types.Initialization {
				api.Use(WeblensAuth(true))
				addInitializationRoutes(api)
				wlog.Info.Println("Ignoring requests from public IPs until weblens is initialized")
				router.Use(initSafety)
			} else {
				addApiRoutes(api)
				addAdminRoutes(admin)
				if local.ServerRole() == types.Core {
					addCoreRoutes(core)
				} else if local.ServerRole() == types.Backup {
					addBackupRoutes(api, admin)
				}
				if util.IsDevMode() {
					attachProfiler()
				}
			}
		}

		if !util.DetachUi() {
			addUiRoutes()
		}

		srv = &http.Server{
			Addr:    util.GetRouterIp() + ":" + util.GetRouterPort(),
			Handler: router,
		}

		types.SERV.SetRouter(srv)
		types.SERV.RouterLock.Unlock()

		wlog.Debug.Println("Starting router at", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		wlog.Debug.Println("Restarting router...")
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func addLoadingRoutes(api *gin.RouterGroup) {
	api.GET("/info", getServerInfo)
	api.GET("/ws", wsConnect)
}

func addInitializationRoutes(api *gin.RouterGroup) {
	api.POST("/initialize", initializeServer)
	api.GET("/info", getServerInfo)
	api.GET("/users", getUsers)
	api.GET("/user", getUserInfo)

	ws := router.Group("/api")
	ws.GET("/ws", wsConnect)
}

func addApiRoutes(api *gin.RouterGroup) {
	router.GET("/ping", ping)

	api.GET("/info", getServerInfo)

	api.GET("/file/:fileId/history", getFolderHistory)
	api.GET("/file/rewind/:folderId/:rewindTime", getPastFolderInfo)
	api.POST("/history/restore", restorePastFiles)

	// Media
	api.GET("/media", getMediaBatch)
	api.GET("/media/types", getMediaTypes)
	api.GET("/media/random", getRandomMedias)
	api.GET("/media/:mediaId/info", getMediaInfo)
	api.GET("/media/:mediaId/thumbnail", getMediaThumbnail)
	api.GET("/media/:mediaId/thumbnail.webp", getMediaThumbnail)
	api.GET("/media/:mediaId/fullres", getMediaFullres)
	api.GET("/media/:mediaId/stream", streamVideo)
	api.GET("/media/:mediaId/:chunkName", streamVideo)
	api.PATCH("/media/hide", hideMedia)
	api.PATCH("/media/date", adjustMediaDate)

	// File
	api.GET("/file/:fileId", getFile)
	api.GET("/file/share/:shareId", getFileShare)
	api.GET("/file/:fileId/download", downloadFile)
	api.PATCH("/file/:fileId", updateFile)

	// Files
	api.GET("/files/:folderId/stats", getFolderStats)
	api.GET("/files/shared", getSharedFiles)
	api.PATCH("/files", moveFiles)
	api.PATCH("/files/trash", trashFiles)
	api.PATCH("/files/untrash", unTrashFiles)
	api.DELETE("/files", deleteFiles)

	// Upload
	api.POST("/upload", newUploadTask)
	api.POST("/upload/:uploadId", newFileUpload)
	api.PUT("/upload/:uploadId/file/:fileId", handleUploadChunk)

	// Folder
	api.GET("/folder/:folderId", getFolder)
	api.GET("/folder/:folderId/search", searchFolder)
	api.POST("/folder", createFolder)

	// Folders
	api.GET("/folders/media", getFoldersMedia)

	// User
	api.GET("/user", getUserInfo)
	api.GET("/users/search", searchUsers)
	api.POST("/user", createUser)
	api.PATCH("/user/:username/password", updateUserPassword)

	// ShareId
	api.POST("/share/files", createFileShare)
	api.PATCH("/share/:shareId/accessors", addUserToFileShare)
	api.DELETE("/share/:shareId", deleteShare)

	// Album
	api.GET("/album/:albumId", getAlbum)
	api.GET("/album/:albumId/preview", albumPreviewMedia)
	api.POST("/album", createAlbum)
	api.POST("/album/:albumId/leave", unshareMeAlbum)
	api.PATCH("/album/:albumId", updateAlbum)
	api.DELETE("/album/:albumId", deleteAlbum)

	// Albums
	api.GET("/albums", getAlbums)

	api.POST("/login", loginUser)
	api.POST("/takeout", createTakeout)

	/* Websocket */
	ws := router.Group("/api")
	ws.GET("/ws", wsConnect)

	/* Static content */
	api.GET("/static/:filename", serveStaticContent)
}

func addAdminRoutes(admin *gin.RouterGroup) {
	admin.GET("/files/external", getExternalDirs)
	admin.GET("/files/external/:folderId", getExternalFolderInfo)
	admin.GET("/files/autocomplete", autocompletePath)
	admin.GET("/file/path", getFileDataFromPath)

	admin.GET("/users", getUsers)
	admin.PATCH("/user/:username/activate", activateUser)
	admin.PATCH("/user/:username/admin", setUserAdmin)
	admin.DELETE("/user/:username", deleteUser)

	admin.GET("/keys", getApiKeys)
	admin.GET("/remotes", getRemotes)
	admin.POST("/scan", recursiveScanDir)
	admin.POST("/cache", clearCache)
	admin.POST("/key", newApiKey)
	admin.DELETE("/key/:keyId", deleteApiKey)
	admin.DELETE("/remote", removeRemote)
}

func addBackupRoutes(api, admin *gin.RouterGroup) {
	api.GET("/snapshots", getSnapshots)
	admin.POST("/backup", launchBackup)
}

func addCoreRoutes(core *gin.RouterGroup) {
	core.POST("/remote", attachRemote)

	// Get all users
	core.GET("/users", getUsersArchive)
	core.GET("/media", getMediaArchive)
	core.GET("/media/:mediaId/content", fetchMediaBytes)

	core.GET("/files", getFilesMeta)
	core.GET("/file/:fileId", getFileMeta)
	core.GET("/file/:fileId/stat", getFileStat)
	core.GET("/file/:fileId/directory", getDirectoryContent)
	core.GET("/file/:fileId/content", getFileBytes)

	core.GET("/history/since/:timestamp", getLifetimesSince)
	core.GET("/history", getHistory)
	core.GET("/history/folder", getFolderHistory)
	core.GET("/ws", wsConnect)
}

// func reverseProxyToCore(coreAddress string) gin.HandlerFunc {
// 	util.Debug.Println("Proxy-ing not found routes to", coreAddress)
// 	index := strings.Index(coreAddress, "//")
// 	coreHost := coreAddress[index+2:]
// 	return func(c *gin.Context) {
// 		director := func(req *http.Request) {
// 			scheme := "http"
// 			if c.Request.TLS != nil {
// 				scheme = "https"
// 			}
//
// 			req.URL.Scheme = scheme
// 			req.URL.Host = coreHost
// 		}
// 		proxy := &httputil.ReverseProxy{Director: director, ErrorLog: util.Error}
// 		proxy.ServeHTTP(c.Writer, c.Request)
// 	}
// }

func addUiRoutes() {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}}
	// indexAbsPath :=
	memFs.loadIndex()
	// r.Use(static.Serve("/", static.LocalFile("../ui/build", true)))
	serveFunc := static.Serve("/", memFs)
	router.Use(
		func(ctx *gin.Context) {
			if strings.HasPrefix(ctx.Request.RequestURI, "/api") {
				ctx.Status(http.StatusNotFound)
				return
			}
			strings.Index(ctx.Request.RequestURI, "/assets")

			if strings.HasPrefix(ctx.Request.RequestURI, "/assets") {
				ctx.Writer.Header().Set("Content-Encoding", "gzip")
			}
			serveFunc(ctx)
		},
	)
	// r.Use(serveUiFile)
	router.NoRoute(
		func(ctx *gin.Context) {
			if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
				// using the real path here makes gin redirect to /, which creates an infinite loop
				// ctx.Writer.Header().Set("Content-Encoding", "gzip")
				ctx.FileFromFS("/index", memFs)
			} else {
				ctx.Status(http.StatusNotFound)
				return
			}
		},
	)
}

func snapshotHeap(ctx *gin.Context) {
	file, err := os.Create("heap.out")
	// file, err := os.Create(filepath.Join(util.GetMediaRoot(), "heap.out"))
	util.FailOnError(err, "")
	err = pprof.Lookup("heap").WriteTo(file, 0)
	if err != nil {
		wlog.ShowErr(err)
	}
}

func attachProfiler() {
	debug := router.Group("/debug")

	debug.GET("/heap", snapshotHeap)
}
