package comm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var router *gin.Engine
var srv *http.Server

var FileService *service.FileServiceImpl
var MediaService models.MediaService
var AccessService models.AccessService
var UserService models.UserService
var ShareService models.ShareService
var InstanceService models.InstanceService
var AlbumService models.AlbumService
var TaskService task.TaskService
var ClientService models.ClientManager
var Caster models.Broadcaster
var Server *http.Server

var routerLock sync.Mutex

func DoRoutes() {
	routerLock.Lock()
	if srv != nil {
		// Wait for request to finish before shutting down router
		err := srv.Shutdown(context.Background())
		log.ErrTrace(err)
		srv = nil
	}

	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(log.ApiLogger(internal.IsDevMode()))
	router.Use(CORSMiddleware())

	api := router.Group("/api")
	api.Use(WeblensAuth(false))

	admin := router.Group("/api")
	admin.Use(WeblensAuth(true))

	core := router.Group("/api/core")
	core.Use(KeyOnlyAuth)

	if !InstanceService.IsLocalLoaded() {
		addLoadingRoutes(api)
	} else {
		local := InstanceService.GetLocal()
		if local.ServerRole() == models.InitServer {
			api.Use(WeblensAuth(true))
			addInitializationRoutes(api)
			// log.Info.Println("Ignoring requests from public IPs until weblens is initialized")
			// router.Use(initSafety)
		} else {
			addApiRoutes(api)
			addAdminRoutes(admin)
			if local.ServerRole() == models.CoreServer {
				addCoreRoutes(core)
			} else if local.ServerRole() == models.BackupServer {
				addBackupRoutes(api, admin)
			}
		}
	}

	if !internal.DetachUi() {
		addUiRoutes()
	}

	srv = &http.Server{
		Addr:    internal.GetRouterIp() + ":" + internal.GetRouterPort(),
		Handler: router,
	}

	if InstanceService.IsLocalLoaded() {
		Caster.PushWeblensEvent("weblens_loaded")
	}

	log.Debug.Println("Starting router at", srv.Addr)
	routerLock.Unlock()
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
	log.Debug.Println("Restarting router...")
}

var upgrader = gorilla.Upgrader{
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
	api.GET("/media/:mediaId/thumbnail.png", getMediaThumbnailPng)
	api.GET("/media/:mediaId/fullres", getMediaFullres)
	api.GET("/media/:mediaId/stream", streamVideo)
	api.GET("/media/:mediaId/:chunkName", streamVideo)
	api.POST("/media/:mediaId/liked", likeMedia)
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
	api.GET("/files/search", searchByFilename)
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
	admin.POST("/scan", scanDir)
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

	// Get all media
	core.GET("/media", getMediaArchive)
	core.GET("/media/:mediaId/content", fetchMediaBytes)

	core.POST("/files", getFilesMeta)
	core.GET("/file/:fileId", getFileMeta)
	core.GET("/file/:fileId/stat", getFileStat)
	core.GET("/file/:fileId/directory", getDirectoryContent)
	core.GET("/file/:fileId/content", getFileBytes)

	core.GET("/history/since/:timestamp", getLifetimesSince)
	core.GET("/history", getHistory)
	core.GET("/history/folder", getFolderHistory)
	core.GET("/ws", wsConnect)
}

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
