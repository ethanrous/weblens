package http

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/task"
	"github.com/gin-gonic/gin"
)

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	upInfo, err := readCtxBody[rest.NewUploadBody](ctx)
	if err != nil {
		return
	}

	// c := models.NewBufferedCaster (pack.ClientService)
	meta := models.UploadFilesMeta{
		ChunkStream:  make(chan models.FileChunk, 10),
		RootFolderId: upInfo.RootFolderId,
		ChunkSize:    upInfo.ChunkSize,
		TotalSize:    upInfo.TotalUploadSize,
		FileService:  pack.FileService,
		MediaService: pack.MediaService,
		TaskService:  pack.TaskService,
		TaskSubber:   pack.ClientService,
		User:         u,
		Caster:       pack.Caster,
	}
	t, err := pack.TaskService.DispatchJob(models.UploadFilesTask, meta, nil)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newFileUpload(ctx *gin.Context) {
	uploadTaskId := ctx.Param("uploadId")
	newFInfo, err := readCtxBody[rest.NewFileBody](ctx)
	if err != nil {
		return
	}

	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	uTask := pack.TaskService.GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Status(http.StatusNotFound)
		return
	}

	parent, err := pack.FileService.GetFileSafe(newFInfo.ParentFolderId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	child, _ := parent.GetChild(newFInfo.NewFileName)
	if child != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "File with the same name already exists in folder"})
		return
	}

	var newFId fileTree.FileId
	err = uTask.Manipulate(
		func(meta task.TaskMetadata) error {
			uploadMeta := meta.(models.UploadFilesMeta)

			newF, err := pack.FileService.CreateFile(parent, newFInfo.NewFileName, uploadMeta.UploadEvent)
			if err != nil {
				return err
			}

			newFId = newF.ID()

			uploadMeta.ChunkStream <- models.FileChunk{
				NewFile: newF, ContentRange: "0-0/" + strconv.FormatInt(newFInfo.FileSize, 10),
			}

			return nil
		},
	)

	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newFId})
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	pack := getServices(ctx)
	uploadId := ctx.Param("uploadId")

	t := pack.TaskService.GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := ctx.Param("fileId")

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := internal.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
	if err != nil {
		log.ShowErr(err)
		// err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		// if err != nil {
		// 	util.ShowErr(err)
		// }
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	err = t.Manipulate(
		func(meta task.TaskMetadata) error {
			chunkData := models.FileChunk{FileId: fileId, Chunk: chunk, ContentRange: ctx.GetHeader("Content-Range")}
			meta.(models.UploadFilesMeta).ChunkStream <- chunkData

			return nil
		},
	)

	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	pack := getServices(ctx)
	err := pack.MediaService.NukeCache()
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)

}

