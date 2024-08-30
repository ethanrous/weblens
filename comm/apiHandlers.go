package comm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strconv"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
)

/* ================ */

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	upInfo, err := readCtxBody[newUploadBody](ctx)
	if err != nil {
		return
	}

	// c := models.NewBufferedCaster(ClientService)
	meta := models.UploadFilesMeta{
		ChunkStream:  make(chan models.FileChunk, 10),
		RootFolderId: upInfo.RootFolderId,
		ChunkSize:    upInfo.ChunkSize,
		TotalSize:    upInfo.TotalUploadSize,
		FileService:  FileService,
		MediaService: MediaService,
		TaskService:  TaskService,
		TaskSubber:   ClientService,
		User: u,
		Caster: Caster,
	}
	t, err := TaskService.DispatchJob(models.UploadFilesTask, meta, nil)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newFileUpload(ctx *gin.Context) {
	uploadTaskId := task.TaskId(ctx.Param("uploadId"))
	newFInfo, err := readCtxBody[newFileBody](ctx)
	if err != nil {
		return
	}

	handleNewFile(uploadTaskId, newFInfo, ctx)
}

func handleNewFile(uploadTaskId task.TaskId, newFInfo newFileBody, ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	uTask := TaskService.GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Status(http.StatusNotFound)
		return
	}

	parent, err := FileService.GetFileSafe(newFInfo.ParentFolderId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	newName := service.MakeUniqueChildName(parent, newFInfo.NewFileName)

	newF, err := FileService.CreateFile(parent, newName)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = uTask.Manipulate(
		func(meta task.TaskMetadata) error {
			meta.(models.UploadFilesMeta).ChunkStream <- models.FileChunk{
				NewFile: newF, ContentRange: "0-0/" + strconv.FormatInt(newFInfo.FileSize, 10),
			}

			// TODO
			// We don't queue the upload task right away, we wait for the first file,
			// then we add the task to the queue here
			// if t.queueState == task.PreQueued {
			// 	t.Q(t.taskPool)
			// }

			return nil
		},
	)

	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newF.ID()})
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	uploadId := task.TaskId(ctx.Param("uploadId"))

	t := TaskService.GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := fileTree.FileId(ctx.Param("fileId"))

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

func getFoldersMedia(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	folderIdsStr := ctx.Query("folderIds")
	var folderIds []fileTree.FileId
	err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	var folders []*fileTree.WeblensFileImpl
	for _, folderId := range folderIds {
		f, err := FileService.GetFileSafe(folderId, u, nil)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusNotFound)
			return
		}
		folders = append(folders, f)
	}

	ctx.JSON(http.StatusOK, gin.H{"medias": MediaService.RecursiveGetMedia(folders...)})
}

func searchFolder(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	folderId := fileTree.FileId(ctx.Param("folderId"))
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir, err := FileService.GetFileSafe(folderId, u, nil)
	if dir == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !dir.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search must performed on a folder, not a regular file"})
		return
	}

	regexStr := "(?i)" + searchStr
	if filterStr != "" {
		regexStr += ".*\\." + filterStr
	}
	r, err := regexp.Compile(regexStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regex string"})
		return
	}

	var files []*fileTree.WeblensFileImpl
	err = dir.RecursiveMap(
		func(w *fileTree.WeblensFileImpl) error {
			if r.MatchString(w.Filename()) {
				if w.Filename() == ".user_trash" {
					return nil
				}
				files = append(files, w)
			}
			return nil
		},
	)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var filesData []FileInfo
	for _, f := range files {
		info, err := formatFileSafe(f, u, nil)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		filesData = append(filesData, info)
	}

	ctx.JSON(http.StatusOK, gin.H{"files": filesData})
}

