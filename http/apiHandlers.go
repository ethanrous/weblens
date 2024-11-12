package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	"github.com/go-chi/chi/v5"
)

// Create new file upload task, and wait for data
func newUploadTask(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}
	upInfo, err := readCtxBody[rest.NewUploadBody](w, r)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJson(w, http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newFileUpload(w http.ResponseWriter, r *http.Request) {
	uploadTaskId := chi.URLParam(r, "uploadId")
	newFInfo, err := readCtxBody[rest.NewFileBody](w, r)
	if err != nil {
		return
	}

	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}
	uTask := pack.TaskService.GetTask(uploadTaskId)
	if uTask == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	parent, err := pack.FileService.GetFileSafe(newFInfo.ParentFolderId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	child, _ := parent.GetChild(newFInfo.NewFileName)
	if child != nil {
		writeJson(w, http.StatusConflict, gin.H{"error": "File with the same name already exists in folder"})
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

	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusCreated, gin.H{"fileId": newFId})
}

// Add chunks of file to previously created task
func handleUploadChunk(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	uploadId := chi.URLParam(r, "uploadId")

	t := pack.TaskService.GetTask(uploadId)
	if t == nil {
		writeJson(w, http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := chi.URLParam(r, "fileId")

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := internal.OracleReader(r.Body, r.ContentLength)
	if err != nil {
		log.ShowErr(err)
		// err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		// if err != nil {
		// 	util.ShowErr(err)
		// }
		writeJson(w, http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	err = t.Manipulate(
		func(meta task.TaskMetadata) error {
			chunkData := models.FileChunk{FileId: fileId, Chunk: chunk, ContentRange: r.Header["Content-Range"][0]}
			meta.(models.UploadFilesMeta).ChunkStream <- chunkData

			return nil
		},
	)

	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func clearCache(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	err := pack.MediaService.NukeCache()
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func createFileShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	shareInfo, err := readCtxBody[rest.NewShareBody](w, r)
	if err != nil {
		return
	}

	f, err := pack.FileService.GetFileSafe(shareInfo.FileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = pack.ShareService.GetFileShare(f)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		return
	} else if !errors.Is(err, werror.ErrNoShare) {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, newShare)
}

func deleteShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	shareId := models.ShareId(chi.URLParam(r, "shareId"))

	s := pack.ShareService.Get(shareId)
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := pack.ShareService.Del(s.ID())
	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func patchShareAccessors(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	share, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(u, share) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[rest.UserListBody](w, r)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var addUsers []*models.User
	for _, un := range ub.AddUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			writeJson(w, http.StatusNotFound, gin.H{"error": "Could not find user with name " + un})
			return
		}
		addUsers = append(addUsers, u)
	}

	var removeUsers []*models.User
	for _, un := range ub.RemoveUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			writeJson(w, http.StatusNotFound, gin.H{"error": "Could not find user with name " + un})
			return
		}
		removeUsers = append(removeUsers, u)
	}

	if len(addUsers) > 0 {
		err = pack.ShareService.AddUsers(share, addUsers)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if len(removeUsers) > 0 {
		err = pack.ShareService.RemoveUsers(share, removeUsers)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func setSharePublic(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	share, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(u, share) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pub, err := readCtxBody[rest.SharePublicityBody](w, r)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = pack.ShareService.SetSharePublic(share, pub.Public)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getRandomMedias(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
	// numStr := r.URL.Query().Get("count")
	// numPhotos, err := strconv.Atoi(numStr)
	// if err != nil {
	// 	w.WriteHeader(comm.StatusBadRequest)
	// 	return
	// }

	// media := media.GetRandomMedia(numPhotos)
	// writeJson(w, comm.StatusOK, gin.H{"medias": media})
}

func initializeServer(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	// Can't init server if already initialized
	role := pack.InstanceService.GetLocal().GetRole()
	if role != models.InitServerRole {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	initBody, err := readCtxBody[rest.InitServerBody](w, r)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if initBody.Role == models.CoreServerRole {
		if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = pack.InstanceService.InitCore(initBody.Name)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		users, err := pack.UserService.GetAll()
		if SafeErrorAndExit(err, w) {
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, expires, err := pack.AccessService.GenerateJwtToken(owner)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := fmt.Sprintf("%s=%s; expires=%s;", SessionTokenCookie, token, expires.Format(time.RFC1123))
		w.Header().Set("Set-Cookie", cookie)

		// hasherFactory := func() fileTree.Hasher {
		// 	return models.NewHasher(pack.TaskService, pack.Caster)
		// }
		//
		// journal, err := fileTree.NewJournal(pack.Db.Collection("fileHistory"), pack.InstanceService.GetLocal().ServerId(), false, hasherFactory)
		// if SafeErrorAndExit(err, w) {
		// 	return
		// }
		//
		// usersTree, err := fileTree.NewFileTree(filepath.Join(env.GetDataRoot(), "users"), "USERS", journal, false)
		// if SafeErrorAndExit(err, w) {
		// 	return
		// }
		// pack.FileService.AddTree(usersTree)
		pack.Server.Restart()
	} else if initBody.Role == models.BackupServerRole {
		if initBody.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if initBody.CoreAddress[len(initBody.CoreAddress)-1:] != "/" {
			initBody.CoreAddress += "/"
		}

		// Initialize the server as backup
		err = pack.InstanceService.InitBackup(initBody.Name, initBody.CoreAddress, initBody.CoreKey)
		if err != nil {
			pack.InstanceService.GetLocal().SetRole(models.InitServerRole)
			log.ShowErr(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		writeJson(w, http.StatusCreated, pack.InstanceService.GetLocal())

		// go pack.Server.Restart()
		return
	} else if initBody.Role == models.RestoreServerRole {
		local := pack.InstanceService.GetLocal()
		if local.Role == models.RestoreServerRole {
			w.WriteHeader(http.StatusOK)
			return
		}

		err = pack.AccessService.AddApiKey(initBody.UsingKeyInfo)
		if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
			safe, code := werror.TrySafeErr(err)
			writeJson(w, code, safe)
			return
		}

		local.SetRole(models.RestoreServerRole)
		pack.Caster.PushWeblensEvent(models.RestoreStartedEvent)

		hasherFactory := func() fileTree.Hasher {
			return models.NewHasher(pack.TaskService, pack.Caster)
		}
		journal, err := fileTree.NewJournal(pack.Db.Collection("fileHistory"), initBody.LocalId, false, hasherFactory)
		if SafeErrorAndExit(err, w) {
			return
		}
		usersTree, err := fileTree.NewFileTree(filepath.Join(env.GetDataRoot(), "users"), "USERS", journal, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		pack.FileService.AddTree(usersTree)

		// pack.Server.UseRestore()
		// pack.Server.UseApi()

		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	writeJson(w, http.StatusCreated, pack.InstanceService.GetLocal())
	// go pack.Server.Restart()
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
func getServerInfo(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	// if  pack.InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	writeJson(w, comm.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
	// 	return
	// }
	var userCount int
	if pack.UserService != nil {
		userCount = pack.UserService.Size()
	}

	serverInfo := rest.InstanceToServerInfo(pack.InstanceService.GetLocal())
	serverInfo.Started = pack.Loaded.Load()
	serverInfo.UserCount = userCount

	writeJson(w,
		http.StatusOK,
		serverInfo,
	)
}

func resetServer(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	err := pack.InstanceService.ResetAll()
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)

	// pack.Server.Restart()
}

func serveStaticContent(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	fullPath := env.GetAppRootDir() + "/static/" + filename
	f, err := os.Open(fullPath)
	if SafeErrorAndExit(err, w) {
		return
	}

	_, err = io.Copy(w, f)
	if SafeErrorAndExit(err, w) {
		return
	}
}
