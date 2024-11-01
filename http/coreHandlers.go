package http

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

func ping(ctx *gin.Context) {
	pack := getServices(ctx)
	local := pack.InstanceService.GetLocal()
	if local == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "weblens not initialized"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": local.ServerId()})
}

func attachRemote(ctx *gin.Context) {
	pack := getServices(ctx)
	local := pack.InstanceService.GetLocal()
	if local.GetRole() == models.BackupServer {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "this weblens server is running in backup mode. core mode is required to attach a remote"},
		)
		return
	}

	nr, err := readCtxBody[newServerBody](ctx)
	if err != nil {
		return
	}

	newRemote := models.NewInstance(nr.Id, nr.Name, nr.UsingKey, models.BackupServer, false, "", local.ServerId())

	err = pack.InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, werror.ErrKeyInUse) {
			ctx.Status(http.StatusConflict)
			return
		}

		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = pack.AccessService.SetKeyUsedBy(nr.UsingKey, newRemote)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
		return
	}

	ctx.JSON(http.StatusCreated, pack.InstanceService.GetLocal())
}

func getFileBytes(ctx *gin.Context) {
	pack := getServices(ctx)
	remote := getInstanceFromCtx(ctx)
	if remote == nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	contentId := ctx.Param("contentId")
	f, err := pack.FileService.GetFileByContentId(contentId)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
		return
	}
	if f.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.File(f.AbsPath())

	// readable, err := f.Readable()
	// if err != nil {
	// 	safe, code := werror.TrySafeErr(err)
	// 	ctx.JSON(code, gin.H{"error": safe})
	// 	return
	// }
	// if closer, ok := readable.(io.Closer); ok {
	// 	defer func() {
	// 		log.Trace.Printf("Closing file %s after reading content", f.Filename())
	// 		log.ErrTrace(closer.Close())
	// 	}()
	// }

	// _, err = io.Copy(ctx.Writer, readable)
	// if err != nil {
	// 	log.ErrTrace(werror.WithStack(err))
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }

	// ctx.Status(http.StatusOK)
}

func getFileMeta(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	fileId := ctx.Param("fileId")
	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, f)
}

func getFilesMeta(ctx *gin.Context) {
	pack := getServices(ctx)
	ids, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}

	if len(ids) != 0 {
		files, lostFiles, err := pack.FileService.GetFiles(ids)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"files": files, "lostFiles": lostFiles})
		return
	}

	var files []*fileTree.WeblensFileImpl
	err = pack.FileService.GetUsersRoot().RecursiveMap(
		func(file *fileTree.WeblensFileImpl) error {
			files = append(files, file)
			return nil
		},
	)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"files": files, "lostFiles": []fileTree.FileId{}})
}

func getRemotes(ctx *gin.Context) {
	pack := getServices(ctx)
	remotes := pack.InstanceService.GetRemotes()
	localRole := pack.InstanceService.GetLocal().GetRole()

	serverInfos := internal.Map(
		remotes, func(srv *models.Instance) models.ServerInfo {
			addr, _ := srv.GetAddress()
			client := pack.ClientService.GetClientByServerId(srv.ServerId())
			online := client != nil && client.Active.Load()

			var backupSize int64 = -1
			if localRole == models.BackupServer {
				backupSize = pack.FileService.Size(srv.ServerId())
			}
			return models.ServerInfo{
				Id:           srv.ServerId(),
				Name:         srv.GetName(),
				UsingKey:     srv.GetUsingKey(),
				Role:         srv.GetRole(),
				IsThisServer: srv.IsLocal(),
				Address:      addr,
				Online:       online,
				ReportedRole: srv.GetReportedRole(),
				LastBackup:   srv.LastBackup,
				BackupSize:   backupSize,
			}
		},
	)

	ctx.JSON(http.StatusOK, serverInfos)
}

func getApiKeysArchive(ctx *gin.Context) {
	pack := getServices(ctx)
	instance := getInstanceFromCtx(ctx)

	if instance == nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	usingKey, err := pack.AccessService.GetApiKey(instance.GetUsingKey())
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	owner := pack.UserService.Get(usingKey.Owner)

	keys, err := pack.AccessService.GetAllKeys(owner)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, keys)
}

func getInstancesArchive(ctx *gin.Context) {
	pack := getServices(ctx)
	remotes := pack.InstanceService.GetRemotes()

	remotes = append(remotes, pack.InstanceService.GetLocal())

	ctx.JSON(http.StatusOK, remotes)
}

func removeRemote(ctx *gin.Context) {
	pack := getServices(ctx)
	body, err := readCtxBody[deleteRemoteBody](ctx)
	if err != nil {
		return
	}

	remote := pack.InstanceService.GetByInstanceId(body.RemoteId)

	err = pack.InstanceService.Del(remote.DbId)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if key := remote.GetUsingKey(); key != "" {
		err = pack.AccessService.SetKeyUsedBy(key, nil)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func restoreHistory(ctx *gin.Context) {
	pack := getServices(ctx)

	lifetimes, err := readCtxBody[[]*fileTree.Lifetime](ctx)
	if err != nil {
		return
	}

	err = pack.FileService.RestoreHistory(lifetimes)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func restoreUsers(ctx *gin.Context) {
	pack := getServices(ctx)

	users, err := readCtxBody[[]*models.User](ctx)
	if err != nil {
		return
	}

	for _, user := range users {
		err = pack.FileService.CreateUserHome(user)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
		err = pack.UserService.Add(user)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func uploadRestoreFile(ctx *gin.Context) {
	pack := getServices(ctx)

	fileId := ctx.Query("fileId")
	if fileId == "" {
		log.Trace.Printf("No fileId given")
		ctx.Status(http.StatusBadRequest)
		return
	}

	journal := pack.FileService.GetJournalByTree("USERS")
	lt := journal.Get(fileId)
	if lt == nil {
		log.Trace.Printf("Could not find lifetime with id %s", fileId)
		ctx.Status(http.StatusNotFound)
		return
	}

	parentId := lt.GetLatestAction().GetParentId()
	if parentId == "" {
		log.Trace.Printf("Did not find parentId on latest action")
		ctx.Status(http.StatusBadRequest)
		return
	}

	parent, err := pack.FileService.GetFileByTree(parentId, "USERS")
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	f, err := pack.FileService.CreateFile(parent, filepath.Base(lt.GetLatestAction().GetDestinationPath()), nil)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	bs, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	_, err = f.Write(bs)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func restoreApiKeys(ctx *gin.Context) {
	pack := getServices(ctx)

	keys, err := readCtxBody[[]models.ApiKeyInfo](ctx)
	if err != nil {
		return
	}

	for _, key := range keys {
		err = pack.AccessService.AddApiKey(key)
		if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func restoreInstances(ctx *gin.Context) {
	pack := getServices(ctx)

	remotes, err := readCtxBody[[]*models.Instance](ctx)
	if err != nil {
		return
	}

	for _, r := range remotes {
		err = pack.InstanceService.Add(r)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	}

	ctx.Status(http.StatusOK)

}

func finalizeRestore(ctx *gin.Context) {
	pack := getServices(ctx)

	err := pack.InstanceService.InitCore(pack.InstanceService.GetLocal().Name)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	go pack.Server.Restart()

	ctx.Status(http.StatusOK)
}
