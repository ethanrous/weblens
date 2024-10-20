package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

// var Server *http.Server

type Server struct {
	Running     bool
	StartupFunc func()

	router     *gin.Engine
	stdServer  *http.Server
	routerLock sync.Mutex
	services   *models.ServicePack
	hostStr    string
}

func NewServer(host, port string, services *models.ServicePack) *Server {
	srv := &Server{
		router:   gin.New(),
		services: services,
		hostStr:  host + ":" + port,
	}

	services.Server = srv

	return srv
}

func (s *Server) Start() {
	for {
		if s.services.StartupChan == nil {
			return
		}

		s.router.Use(withServices(s.services))
		s.router.Use(gin.Recovery())
		s.router.Use(log.ApiLogger(log.GetLogLevel()))

		s.router.GET("/ping", ping)
		s.router.GET("/api/info", getServerInfo)
		s.router.GET("/api/ws", WeblensAuth(false, false, s.services), wsConnect)

		if !env.DetachUi() {
			s.UseUi()
		}

		go s.StartupFunc()
		<-s.services.StartupChan

		s.routerLock.Lock()
		s.stdServer = &http.Server{
			Addr:    s.hostStr,
			Handler: s.router.Handler(),
		}

		s.Running = true
		log.Info.Printf("Starting router at %s", s.hostStr)
		s.routerLock.Unlock()
		err := s.stdServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error.Fatalln(err)
		}
		s.routerLock.Lock()
		s.Running = false
		s.stdServer = nil

		s.router = gin.New()
		s.routerLock.Unlock()
	}
}

func (s *Server) UseInit() {
	log.Debug.Println("Adding initialization routes")

	init := s.router.Group("/api/init")

	init.POST("", initializeServer)
	init.GET("/users", getUsers)
	init.GET("/user", getUserInfo)
}

func (s *Server) UseApi() {
	log.Trace.Println("Using api routes")

	api := s.router.Group("/api")
	api.Use(WeblensAuth(false, false, s.services))

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
	api.POST("/medias", getMediaByIds)
	api.PATCH("/media/visibility", hideMedia)
	api.PATCH("/media/date", adjustMediaDate)

	// File
	api.GET("/file/:fileId", getFile)
	api.GET("/file/:fileId/history", getFolderHistory)
	api.GET("/file/:fileId/text", getFileText)
	api.GET("/file/share/:shareId", getFileShare)
	api.GET("/file/:fileId/download", downloadFile)
	api.PATCH("/file/:fileId", updateFile)

	// Files
	api.GET("/files/:folderId/stats", getFolderStats)
	api.GET("/files/shared", getSharedFiles)
	api.GET("/files/search", searchByFilename)
	api.POST("/files/restore", restoreFiles)
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
	api.PATCH("/folder/:folderId/cover", setFolderCover)

	// Folders
	api.GET("/folders/media", getFoldersMedia)

	// Username
	api.GET("/user", getUserInfo)
	api.GET("/users/search", searchUsers)
	api.POST("/user", createUser)
	api.PATCH("/user/:username/password", updateUserPassword)

	// Share
	api.POST("/share/files", createFileShare)
	api.PATCH("/share/:shareId/accessors", patchShareAccessors)
	api.PATCH("/share/:shareId/public", setSharePublic)
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

	login := s.router.Group("/api").Use(WeblensAuth(false, true, s.services))
	login.POST("/login", loginUser)

	api.POST("/takeout", createTakeout)
	api.GET("/takeout/:fileId", downloadTakeout)

	/* Static content */
	api.GET("/static/:filename", serveStaticContent)

	s.UseAdmin()
}

func (s *Server) UseWebdav(fileService models.FileService, caster models.FileCaster) {
	fs := service.WebdavFs{
		WeblensFs: fileService,
		Caster:    caster,
	}

	handler := &webdav.Handler{
		FileSystem: fs,
		// FileSystem: webdav.Dir(env.GetDataRoot()),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Error.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Info.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}

	go http.ListenAndServe(":8081", handler)
}

func (s *Server) UseInterserverRoutes() {
	log.Trace.Println("Using interserver routes")

	core := s.router.Group("/api/core")
	core.Use(KeyOnlyAuth(s.services))

	// core.POST("/remote", attachRemote)

	core.GET("/media/:mediaId/content", fetchMediaBytes)

	core.POST("/files", getFilesMeta)
	core.GET("/file/:fileId", getFileMeta)
	core.GET("/file/:fileId/stat", getFileStat)
	core.GET("/file/:fileId/directory", getDirectoryContent)
	core.GET("/file/content/:contentId", getFileBytes)

	core.GET("/history/since", getLifetimesSince)
	core.GET("/history/folder", getFolderHistory)

	backup := core.Group("/backup")

	backup.GET("/history", getHistory)

	// Get all users
	backup.GET("/users", getUsersArchive)

	// Get all media
	backup.GET("/media", getMediaArchive)

	// Get all API keys
	backup.GET("/keys", getApiKeysArchive)

	// Get all instances
	backup.GET("/instances", getInstancesArchive)

}

func (s *Server) UseRestore() {
	log.Trace.Println("Using restore routes")

	restore := s.router.Group("/api/core/restore")
	restore.Use(WeblensAuth(false, false, s.services))

	restore.POST("/history", restoreHistory)
	restore.POST("/users", restoreUsers)
	restore.POST("/file", uploadRestoreFile)
	restore.POST("/keys", restoreApiKeys)
	restore.POST("/instances", restoreInstances)
	restore.POST("/complete", finalizeRestore)
}

func (s *Server) UseAdmin() {
	log.Trace.Println("Using admin routes")

	admin := s.router.Group("/api")
	admin.Use(WeblensAuth(true, false, s.services))

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

	admin.POST("/remote", attachRemote)
	admin.DELETE("/remote", removeRemote)

	admin.POST("/backup", launchBackup)
	admin.POST("/restore", restoreToCore)

	if s.services.InstanceService.GetLocal().Role == models.BackupServer {
		admin.POST("/core/attach", attachNewCoreRemote)
	}

	/* DANGER */
	admin.POST("/reset", resetServer)
}

func (s *Server) UseUi() {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}, Pack: s.services}
	memFs.loadIndex()

	serveFunc := static.Serve("/", memFs)
	s.router.Use(
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

	s.router.NoRoute(
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

func (s *Server) Restart() {
	s.services.StartupChan = make(chan bool)
	s.Stop()
}

func (s *Server) Stop() {
	log.Warning.Println("Stopping server", s.services.InstanceService.GetLocal().GetName())
	s.services.Caster.PushWeblensEvent(models.ServerGoingDownEvent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := s.stdServer.Shutdown(ctx)
	log.ErrTrace(err)
	log.ErrTrace(ctx.Err())

	for _, c := range s.services.ClientService.GetAllClients() {
		s.services.ClientService.ClientDisconnect(c)
	}
}
