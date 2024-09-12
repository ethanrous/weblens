package http

import (
	"errors"
	"io"
	"net/http"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
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

	newRemote := models.NewInstance(nr.Id, nr.Name, nr.UsingKey, models.BackupServer, false, "")

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

	fileId := ctx.Param("fileId")
	f, err := pack.FileService.GetUserFile(fileId)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if f.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	readable, err := f.Readable()
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if closer, ok := readable.(io.Closer); ok {
		defer func() {
			log.ErrTrace(closer.Close())
		}()
	}

	_, err = io.Copy(ctx.Writer, readable)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
	}

	// ctx.File(f.GetAbsPath())
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
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	if len(ids) != 0 {
		files, err := pack.FileService.GetFiles(ids)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
		ctx.JSON(http.StatusOK, files)
		return
	}

	var files []*fileTree.WeblensFileImpl
	err = pack.FileService.GetMediaRoot().RecursiveMap(
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

	ctx.JSON(http.StatusOK, files)
}

func getRemotes(ctx *gin.Context) {
	pack := getServices(ctx)
	remotes := pack.InstanceService.GetRemotes()

	serverInfos := internal.Map(
		remotes, func(srv *models.Instance) models.ServerInfo {
			addr, _ := srv.GetAddress()
			return models.ServerInfo{
				Id:           srv.ServerId(),
				Name:         srv.GetName(),
				UsingKey:     srv.GetUsingKey(),
				Role:   srv.GetRole(),
				IsThisServer: srv.IsLocal(),
				Address:      addr,
				Online: pack.ClientService.GetClientByServerId(srv.ServerId()) != nil,
			}
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"remotes": serverInfos})
}

func removeRemote(ctx *gin.Context) {
	pack := getServices(ctx)
	body, err := readCtxBody[deleteRemoteBody](ctx)
	if err != nil {
		return
	}

	err = pack.InstanceService.Del(body.RemoteId)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
