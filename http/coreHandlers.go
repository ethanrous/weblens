package http

import (
	"errors"
	"io"
	"net/http"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
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
	if local.ServerRole() == BackupServer {
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

	newRemote := weblens.New(nr.Id, nr.Name, nr.UsingKey, BackupServer, false, "")

	err = InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, dataStore.ErrKeyInUse) {
			ctx.Status(http.StatusConflict)
			return
		}

		wlog.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, InstanceService.GetLocal())
}

func getFileBytes(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := types.SERV.FileTree.Get(fileId)
	if f == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if f.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	readable, err := f.Readable()
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	defer readable.Close()

	_, err = io.Copy(ctx.Writer, readable)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
	}

	// ctx.File(f.GetAbsPath())
}

func getFileMeta(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := types.SERV.FileTree.Get(fileId)
	if f == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, f)
}

func getFilesMeta(ctx *gin.Context) {
	files := []*fileTree.WeblensFile{}
	err := types.SERV.FileTree.GetRoot().RecursiveMap(
		func(file *fileTree.WeblensFile) error {
			files = append(files, file)
			return nil
		},
	)
	if err != nil {
		wlog.ErrTrace(err)
	}
	// files, err := types.SERV.FileTree.GetAllFiles()
	// if err != nil {
	// 	util.ErrTrace(err)
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }

	ctx.JSON(http.StatusOK, files)

	// fIds, err := readCtxBody[[]types.FileId](ctx)
	// if err != nil {
	// 	return
	// }
	// var files []map[string]any
	// var notFound []types.FileId
	//
	// if len(fIds) == 0 {
	// }
	//
	// for _, id := range fIds {
	// 	f := types.SERV.FileTree.Get(id)
	// 	if f == nil {
	// 		notFound = append(notFound, id)
	// 	} else {
	// 		files = append(files, f.MarshalArchive())
	// 	}
	// }
	// ctx.JSON(http.StatusOK, gin.H{"files": files, "notFound": notFound})
}

func getRemotes(ctx *gin.Context) {
	srvs := InstanceService.GetRemotes()

	serverInfos := internal.Map(
		srvs, func(srv types.Instance) serverInfo {
			addr, _ := srv.GetAddress()
			return serverInfo{
				Id:           srv.ServerId(),
				Name:         srv.GetName(),
				UsingKey:     srv.GetUsingKey(),
				Role:         srv.ServerRole(),
				IsThisServer: srv.IsLocal(),
				Address:      addr,
				Online:       types.SERV.ClientManager.GetClientByInstanceId(srv.ServerId()) != nil,
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
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