func getFile(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}

	fileId := fileTree.FileId(ctx.Param("fileId"))
	file, err := FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := formatFileSafe(file, u, nil)
	if err != nil {
		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

func updateFile(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	fileId := fileTree.FileId(ctx.Param("fileId"))
	updateInfo, err := readCtxBody[fileUpdateBody](ctx)
	if err != nil {
		return
	}

	file, err := FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if FileService.IsFileInTrash(file) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot rename file in trash"})
		return
	}

	err = FileService.RenameFile(file, updateInfo.NewName, Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	fileIds, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}
	u := getUserFromCtx(ctx)

	var failed []fileTree.FileId

	for _, fileId := range fileIds {
		file, err := FileService.GetFileSafe(fileId, u, nil)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to trash"})
			return
		}

		if u != FileService.GetFileOwner(file) {
			ctx.Status(http.StatusNotFound)
			return
		}

		err = FileService.MoveFileToTrash(file, u, nil, Caster)
		if err != nil {
			log.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func deleteFiles(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	fileIds, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}
	// caster := models.NewBufferedCaster(ClientService)
	// defer caster.Close()

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := FileService.GetFileSafe(fileId, u, nil)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		} else if u != FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		} else if !FileService.IsFileInTrash(file) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete file not in trash"})
			return
		}
		files = append(files, file)
	}

	err = FileService.PermanentlyDeleteFiles(files, Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func unTrashFiles(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []fileTree.FileId
	err = json.Unmarshal(bodyBytes, &fileIds)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		log.ShowErr(err)
		return
	}

	caster := models.NewBufferedCaster(ClientService)
	defer caster.Close()

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := FileService.GetFileSafe(fileId, u, nil)
		if err != nil || u != FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to untrash"})
			return
		}
		files = append(files, file)
	}

	err = FileService.ReturnFilesFromTrash(files, caster)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func createTakeout(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		u = UserService.GetPublicUser()
	}

	takeoutRequest, err := readCtxBody[takeoutFiles](ctx)
	if err != nil {
		return
	}
	if len(takeoutRequest.FileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot takeout 0 files"})
		return
	}

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range takeoutRequest.FileIds {
		file, err := FileService.GetFileSafe(fileId, u, share)
		if err == nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		files = append(files, file)
	}

	// If we only have 1 file, and it is not a directory, we should have requested to just
	// simply download that file on it's own, not zip it.
	if len(files) == 1 && !files[0].IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := models.NewSimpleCaster(ClientService)
	meta := models.ZipMeta{
		Files:     files,
		Requester: u,
		Share:     share,
		Caster:    caster,
	}
	t, err := TaskService.DispatchJob(models.CreateZipTask, meta, nil)

	completed, status := t.Status()
	if completed && status == task.TaskSuccess {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId()})
	}
}

func downloadFile(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		u = UserService.GetPublicUser()
	}

	fileId := fileTree.FileId(ctx.Param("fileId"))

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	file, err := FileService.GetFileSafe(fileId, u, share)
	if err != nil {
		ctx.JSON(http.StatusNotFound, err)
		return
	}

	ctx.File(file.GetAbsPath())
}

func loginUser(ctx *gin.Context) {
	userCredentials, err := readCtxBody[loginBody](ctx)
	if err != nil {
		return
	}

	u := UserService.Get(userCredentials.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if u.CheckLogin(userCredentials.Password) {
		log.Info.Printf("Valid login for [%s]\n", userCredentials.Username)

		if token := u.GetToken(); token == "" {
			ctx.Status(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, gin.H{"token": token, "user": u})
		}
	} else {
		ctx.Status(http.StatusNotFound)
	}

}

func getUserInfo(ctx *gin.Context) {
	// if types.SERV.GetFileTreeSafley() == nil {
	// 	ctx.Status(comm.StatusServiceUnavailable)
	// 	return
	// }
	u := getUserFromCtx(ctx)
	if u == nil {
		if InstanceService.GetLocal().ServerRole() == models.InitServer {
			ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return
		}
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, u)
}

func getUsers(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil || !u.IsAdmin() {
		ctx.Status(http.StatusNotFound)
		return
	}

	usersIter, err := UserService.GetAll()
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, slices.Collect(usersIter))
}

