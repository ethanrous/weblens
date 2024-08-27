package comm

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
	local := InstanceService.GetLocal()
	if local == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "weblens not initialized"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": local.ServerId()})
}

func attachRemote(ctx *gin.Context) {
	local := InstanceService.GetLocal()
	if local.ServerRole() == models.BackupServer {
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

	err = InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, werror.ErrKeyInUse) {
			ctx.Status(http.StatusConflict)
			return
		}

		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, InstanceService.GetLocal())
}

func getFileBytes(ctx *gin.Context) {
	u := getUserFromCtx(ctx)

	fileId := fileTree.FileId(ctx.Param("fileId"))
	f, err := FileService.GetFileSafe(fileId, u, nil)
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
	defer readable.Close()

	_, err = io.Copy(ctx.Writer, readable)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
	}

	// ctx.File(f.GetAbsPath())
}

func getFileMeta(ctx *gin.Context) {
	u := getUserFromCtx(ctx)

	fileId := fileTree.FileId(ctx.Param("fileId"))
	f, err := FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, f)
}

func getFilesMeta(ctx *gin.Context) {
	files := []*fileTree.WeblensFile{}
	err := FileService.GetMediaRoot().RecursiveMap(
		func(file *fileTree.WeblensFile) error {
			files = append(files, file)
			return nil
		},
	)
	if err != nil {
		log.ErrTrace(err)
	}
	// files, err := FileService.GetAllFiles()
	// if err != nil {
	// 	util.ErrTrace(err)
	// 	ctx.Status(comm.StatusInternalServerError)
	// 	return
	// }

	ctx.JSON(http.StatusOK, files)

	// fIds, err := readCtxBody[[]fileTree.FileId](ctx)
	// if err != nil {
	// 	return
	// }
	// var files []map[string]any
	// var notFound []fileTree.FileId
	//
	// if len(fIds) == 0 {
	// }
	//
	// for _, id := range fIds {
	// 	f := FileService.Get(id)
	// 	if f == nil {
	// 		notFound = append(notFound, id)
	// 	} else {
	// 		files = append(files, f.MarshalArchive())
	// 	}
	// }
	// ctx.JSON(comm.StatusOK, gin.H{"files": files, "notFound": notFound})
}

func getRemotes(ctx *gin.Context) {
	srvs := InstanceService.GetRemotes()

	serverInfos := internal.Map(
		srvs, func(srv *models.Instance) models.ServerInfo {
			addr, _ := srv.GetAddress()
			return models.ServerInfo{
				Id:           srv.ServerId(),
				Name:         srv.GetName(),
				UsingKey:     srv.GetUsingKey(),
				Role:         srv.ServerRole(),
				IsThisServer: srv.IsLocal(),
				Address:      addr,
				Online: ClientService.GetClientByInstanceId(srv.ServerId()) != nil,
			}
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"remotes": serverInfos})
}

func removeRemote(ctx *gin.Context) {
	body, err := readCtxBody[deleteRemoteBody](ctx)
	if err != nil {
		return
	}

	err = InstanceService.Del(body.RemoteId)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
