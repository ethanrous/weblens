package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
)

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	upInfo, err := readCtxBody[newUploadBody](ctx)
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
	newFInfo, err := readCtxBody[newFileBody](ctx)
	if err != nil {
		return
	}

	handleNewFile(uploadTaskId, newFInfo, ctx)
}

func handleNewFile(uploadTaskId task.Id, newFInfo newFileBody, ctx *gin.Context) {
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
	}

	newF, err := pack.FileService.CreateFile(parent, newFInfo.NewFileName)
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

func getFoldersMedia(ctx *gin.Context) {
	pack := getServices(ctx)
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
		f, err := pack.FileService.GetFileSafe(folderId, u, nil)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusNotFound)
			return
		}
		folders = append(folders, f)
	}

	ctx.JSON(http.StatusOK, gin.H{"medias": pack.MediaService.RecursiveGetMedia(folders...)})
}

func searchFolder(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	folderId := ctx.Param("folderId")
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir, err := pack.FileService.GetFileSafe(folderId, u, nil)
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
		info, err := formatFileSafe(f, u, nil, pack)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		filesData = append(filesData, info)
	}

	ctx.JSON(http.StatusOK, gin.H{"files": filesData})
}

func createTakeout(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
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
		file, err := pack.FileService.GetFileSafe(fileId, u, share)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, gin.H{"error": safe})
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

	caster := models.NewSimpleCaster(pack.ClientService)
	meta := models.ZipMeta{
		Files:       files,
		Requester:   u,
		Share:       share,
		Caster:      caster,
		FileService: pack.FileService,
	}
	t, err := pack.TaskService.DispatchJob(models.CreateZipTask, meta, nil)

	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	completed, status := t.Status()
	if completed && status == task.TaskSuccess {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId()})
	}
}

func downloadFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := ctx.Param("fileId")

	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	file, err := pack.FileService.GetFileSafe(fileId, u, share)
	if err != nil {
		ctx.JSON(http.StatusNotFound, err)
		return
	}

	ctx.File(file.GetAbsPath())
}

func downloadTakeout(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := ctx.Param("fileId")

	// share, err := getShareFromCtx[*models.FileShare](ctx)
	// if err != nil {
	// 	return
	// }

	file, err := pack.FileService.GetZip(fileId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, err)
		return
	}

	ctx.File(file.GetAbsPath())
}

func loginUser(ctx *gin.Context) {
	pack := getServices(ctx)
	userCredentials, err := readCtxBody[loginBody](ctx)
	if err != nil {
		return
	}

	u := pack.UserService.Get(userCredentials.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if u.CheckLogin(userCredentials.Password) {
		log.Info.Printf("Valid login for [%s]\n", userCredentials.Username)

		var token string
		if token = u.GetToken(); token == "" {
			token, err = pack.AccessService.GenerateJwtToken(u)
			if err != nil || token == "" {
				log.Error.Println("Could not get login token")
				ctx.Status(http.StatusInternalServerError)
			}

		}
		ctx.JSON(http.StatusOK, gin.H{"token": token, "user": u})
	} else {
		log.Error.Printf("Invalid login for [%s]", userCredentials.Username)
		ctx.Status(http.StatusNotFound)
	}

}

func getUserInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		if pack.InstanceService.GetLocal().ServerRole() == models.InitServer {
			ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return
		}
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, u)
}

func getUsers(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil || !u.IsAdmin() {
		ctx.Status(http.StatusNotFound)
		return
	}

	usersIter, err := pack.UserService.GetAll()
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, slices.Collect(usersIter))
}

func updateUserPassword(ctx *gin.Context) {
	pack := getServices(ctx)
	reqUser := getUserFromCtx(ctx)
	if reqUser == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	updateUsername := ctx.Param("username")
	updateUser := pack.UserService.Get(updateUsername)

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

	err = pack.UserService.UpdateUserPassword(
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
	pack := getServices(ctx)
	owner := getUserFromCtx(ctx)
	if !owner.IsOwner() {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	update, err := readCtxBody[newUserBody](ctx)
	if err != nil {
		return
	}

	username := ctx.Param("username")
	u := pack.UserService.Get(username)

	err = pack.UserService.SetUserAdmin(u, update.Admin)
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
	pack := getServices(ctx)
	username := ctx.Param("username")
	u := pack.UserService.Get(username)
	err := pack.UserService.ActivateUser(u)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	pack := getServices(ctx)
	username := ctx.Param("username")
	// User to delete username
	// *cannot* use getUserFromCtx() here because that
	// will grab the user making the request, not the
	// username from the Param  \/
	u := pack.UserService.Get(username)
	if u == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user with given username does not exist"})
		return
	}
	err := pack.UserService.Del(u.GetUsername())
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

func searchUsers(ctx *gin.Context) {
	pack := getServices(ctx)
	filter := ctx.Query("filter")
	if len(filter) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autocomplete must contain at least 2 characters"})
		return
	}

	users, err := pack.UserService.SearchByUsername(filter)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": slices.Collect(users)})
}

func createFileShare(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shareInfo, err := readCtxBody[newShareBody](ctx)
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

	ub, err := readCtxBody[userListBody](ctx)
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

	pub, err := readCtxBody[sharePublicityBody](ctx)
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

func newApiKey(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if !u.IsAdmin() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	newKey, err := pack.AccessService.GenerateApiKey(u)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"key": newKey})
}

func getApiKeys(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	keys, err := pack.AccessService.GetAllKeys(u)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
}

func deleteApiKey(ctx *gin.Context) {
	pack := getServices(ctx)
	key := models.WeblensApiKey(ctx.Param("keyId"))
	keyInfo, err := pack.AccessService.GetApiKey(key)
	if err != nil || keyInfo.Key == "" {
		log.ShowErr(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	err = pack.AccessService.DeleteApiKey(key)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getFolderStats(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := ctx.Param("folderId")

	rootFolder, err := pack.FileService.GetFileSafe(fileId, u, nil)
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
	pack := getServices(ctx)
	// Can't init server if already initialized
	if pack.InstanceService.GetLocal().ServerRole() != models.InitServer {
		ctx.Status(http.StatusNotFound)
		return
	}

	si, err := readCtxBody[initServerBody](ctx)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if (si.Role != models.CoreServer && si.Role != models.BackupServer) || si.Name == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	if si.Role == models.CoreServer {
		if si.Username == "" || si.Password == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}

		owner, err := models.NewUser(si.Username, si.Password, true, true)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = pack.UserService.Add(owner)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = pack.InstanceService.InitCore(si.Name)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}

		pack.Server.UseCore()
	} else if si.Role == models.BackupServer {

		if si.CoreAddress[len(si.CoreAddress)-1:] != "/" {
			si.CoreAddress += "/"
		}
		err = pack.InstanceService.InitBackup(si.Name, si.CoreAddress, si.CoreKey)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
	}

	pack.Server.UseApi()

	ctx.Status(http.StatusCreated)
}

func getServerInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	// if  pack.InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	ctx.JSON(comm.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
	// 	return
	// }

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"info":      pack.InstanceService.GetLocal(),
			"started":   pack.Loaded.Load(),
			"userCount": pack.UserService.Size(),
		},
	)
}

func serveStaticContent(ctx *gin.Context) {
	filename := ctx.Param("filename")
	fullPath := env.GetAppRootDir() + "/static/" + filename
	ctx.File(fullPath)
}
