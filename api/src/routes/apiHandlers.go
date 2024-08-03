package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/dataStore/share"
	"github.com/ethrousseau/weblens/api/routes/proxy"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util/wlog"

	"github.com/ethrousseau/weblens/api/util"

	"github.com/gin-gonic/gin"
)

func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
	if ctx.Request.Method == "GET" {
		err = types.WeblensErrorMsg("trying to get body of get request")
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is not in expected JSON format"})
		return
	}

	return
}

func readRespBody[T any](resp *http.Response) (obj T, err error) {
	var bodyB []byte
	if resp.ContentLength == 0 {
		return obj, ErrNoBody
	} else if resp.ContentLength == -1 {
		wlog.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = util.OracleReader(resp.Body, resp.ContentLength)
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyB, &obj)
	return
}

func readRespBodyRaw(resp *http.Response) (bodyB []byte, err error) {
	if resp.ContentLength == 0 {
		return nil, ErrNoBody
	} else if resp.ContentLength == -1 {
		wlog.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = util.OracleReader(resp.Body, resp.ContentLength)
	}
	return
}

func getUserFromCtx(ctx *gin.Context) types.User {
	user, ok := ctx.Get("user")
	if !ok {
		return nil
	}
	return user.(types.User)
	// return dataStore.GetUser(types.Username(user.GetUsername()))
}

/* ================ */

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	upInfo, err := readCtxBody[newUploadBody](ctx)
	if err != nil {
		return
	}
	c := NewBufferedCaster()
	t := types.SERV.TaskDispatcher.WriteToFile(upInfo.RootFolderId, upInfo.ChunkSize, upInfo.TotalUploadSize, c)
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newFileUpload(ctx *gin.Context) {
	uploadTaskId := types.TaskId(ctx.Param("uploadId"))
	newFInfo, err := readCtxBody[newFileBody](ctx)
	if err != nil {
		return
	}

	handleNewFile(uploadTaskId, newFInfo, ctx)
}