func createFileShare(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	shareInfo, err := readCtxBody[rest.NewShareBody](ctx)
	if err != nil {
		return
	}

	f, err := pack.FileService.GetFileSafe(shareInfo.FileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	_, err = pack.ShareService.GetFileShare(f)
	if err == nil {
		ctx.Status(http.StatusConflict)
		return
	} else if !errors.Is(err, werror.ErrNoShare) {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	accessors := internal.Map(
		shareInfo.Users, func(un models.Username) *models.User {
			return pack.UserService.Get(un)
		},
	)
	newShare := models.NewFileShare(f, u, accessors, shareInfo.Public, shareInfo.Wormhole)

	err = pack.ShareService.Add(newShare)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, newShare)
}

func deleteShare(ctx *gin.Context) {
	pack := getServices(ctx)
	shareId := models.ShareId(ctx.Param("shareId"))

	s := pack.ShareService.Get(shareId)
	if s == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	err := pack.ShareService.Del(s.ID())
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func patchShareAccessors(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(user, share) {
		ctx.Status(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[rest.UserListBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var addUsers []*models.User
	for _, un := range ub.AddUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find user with name " + un})
			return
		}
		addUsers = append(addUsers, u)
	}

	var removeUsers []*models.User
	for _, un := range ub.RemoveUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find user with name " + un})
			return
		}
		removeUsers = append(removeUsers, u)
	}

	if len(addUsers) > 0 {
		err = pack.ShareService.AddUsers(share, addUsers)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	if len(removeUsers) > 0 {
		err = pack.ShareService.RemoveUsers(share, removeUsers)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func setSharePublic(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(user, share) {
		ctx.Status(http.StatusNotFound)
		return
	}

	pub, err := readCtxBody[rest.SharePublicityBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = pack.ShareService.SetSharePublic(share, pub.Public)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}



func getRandomMedias(ctx *gin.Context) {
	ctx.Status(http.StatusNotImplemented)
	return
	// numStr := ctx.Query("count")
	// numPhotos, err := strconv.Atoi(numStr)
	// if err != nil {
	// 	ctx.Status(comm.StatusBadRequest)
	// 	return
	// }

	// media := media.GetRandomMedia(numPhotos)
	// ctx.JSON(comm.StatusOK, gin.H{"medias": media})
}

func initializeServer(ctx *gin.Context) {
	pack := getServices(ctx)
	// Can't init server if already initialized
	role := pack.InstanceService.GetLocal().GetRole()
	if role != models.InitServer {
		ctx.Status(http.StatusNotFound)
		return
	}

	initBody, err := readCtxBody[rest.InitServerBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if initBody.Role == models.CoreServer {
		if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}

		err = pack.InstanceService.InitCore(initBody.Name)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}

		users, err := pack.UserService.GetAll()
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		for u := range users {
			err = pack.UserService.Del(u.GetUsername())
			if err != nil {
				log.ShowErr(err)
			}
		}

		owner, err := pack.UserService.CreateOwner(initBody.Username, initBody.Password)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		token, expires, err := pack.AccessService.GenerateJwtToken(owner)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		cookie := fmt.Sprintf("%s=%s; expires=%s;", SessionTokenCookie, token, expires.Format(time.RFC1123))
		ctx.Header("Set-Cookie", cookie)
	} else if initBody.Role == models.BackupServer {
		if initBody.Name == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if initBody.CoreAddress[len(initBody.CoreAddress)-1:] != "/" {
			initBody.CoreAddress += "/"
		}

		// Initialize the server as backup
		err = pack.InstanceService.InitBackup(initBody.Name, initBody.CoreAddress, initBody.CoreKey)
		if err != nil {
			pack.InstanceService.GetLocal().SetRole(models.InitServer)
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}

		ctx.JSON(http.StatusCreated, pack.InstanceService.GetLocal())

		go pack.Server.Restart()
		return
	} else if initBody.Role == models.RestoreServer {
		local := pack.InstanceService.GetLocal()
		if local.Role == models.RestoreServer {
			ctx.Status(http.StatusOK)
			return
		}

		err = pack.AccessService.AddApiKey(initBody.UsingKeyInfo)
		if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}

		local.SetRole(models.RestoreServer)
		pack.Caster.PushWeblensEvent(models.RestoreStartedEvent)

		hasherFactory := func() fileTree.Hasher {
			return models.NewHasher(pack.TaskService, pack.Caster)
		}
		journal, err := fileTree.NewJournal(pack.Db.Collection("fileHistory"), initBody.LocalId, false, hasherFactory)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		usersTree, err := fileTree.NewFileTree(filepath.Join(env.GetDataRoot(), "users"), "USERS", journal, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		pack.FileService.AddTree(usersTree)

		pack.Server.UseRestore()
		pack.Server.UseApi()

		ctx.Status(http.StatusOK)
		return
	} else {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.JSON(http.StatusCreated, pack.InstanceService.GetLocal())
	go pack.Server.Restart()
}

// GetServerInfo godoc
//
//	@ID			GetServerInfo
//
//	@Summary	Get server info
//	@Tags		Servers
//	@Produce	json
//	@Success	200 {object}	rest.ServerInfo	"Server info"
//	@Router		/info [get]
func getServerInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	// if  pack.InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	ctx.JSON(comm.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
	// 	return
	// }
	var userCount int
	if pack.UserService != nil {
		userCount = pack.UserService.Size()
	}

	serverInfo := rest.InstanceToServerInfo(pack.InstanceService.GetLocal())
	serverInfo.Started = pack.Loaded.Load()
	serverInfo.UserCount = userCount

	ctx.JSON(
		http.StatusOK,
		serverInfo,
	)
}

func resetServer(ctx *gin.Context) {
	pack := getServices(ctx)
	err := pack.InstanceService.ResetAll()
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)

	pack.Server.Restart()
}

func serveStaticContent(ctx *gin.Context) {
	filename := ctx.Param("filename")
	fullPath := env.GetAppRootDir() + "/static/" + filename
	ctx.File(fullPath)
}