func updateUserPassword(ctx *gin.Context) {
	reqUser := getUserFromCtx(ctx)
	if reqUser == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	updateUsername := models.Username(ctx.Param("username"))
	updateUser := UserService.Get(updateUsername)

	if updateUser == nil {
		ctx.Status(http.StatusNotFound)
	}

	passUpd, err := readCtxBody[passwordUpdateBody](ctx)
	if err != nil {
		return
	}

	if updateUser.GetUsername() != reqUser.GetUsername() && !reqUser.IsOwner() {
		ctx.Status(http.StatusNotFound)
		return
	}

	if (passUpd.OldPass == "" && !reqUser.IsOwner()) || passUpd.NewPass == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both oldPassword and newPassword fields are required"})
		return
	}

	err = UserService.UpdateUserPassword(
		updateUser.GetUsername(), passUpd.OldPass, passUpd.NewPass, reqUser.IsOwner(),
	)

	if err != nil {
		log.ShowErr(err)
		switch {
		case errors.Is(err.(error), werror.ErrBadPassword):
			ctx.Status(http.StatusUnauthorized)
		default:
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.Status(http.StatusOK)
}

func setUserAdmin(ctx *gin.Context) {
	owner := getUserFromCtx(ctx)
	if !owner.IsOwner() {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	update, err := readCtxBody[newUserBody](ctx)
	if err != nil {
		return
	}

	username := models.Username(ctx.Param("username"))
	u := UserService.Get(username)

	err = UserService.SetUserAdmin(u, update.Admin)
	if err != nil {
		if errors.Is(err, werror.ErrUserNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func activateUser(ctx *gin.Context) {
	username := models.Username(ctx.Param("username"))
	u := UserService.Get(username)
	err := UserService.ActivateUser(u)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := models.Username(ctx.Param("username"))
	// User to delete username
	// *cannot* use getUserFromCtx() here because that
	// will grab the user making the request, not the
	// username from the Param  \/
	u := UserService.Get(username)
	if u == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user with given username does not exist"})
		return
	}
	err := UserService.Del(u.GetUsername())
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	err := MediaService.NukeCache()
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)

}

func searchUsers(ctx *gin.Context) {
	filter := ctx.Query("filter")
	if len(filter) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autocomplete must contain at least 2 characters"})
		return
	}

	users, err := UserService.SearchByUsername(filter)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": slices.Collect(users)})
}

func createFileShare(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shareInfo, err := readCtxBody[newShareBody](ctx)
	if err != nil {
		return
	}

	f, err := FileService.GetFileSafe(shareInfo.FileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	_, err = ShareService.GetFileShare(f)
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
			return UserService.Get(un)
		},
	)
	newShare := models.NewFileShare(f, u, accessors, shareInfo.Public, shareInfo.Wormhole)

	err = ShareService.Add(newShare)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, newShare)
}

