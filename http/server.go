package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/docs"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginswag "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

type Server struct {
	Running     bool
	StartupFunc func()

	router     *gin.Engine
	stdServer  *http.Server
	routerLock sync.Mutex
	services   *models.ServicePack
	hostStr    string
}

// @title						Weblens API
// @version					1.0
// @description				Programmatic access to the Weblens server
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
// @host						localhost:8080
// @BasePath					/api/
//
// @securityDefinitions.apikey	SessionAuth
// @in							cookie
// @name						weblens-session-token
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
//
// @scope.admin				Grants read and write access to privileged data
func NewServer(host, port string, services *models.ServicePack) *Server {

	proxyHost := env.GetProxyAddress()
	if strings.HasPrefix(proxyHost, "http") {
		i := strings.Index(proxyHost, "://")
		proxyHost = proxyHost[i+3:]
	}
	docs.SwaggerInfo.Host = proxyHost

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
		s.router.GET("/api/ws", AllowPublic(), WeblensAuth(s.services), wsConnect)
		s.router.GET("/docs/*any", ginswag.WrapHandler(swaggerFiles.Handler))

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

	init := s.router.Group("/api")

	init.POST("/server", initializeServer)

	init.GET("/user", getUserInfo)
	// init.GET("/users", getUsers)
}

func (s *Server) UseApi() {
	log.Trace.Println("Using api routes")

	public := s.router.Group("/api")
	public.Use(AllowPublic())

	api := s.router.Group("/api")
	api.Use(CORSMiddleware())

	api.Use(WeblensAuth(s.services))
	public.Use(WeblensAuth(s.services))

	// Media
	media := api.Group("/media")
	media.GET("", getMediaBatch)
	media.GET("/types", getMediaTypes)
	media.GET("/random", getRandomMedias)
	media.GET("/:mediaId", getMediaImage)
	media.GET("/:mediaId/info", getMediaInfo)
	// media.GET("/:mediaId/thumbnail", getMediaThumbnail)
	// media.GET("/:mediaId/thumbnail.webp", getMediaThumbnail)
	// media.GET("/:mediaId/thumbnail.png", getMediaThumbnailPng)
	// media.GET("/:mediaId/fullres", getMediaFullres)
	media.GET("/:mediaId/stream", streamVideo)
	media.GET("/:mediaId/:chunkName", streamVideo)
	media.POST("/:mediaId/liked", likeMedia)
	// media.POST("s", getMediaByIds)
	media.PATCH("/visibility", hideMedia)
	media.PATCH("/date", adjustMediaDate)

	// Files
	files := api.Group("/files")
	files.GET("/:fileId", getFile)
	files.GET("/:fileId/text", getFileText)
	files.GET("/:fileId/stats", getFileStats)
	files.GET("/:fileId/download", downloadFile)
	files.GET("/:fileId/history", getFolderHistory)
	files.GET("/search", searchByFilename)
	files.GET("/shared", getSharedFiles)

	files.POST("/restore", restoreFiles)

	files.PATCH("/:fileId", updateFile)
	files.PATCH("", moveFiles)
	files.PATCH("/trash", trashFiles)
	files.PATCH("/untrash", unTrashFiles)
	files.DELETE("", deleteFiles)

	// Folder
	folder := api.Group("/folder")
	folder.POST("", createFolder)
	public.GET("/folder/:folderId", getFolder)
	folder.PATCH("/:folderId/cover", setFolderCover)

	// Upload
	api.POST("/upload", newUploadTask)
	api.POST("/upload/:uploadId", newFileUpload)
	api.PUT("/upload/:uploadId/file/:fileId", handleUploadChunk)

	// Takeout
	api.POST("/takeout", createTakeout)

	// Users
	users := api.Group("/users")
	users.GET("", RequireAdmin(), getUsers)
	users.GET("/me", getUserInfo)
	users.GET("/search", searchUsers)
	users.POST("", createUser)

	// Must not use weblens auth here, as the user is not logged in yet
	public.POST("/users/auth", loginUser)

	users.POST("/logout", logoutUser)
	users.PATCH("/:username/password", RequireAdmin(), updateUserPassword)
	users.PATCH("/:username/admin", RequireAdmin(), setUserAdmin)
	users.DELETE("/:username", RequireAdmin(), deleteUser)

	// Share
	share := api.Group("/share")
	share.GET("/:shareId", getFileShare)
	share.POST("/files", createFileShare)
	share.PATCH("/:shareId/accessors", patchShareAccessors)
	share.PATCH("/:shareId/public", setSharePublic)
	share.DELETE("/:shareId", deleteShare)

	// Album
	api.GET("/album/:albumId", getAlbum)
	api.GET("/album/:albumId/preview", albumPreviewMedia)
	api.POST("/album", createAlbum)
	api.POST("/album/:albumId/leave", unshareMeAlbum)
	api.PATCH("/album/:albumId", updateAlbum)
	api.DELETE("/album/:albumId", deleteAlbum)

	// Albums
	api.GET("/albums", getAlbums)

	// ApiKeys
	keys := api.Group("/keys")
	keys.Use(RequireAdmin())
	keys.GET("", getApiKeys)
	keys.POST("", newApiKey)
	keys.DELETE("/:keyId", deleteApiKey)

	/* Static content */
	api.GET("/static/:filename", serveStaticContent)

	s.UseAdmin()
}

func (s *Server) UseWebdav(fileService models.FileService, caster models.FileCaster) {
	// fs := service.WebdavFs{
	// 	WeblensFs: fileService,
	// 	Caster:    caster,
	// }

	// handler := &webdav.Handler{
	// 	FileSystem: fs,
	// 	// FileSystem: webdav.Dir(env.GetDataRoot()),
	// 	LockSystem: webdav.NewMemLS(),
	// 	Logger: func(r *http.Request, err error) {
	// 		if err != nil {
	// 			log.Error.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
	// 		} else {
	// 			log.Info.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
	// 		}
	// 	},
	// }

	// go http.ListenAndServe(":8081", handler)
}

func (s *Server) UseInterserverRoutes() {
	log.Trace.Println("Using interserver routes")

	core := s.router.Group("/api/core")
	core.Use(KeyOnlyAuth(s.services))

	// core.POST("/remote", attachRemote)

	core.POST("/files", getFilesMeta)
	core.GET("/file/:fileId", getFileMeta)
	core.GET("/file/:fileId/stat", getFileStat)
	core.GET("/file/:fileId/directory", getDirectoryContent)
	core.GET("/file/content/:contentId", getFileBytes)

	core.GET("/history/since", getLifetimesSince)
	core.GET("/history/folder", getFolderHistory)

	core.GET("/backup", doFullBackup)
}

func (s *Server) UseRestore() {
	log.Trace.Println("Using restore routes")

	restore := s.router.Group("/api/core/restore")
	restore.Use(WeblensAuth(s.services))

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
	admin.Use(WeblensAuth(s.services))
	admin.Use(RequireAdmin())

	admin.GET("/files/external", getExternalDirs)
	admin.GET("/files/external/:folderId", getExternalFolderInfo)
	admin.GET("/files/autocomplete", autocompletePath)

	admin.PATCH("/user/:username/activate", activateUser)

	admin.POST("/scan", scanDir)
	admin.POST("/cache", clearCache)

	admin.GET("/remotes", getRemotes)
	admin.POST("/remotes", attachRemote)
	admin.DELETE("/remotes", removeRemote)

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
