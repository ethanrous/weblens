package http

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

func ping(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	local := pack.InstanceService.GetLocal()
	if local == nil {
		writeJson(w, http.StatusServiceUnavailable, rest.WeblensErrorInfo{Error: "weblens not initialized"})
		return
	}
	writeJson(w, http.StatusOK, gin.H{"id": local.ServerId()})
}

func getFileBytes(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	remote := getInstanceFromCtx(r)
	if remote == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	contentId := chi.URLParam(r, "contentId")
	f, err := pack.FileService.GetFileByContentId(contentId)
	if SafeErrorAndExit(err, w) {
		return
	}
	if f.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, f.AbsPath())

	// readable, err := f.Readable()
	// if err != nil {
	// 	safe, code := werror.TrySafeErr(err)
	// 	writeJson(w, code, rest.WeblensErrorInfo{Error: safe})
	// 	return
	// }
	// if closer, ok := readable.(io.Closer); ok {
	// 	defer func() {
	// 		log.Trace.Func(func(l log.Logger) {l.Printf("Closing file %s after reading content", f.Filename())})
	// 		log.ErrTrace(closer.Close())
	// 	}()
	// }

	// _, err = io.Copy(ctx.Writer, readable)
	// if err != nil {
	// 	log.ErrTrace(werror.WithStack(err))
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)
}

func getFileMeta(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	fileId := chi.URLParam(r, "fileId")
	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeJson(w, http.StatusOK, f)
}

func getFilesMeta(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	ids, err := readCtxBody[[]fileTree.FileId](w, r)
	if err != nil {
		return
	}

	if len(ids) != 0 {
		files, lostFiles, err := pack.FileService.GetFiles(ids)
		if SafeErrorAndExit(err, w) {
			return
		}
		writeJson(w, http.StatusOK, gin.H{"files": files, "lostFiles": lostFiles})
		return
	}

	var files []*fileTree.WeblensFileImpl
	usersTree := pack.FileService.GetFileTreeByName("USERS")
	err = usersTree.GetRoot().RecursiveMap(
		func(file *fileTree.WeblensFileImpl) error {
			files = append(files, file)
			return nil
		},
	)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusOK, gin.H{"files": files, "lostFiles": []fileTree.FileId{}})
}

func getApiKeysArchive(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	instance := getInstanceFromCtx(r)

	if instance == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usingKey, err := pack.AccessService.GetApiKey(instance.GetUsingKey())
	if SafeErrorAndExit(err, w) {
		return
	}

	owner := pack.UserService.Get(usingKey.Owner)

	keys, err := pack.AccessService.GetAllKeys(owner)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusOK, keys)
}

func getInstancesArchive(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	remotes := pack.InstanceService.GetRemotes()

	remotes = append(remotes, pack.InstanceService.GetLocal())

	writeJson(w, http.StatusOK, remotes)
}

func restoreHistory(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	lifetimes, err := readCtxBody[[]*fileTree.Lifetime](w, r)
	if err != nil {
		return
	}

	err = pack.FileService.RestoreHistory(lifetimes)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func restoreUsers(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	users, err := readCtxBody[[]*models.User](w, r)
	if err != nil {
		return
	}

	for _, user := range users {
		err = pack.FileService.CreateUserHome(user)
		if SafeErrorAndExit(err, w) {
			return
		}
		err = pack.UserService.Add(user)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func uploadRestoreFile(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	fileId := r.URL.Query().Get("fileId")
	if fileId == "" {
		log.Trace.Func(func(l log.Logger) { l.Printf("No fileId given") })
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	journal := pack.FileService.GetJournalByTree("USERS")
	lt := journal.Get(fileId)
	if lt == nil {
		log.Trace.Func(func(l log.Logger) { l.Printf("Could not find lifetime with id %s", fileId) })
		w.WriteHeader(http.StatusNotFound)
		return
	}

	parentId := lt.GetLatestAction().GetParentId()
	if parentId == "" {
		log.Trace.Func(func(l log.Logger) { l.Printf("Did not find parentId on latest action") })
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parent, err := pack.FileService.GetFileByTree(parentId, "USERS")
	if SafeErrorAndExit(err, w) {
		return
	}

	f, err := pack.FileService.CreateFile(parent, filepath.Base(lt.GetLatestAction().GetDestinationPath()), nil, pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	bs, err := io.ReadAll(r.Body)
	if SafeErrorAndExit(err, w) {
		return
	}

	_, err = f.Write(bs)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func restoreApiKeys(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	keys, err := readCtxBody[[]models.ApiKey](w, r)
	if err != nil {
		return
	}

	for _, key := range keys {
		err = pack.AccessService.AddApiKey(key)
		if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
			safe, code := log.TrySafeErr(err)
			writeJson(w, code, safe)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func restoreInstances(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	remotes, err := readCtxBody[[]*models.Instance](w, r)
	if err != nil {
		return
	}

	for _, r := range remotes {
		err = pack.InstanceService.Add(r)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	w.WriteHeader(http.StatusOK)

}

func finalizeRestore(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	err := pack.InstanceService.InitCore(pack.InstanceService.GetLocal().Name)
	if SafeErrorAndExit(err, w) {
		return
	}

	pack.Server.Restart(true)

	w.WriteHeader(http.StatusOK)
}