func deleteShare(ctx *gin.Context) {
	shareId := models.ShareId(ctx.Param("shareId"))

	s := ShareService.Get(shareId)
	if s == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	err := ShareService.Del(s.GetShareId())
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func addUserToFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	if !AccessService.CanUserModifyShare(user, share) {
		ctx.Status(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[userListBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	users := internal.Map(ub.Users, func(un models.Username) *models.User { return UserService.Get(un) })
	err = ShareService.AddUsers(share, users)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func newApiKey(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if !u.IsAdmin() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	newKey, err := AccessService.GenerateApiKey(u)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"key": newKey})
}

func getApiKeys(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	keys, err := AccessService.GetAllKeys(u)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
}

func deleteApiKey(ctx *gin.Context) {
	key := models.WeblensApiKey(ctx.Param("keyId"))
	keyInfo, err := AccessService.Get(key)
	if err != nil || keyInfo.Key == "" {
		log.ShowErr(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	err = AccessService.Del(key)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getFolderStats(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		u = UserService.GetPublicUser()
	}

	fileId := fileTree.FileId(ctx.Param("folderId"))

	rootFolder, err := FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !rootFolder.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusNotImplemented)

	// t := types.SERV.TaskDispatcher.GatherFsStats(rootFolder, Caster)
	// t.Wait()
	// res := t.GetResult("sizesByExtension")
	//
	// ctx.JSON(comm.StatusOK, res)
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
	// Can't init server if already initialized
	if InstanceService.GetLocal().ServerRole() != models.InitServer {
		ctx.Status(http.StatusNotFound)
		return
	}

	si, err := readCtxBody[initServerBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if si.Role == models.CoreServer {
		localCore := models.NewInstance("", si.Name, si.CoreKey, models.BackupServer, true, si.CoreAddress)
		err := InstanceService.InitCore(localCore)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		ctx.Status(http.StatusCreated)
	} else if si.Role == models.BackupServer {

		if si.CoreAddress[len(si.CoreAddress)-1:] != "/" {
			si.CoreAddress += "/"
		}

		// proxyStore := proxy.NewProxyStore(si.CoreAddress, si.CoreKey)
		// proxyStore.Init(types.SERV.StoreService)
		// types.SERV.SetStore(proxyStore)
		//
		// err = InstanceService.InitBackup(si.Name, si.CoreAddress, si.CoreKey, proxyStore)
		// if err != nil {
		// 	types.SERV.SetStore(dbStore)
		// 	wlog.ShowErr(err)
		// 	ctx.Status(http.StatusBadRequest)
		// 	return
		// }
		//
		// err = UserService.Init(proxyStore)
		// if err != nil {
		// 	wlog.ShowErr(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
		//
		// err = FileService.GetJournal().Init(proxyStore)
		// if err != nil {
		// 	wlog.ShowErr(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
		//
		// err = FileService.Init(proxyStore)
		// if err != nil {
		// 	wlog.ShowErr(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
		//
		// hashCaster := NewCaster(ClientService)
		// err = FileService.InitMediaRoot(hashCaster)
		// if err != nil {
		// 	wlog.ErrTrace(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
		//
		// core := InstanceService.GetCore()
		// err = WebsocketToCore(core, ClientService)
		// if err != nil {
		// 	wlog.ErrTrace(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
		//
		// for types.SERV.ClientManager.GetClientByInstanceId(core.ServerId()) == nil {
		// 	wlog.Info.Println("Waiting for core websocket to connect")
		// 	time.Sleep(RetryInterval)
		// }
		//
		// meta := weblens.BackupMeta{RemoteId: core.ServerId()}
		// _, err = TaskService.DispatchJob(weblens.BackupTask, meta, nil)
		// if err != nil {
		// 	wlog.ShowErr(err)
		// 	ctx.Status(http.StatusInternalServerError)
		// 	return
		// }
	}

	// We must spawn a go routine for a router restart coming from an HTTP request,
	// or else we will enter a deadlock where the router waits for this HTTP request to finish,
	// and this thread waits for the router to close...
	go func() {
		err = Server.Shutdown(context.Background())
		if err != nil {
			log.ErrTrace(err)
		}
	}()

	ctx.Status(http.StatusCreated)
}

func getServerInfo(ctx *gin.Context) {
	// if InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	ctx.JSON(comm.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
	// 	return
	// }

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"info":      InstanceService.GetLocal(),
			"started":   InstanceService.IsLocalLoaded(),
			"userCount": UserService.Size(),
		},
	)
}

func serveStaticContent(ctx *gin.Context) {
	filename := ctx.Param("filename")
	fullPath := internal.GetAppRootDir() + "/static/" + filename
	ctx.File(fullPath)
}
