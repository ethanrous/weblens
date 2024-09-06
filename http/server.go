package http

import (
	"net/http"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

// var Server *http.Server

type Server struct {
	Running    bool
	router     *gin.Engine
	routerLock sync.Mutex
	services   *models.ServicePack
	hostStr    string
}

func NewServer(host, port string, services *models.ServicePack) *Server {
	srv := &Server{
		router:   gin.New(),
		services: services,
		hostStr: host + ":" + port,
	}

	srv.router.Use(withServices(services))
	srv.router.Use(gin.Recovery())
	srv.router.Use(log.ApiLogger(env.GetLogLevel()))
	// srv.router.Use(CORSMiddleware())

	services.Server = srv

	return srv
}

func (s *Server) Start() {
	s.router.GET("/ping", ping)
	s.router.GET("/api/info", getServerInfo)
	s.router.GET("/api/ws", wsConnect)

	if !env.DetachUi() {
		s.UseUi()
	}

	s.Running = true
	defer func() { s.Running = false }()
	log.Info.Printf("Starting router at %s", s.hostStr)
	log.Error.Fatalln(s.router.Run(s.hostStr))
}

func (s *Server) UseInit() {
	log.Info.Println("Adding initialization routes")

	init := s.router.Group("/api/init")

	init.POST("", initializeServer)
	init.GET("/users", getUsers)
	init.GET("/user", getUserInfo)
}

func (s *Server) UseApi() {
	log.Debug.Println("Using api routes")

	api := s.router.Group("/api")
	api.Use(WeblensAuth(false, s.services))

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
	api.GET("/file/:fileId/text", getFileText)
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

	api.POST("/login", loginUser)
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
		// FileSystem: webdav.Dir(env.GetMediaRoot()),
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

func (s *Server) UseCore() {
	log.Trace.Println("Using core routes")

	core := s.router.Group("/api/core")
	core.Use(KeyOnlyAuth(s.services))

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
	// core.GET("/ws", wsConnect)
}

func (s *Server) UseAdmin() {
	log.Debug.Println("Using admin routes")

	admin := s.router.Group("/api")
	admin.Use(WeblensAuth(true, s.services))

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