func handleNewFile(uploadTaskId types.TaskId, newFInfo newFileBody, ctx *gin.Context) {
	uTask := types.SERV.TaskDispatcher.GetWorkerPool().GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Status(http.StatusNotFound)
		return
	}

	parent := types.SERV.FileTree.Get(newFInfo.ParentFolderId)
	if parent == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	newName := dataStore.MakeUniqueChildName(parent, newFInfo.NewFileName)

	newF, err := types.SERV.FileTree.Touch(parent, newName, true, nil, types.SERV.Caster)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = uTask.NewFileInStream(newF, newFInfo.FileSize)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newF.ID()})
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	uploadId := types.TaskId(ctx.Param("uploadId"))

	t := types.SERV.TaskDispatcher.GetWorkerPool().GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := types.FileId(ctx.Param("fileId"))

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := util.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
	if err != nil {
		wlog.ShowErr(err)
		// err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		// if err != nil {
		// 	util.ShowErr(err)
		// }
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	err = t.AddChunkToStream(fileId, chunk, ctx.GetHeader("Content-Range"))
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getFoldersMedia(ctx *gin.Context) {
	folderIdsStr := ctx.Query("folderIds")
	var folderIds []types.FileId
	err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	folders := util.Map(
		folderIds, func(fId types.FileId) types.WeblensFile {
			return types.SERV.FileTree.Get(fId)
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"medias": dataStore.RecursiveGetMedia(types.SERV.MediaRepo, folders...)})
}

func searchFolder(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	folderId := types.FileId(ctx.Param("folderId"))
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir := types.SERV.FileTree.Get(folderId)
	if dir == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !dir.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search must performed on a folder, not a regular file"})
		return
	}

	acc := dataStore.NewAccessMeta(u)
	if !acc.CanAccessFile(dir) {
		ctx.Status(http.StatusNotFound)
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

	var files []types.WeblensFile
	err = dir.RecursiveMap(
		func(w types.WeblensFile) error {
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
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	filesData := util.Map(
		files, func(w types.WeblensFile) types.FileInfo {
			d, err := w.FormatFileInfo(acc)
			wlog.ErrTrace(err)
			return d
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"files": filesData})
}

func getFile(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}

	fileId := types.FileId(ctx.Param("fileId"))
	file := types.SERV.FileTree.Get(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	acc := dataStore.NewAccessMeta(u)
	formattedInfo, err := file.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

func updateFile(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	updateInfo, err := readCtxBody[fileUpdateBody](ctx)
	if err != nil {
		return
	}

	if fileId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "fileId is required to update file"})
		return
	}

	file := types.SERV.FileTree.Get(fileId)
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if dataStore.IsFileInTrash(file) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot rename file in trash"})
		return
	}

	// If the directory does not change, just assume this is a rename
	if updateInfo.NewParentId == "" {
		updateInfo.NewParentId = file.GetParent().ID()
	}

	caster := NewBufferedCaster()
	defer caster.Close()
	event := history.NewFileEvent()
	t := types.SERV.TaskDispatcher.MoveFile(fileId, updateInfo.NewParentId, updateInfo.NewName, event, caster)
	t.Wait()

	if t.ReadError() != nil {
		wlog.Error.Println(t.ReadError())
		ctx.Status(http.StatusBadRequest)
		return
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(event)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	fileIds, err := readCtxBody[[]types.FileId](ctx)
	if err != nil {
		return
	}
	u := getUserFromCtx(ctx)

	caster := NewBufferedCaster()
	defer caster.Close()

	var failed []types.FileId

	acc := dataStore.NewAccessMeta(u)
	shareId := types.ShareId(ctx.Query("shareId"))
	if shareId != "" {
		sh := types.SERV.ShareService.Get(shareId)
		if sh == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		err = acc.AddShare(sh)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	event := history.NewFileEvent()

	for _, fileId := range fileIds {
		file := types.SERV.FileTree.Get(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to trash"})
			return
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot trash file"})
			return
		}
		err := dataStore.MoveFileToTrash(file, acc, event, caster)
		if err != nil {
			wlog.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(event)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func deleteFiles(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	fileIds, err := readCtxBody[[]types.FileId](ctx)
	if err != nil {
		return
	}
	var failed []types.FileId

	caster := NewBufferedCaster()
	defer caster.Close()

	for _, fileId := range fileIds {
		file := types.SERV.FileTree.Get(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot delete file"})
			return
		}

		err := dataStore.PermanentlyDeleteFile(file, caster)
		if err != nil {
			wlog.ErrTrace(err)
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

func unTrashFiles(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []types.FileId
	err = json.Unmarshal(bodyBytes, &fileIds)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		wlog.ShowErr(err)
		return
	}

	caster := NewBufferedCaster()
	defer caster.Close()

	var failed []types.FileId

	event := history.NewFileEvent()

	for _, fileId := range fileIds {
		file := types.SERV.FileTree.Get(fileId)
		if file == nil {
			wlog.ErrTrace(types.ErrNoFile(fileId))
			failed = append(failed, fileId)
			continue
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot un-trash file"})
			return
		}

		err = dataStore.ReturnFileFromTrash(file, event, caster)
		if err != nil {
			wlog.ErrTrace(err)
			failed = append(failed, fileId)
		}
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(event)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func createTakeout(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		u = types.SERV.UserService.GetPublicUser()
	}
	// if u == nil {
	//	ctx.Status(http.StatusUnauthorized)
	//	return
	// }

	takeoutRequest, err := readCtxBody[takeoutFiles](ctx)
	if err != nil {
		return
	}
	if len(takeoutRequest.FileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot takeout 0 files"})
		return
	}

	files := util.Map(
		takeoutRequest.FileIds, func(fileId types.FileId) types.WeblensFile { return types.SERV.FileTree.Get(fileId) },
	)

	acc := dataStore.NewAccessMeta(u)
	shareId := types.ShareId(ctx.Query("shareId"))
	if shareId != "" {
		sh := types.SERV.ShareService.Get(shareId)
		if sh == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find share"})
			return
		}
		err := acc.AddShare(sh)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.Status(http.StatusNotFound)
			return
		}
	}

	for _, file := range files {
		file.GetAbsPath() // Make sure directories have trailing slash
		if file == types.WeblensFile(nil) || !acc.CanAccessFile(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 file, and it is not a directory, we should have requested to just download that file
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := NewCaster()
	t := types.SERV.TaskDispatcher.CreateZip(files, u.GetUsername(), shareId, caster)

	completed, status := t.Status()
	if completed && status == dataProcess.TaskSuccess {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId()})
	}
}

func downloadFile(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		u = types.SERV.UserService.GetPublicUser()
	}

	fileId := types.FileId(ctx.Param("fileId"))
	shareId := types.ShareId(ctx.Query("shareId"))

	file := types.SERV.FileTree.Get(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	acc := dataStore.NewAccessMeta(u)

	sh := types.SERV.ShareService.Get(shareId)
	if sh != nil {
		err := acc.AddShare(sh)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.Status(http.StatusNotFound)
			return
		}
	}

	if !acc.CanAccessFile(file) {
		wlog.Debug.Println("No auth")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	ctx.File(file.GetAbsPath())
}

func loginUser(ctx *gin.Context) {
	userCredentials, err := readCtxBody[loginBody](ctx)
	if err != nil {
		return
	}

	u := types.SERV.UserService.Get(userCredentials.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if u.CheckLogin(userCredentials.Password) {
		wlog.Info.Printf("Valid login for [%s]\n", userCredentials.Username)

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
	// 	ctx.Status(http.StatusServiceUnavailable)
	// 	return
	// }
	u := getUserFromCtx(ctx)
	if u == nil {
		if types.SERV.InstanceService.GetLocal().ServerRole() == types.Initialization {
			ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return
		}
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, u)
}

func getUsers(ctx *gin.Context) {
	// util.ShowErr(types.NewWeblensError("TODO - getUsers"))
	// ctx.Status(http.StatusNotImplemented)

	us, err := types.SERV.UserService.GetAll()
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, us)
}

func updateUserPassword(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	passUpd, err := readCtxBody[passwordUpdateBody](ctx)
	if err != nil {
		return
	}

	if passUpd.OldPass == "" || passUpd.NewPass == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both oldPassword and newPassword fields are required"})
		return
	}
	err = u.UpdatePassword(passUpd.OldPass, passUpd.NewPass)
	if err != nil {
		wlog.ShowErr(err)
		switch {
		case errors.Is(err.(error), types.ErrBadPassword):
			ctx.Status(http.StatusUnauthorized)
		default:
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
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

	username := types.Username(ctx.Param("username"))
	u := types.SERV.UserService.Get(username)

	err = types.SERV.UserService.SetUserAdmin(u, update.Admin)
	if err != nil {
		if errors.Is(err, dataStore.ErrNoUser) {
			ctx.Status(http.StatusNotFound)
			return
		}
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func activateUser(ctx *gin.Context) {
	username := types.Username(ctx.Param("username"))
	u := types.SERV.UserService.Get(username)

	if err := u.Activate(); err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := types.Username(ctx.Param("username"))
	// User to delete username
	// *cannot* use getUserFromCtx() here because that
	// will grab the user making the request, not the
	// username from the Param  \/
	u := types.SERV.UserService.Get(username)
	if u == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user with given username does not exist"})
		return
	}
	err := types.SERV.UserService.Del(u.GetUsername())
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	err := types.SERV.MediaRepo.NukeCache()
	if err != nil {
		wlog.ShowErr(err)
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

	users, err := types.SERV.UserService.SearchByUsername(filter)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
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
	// if len(shareInfo.Users) != 0 && shareInfo.Public {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot create public share and specify users"})
	// 	return
	// }

	f := types.SERV.FileTree.Get(shareInfo.FileId)
	if f.GetShare() != nil {
		ctx.Status(http.StatusConflict)
		return
	}

	accessors := util.Map(
		shareInfo.Users, func(un types.Username) types.User {
			return types.SERV.UserService.Get(un)
		},
	)
	newShare := share.NewFileShare(f, u, accessors, shareInfo.Public, shareInfo.Wormhole)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	err = types.SERV.ShareService.Add(newShare)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"shareData": newShare})
}

func deleteShare(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))

	s := types.SERV.ShareService.Get(shareId)
	if s == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	err := types.SERV.ShareService.Del(s.GetShareId())
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func addUserToFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	shareID := types.ShareId(ctx.Param("shareId"))
	sh := types.SERV.ShareService.Get(shareID)

	acc := dataStore.NewAccessMeta(user)
	if !acc.CanAccessShare(sh) {
		ctx.Status(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[userListBody](ctx)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	users := util.Map(ub.Users, func(un types.Username) types.User { return types.SERV.UserService.Get(un) })
	err = sh.AddUsers(users)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func newApiKey(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	acc := dataStore.NewAccessMeta(u).SetRequestMode(dataStore.ApiKeyCreate)
	newKey, err := types.SERV.AccessService.GenerateApiKey(acc)
	if err != nil {
		wlog.ShowErr(err)
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
	acc := dataStore.NewAccessMeta(u).SetRequestMode(dataStore.ApiKeyGet)
	keys, err := types.SERV.AccessService.GetAllKeys(acc)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
}

func deleteApiKey(ctx *gin.Context) {
	key := types.WeblensApiKey(ctx.Param("keyId"))
	keyInfo := types.SERV.AccessService.Get(key)
	if keyInfo.Key == "" {
		ctx.Status(http.StatusNotFound)
		return
	}

	err := types.SERV.AccessService.Del(key)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getFolderStats(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("folderId"))
	rootFolder := types.SERV.FileTree.Get(fileId)
	if rootFolder == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !rootFolder.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	t := types.SERV.TaskDispatcher.GatherFsStats(rootFolder, types.SERV.Caster)
	t.Wait()
	res := t.GetResult("sizesByExtension")

	ctx.JSON(http.StatusOK, res)
}

func getRandomMedias(ctx *gin.Context) {
	ctx.Status(http.StatusNotImplemented)
	return
	// numStr := ctx.Query("count")
	// numPhotos, err := strconv.Atoi(numStr)
	// if err != nil {
	// 	ctx.Status(http.StatusBadRequest)
	// 	return
	// }

	// media := media.GetRandomMedia(numPhotos)
	// ctx.JSON(http.StatusOK, gin.H{"medias": media})
}

func initializeServer(ctx *gin.Context) {
	// Can't init server if already initialized
	if types.SERV.InstanceService.GetLocal().ServerRole() != types.Initialization {
		ctx.Status(http.StatusNotFound)
		return
	}

	si, err := readCtxBody[initServerBody](ctx)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if si.Role == types.Core {
		localCore := instance.New("", si.Name, si.CoreKey, types.Backup, true, si.CoreAddress)
		err := types.SERV.InstanceService.InitCore(localCore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		ctx.Status(http.StatusCreated)

	} else if si.Role == types.Backup {
		dbStore := types.SERV.StoreService

		if si.CoreAddress[len(si.CoreAddress)-1:] != "/" {
			si.CoreAddress += "/"
		}

		proxyStore := proxy.NewProxyStore(si.CoreAddress, si.CoreKey)
		proxyStore.Init(types.SERV.StoreService)
		types.SERV.SetStore(proxyStore)

		err = types.SERV.InstanceService.InitBackup(si.Name, si.CoreAddress, si.CoreKey, proxyStore)
		if err != nil {
			types.SERV.SetStore(dbStore)
			wlog.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}

		err = types.SERV.UserService.Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = types.SERV.FileTree.GetJournal().Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = types.SERV.FileTree.Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		for _, remote := range types.SERV.InstanceService.GetRemotes() {
			if remote.IsLocal() {
				continue
			}
			types.SERV.TaskDispatcher.Backup(remote.ServerId(), types.SERV.Caster)
		}
	}

	// We must spawn a go routine for a router restart coming from an HTTP request,
	// or else we will enter a deadlock where the router waits for this HTTP request to finish,
	// and this thread waits for the router to close...
	go func() {
		err = types.SERV.RestartRouter()
		if err != nil {
			wlog.ErrTrace(err)
		}
	}()

	ctx.Status(http.StatusCreated)
}

func getServerInfo(ctx *gin.Context) {
	// if types.SERV.InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
	// 	return
	// }

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"info":      types.SERV.InstanceService.GetLocal(),
			"started":   types.SERV.InstanceService.IsLocalLoaded(),
			"userCount": types.SERV.UserService.Size(),
		},
	)
}

func serveStaticContent(ctx *gin.Context) {
	filename := ctx.Param("filename")
	fullPath := util.GetAppRootDir() + "/static/" + filename
	ctx.File(fullPath)
}
