package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/proxy"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/websocket"

	"github.com/gin-gonic/gin"
)

func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
	if ctx.Request.Method == "GET" {
		err = error2.WErrMsg("trying to get body of get request")
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
		return obj, error2.ErrNoBody
	} else if resp.ContentLength == -1 {
		wlog.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyB, &obj)
	return
}

func readRespBodyRaw(resp *http.Response) (bodyB []byte, err error) {
	if resp.ContentLength == 0 {
		return nil, error2.ErrNoBody
	} else if resp.ContentLength == -1 {
		wlog.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	return
}

func getUserFromCtx(ctx *gin.Context) *weblens.User {
	user, ok := ctx.Get("user")
	if !ok {
		return nil
	}
	return user.(*weblens.User)
	// return dataStore.GetUser(weblens.Username(user.GetUsername()))
}

/* ================ */

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	upInfo, err := readCtxBody[newUploadBody](ctx)
	if err != nil {
		return
	}
	c := websocket.NewBufferedCaster()
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

	parent, err := FileService.GetFileByIdAndRoot(newFInfo.ParentFolderId, "MEDIA")
	if parent == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	newName := weblens.MakeUniqueChildName(parent, newFInfo.NewFileName)

	newF, err := FileService.Touch(parent, newName, true, nil, types.SERV.Caster)
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

	fileId := fileTree.FileId(ctx.Param("fileId"))

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := internal.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
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
	var folderIds []fileTree.FileId
	err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	folders := internal.Map(
		folderIds, func(fId fileTree.FileId) *fileTree.WeblensFile {
			return FileService.GetFileByIdAndRoot(fId, "MEDIA")
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"medias": weblens.RecursiveGetMedia(types.SERV.MediaRepo, folders...)})
}

func searchFolder(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	folderId := fileTree.FileId(ctx.Param("folderId"))
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir, err := FileService.GetFileByIdAndRoot(folderId, "MEDIA")
	if dir == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !dir.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search must performed on a folder, not a regular file"})
		return
	}

	if !AccessService.CanUserAccessFile(u, dir) {
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

	var files []*fileTree.WeblensFile
	err = dir.RecursiveMap(
		func(w *fileTree.WeblensFile) error {
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

	filesData := internal.Map(
		files, func(w *fileTree.WeblensFile) weblens.FileInfo {
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

	fileId := fileTree.FileId(ctx.Param("fileId"))
	file, err, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := FileService.FormatFileInfo(file, u, nil)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

func updateFile(ctx *gin.Context) {
	fileId := fileTree.FileId(ctx.Param("fileId"))
	updateInfo, err := readCtxBody[fileUpdateBody](ctx)
	if err != nil {
		return
	}

	if fileId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "fileId is required to update file"})
		return
	}

	file, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if FileService.IsFileInTrash(file) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot rename file in trash"})
		return
	}

	// If the directory does not change, just assume this is a rename
	if updateInfo.NewParentId == "" {
		updateInfo.NewParentId = file.GetParent().ID()
	}

	caster := websocket.NewBufferedCaster()
	defer caster.Close()
	event := fileTree.NewFileEvent()
	t := types.SERV.TaskDispatcher.MoveFile(fileId, updateInfo.NewParentId, updateInfo.NewName, event, caster)
	t.Wait()

	if t.ReadError() != nil {
		wlog.Error.Println(t.ReadError())
		ctx.Status(http.StatusBadRequest)
		return
	}

	err = FileService.GetJournal().LogEvent(event)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
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

	caster := websocket.NewBufferedCaster()
	defer caster.Close()

	var failed []fileTree.FileId

	acc := dataStore.NewAccessMeta(u)
	shareId := weblens.ShareId(ctx.Query("shareId"))
	if shareId != "" {
		sh := ShareService.Get(shareId)
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

	event := fileTree.NewFileEvent()

	for _, fileId := range fileIds {
		file, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to trash"})
			return
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot trash file"})
			return
		}
		err := weblens.MoveFileToTrash(file, acc, event, caster)
		if err != nil {
			wlog.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
	}

	err = FileService.GetJournal().LogEvent(event)
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
	fileIds, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}
	var failed []fileTree.FileId

	caster := websocket.NewBufferedCaster()
	defer caster.Close()

	for _, fileId := range fileIds {
		file, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot delete file"})
			return
		}

		err := weblens.PermanentlyDeleteFile(file, caster)
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

	var fileIds []fileTree.FileId
	err = json.Unmarshal(bodyBytes, &fileIds)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		wlog.ShowErr(err)
		return
	}

	caster := websocket.NewBufferedCaster()
	defer caster.Close()

	var failed []fileTree.FileId

	event := fileTree.NewFileEvent()

	for _, fileId := range fileIds {
		file, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
		if file == nil {
			wlog.ErrTrace(types.ErrNoFile(fileId))
			failed = append(failed, fileId)
			continue
		}
		if u != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot un-trash file"})
			return
		}

		err = weblens.ReturnFileFromTrash(file, event, caster)
		if err != nil {
			wlog.ErrTrace(err)
			failed = append(failed, fileId)
		}
	}

	err = FileService.GetJournal().LogEvent(event)
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
		u = UserService.GetPublicUser()
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

	files := internal.Map(
		takeoutRequest.FileIds,
		func(fileId fileTree.FileId) *fileTree.WeblensFile { return FileService.GetFileByIdAndRoot(fileId, "MEDIA") },
	)

	acc := dataStore.NewAccessMeta(u)
	shareId := weblens.ShareId(ctx.Query("shareId"))
	if shareId != "" {
		sh := ShareService.Get(shareId)
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
		if file == *fileTree.WeblensFile(nil) || !acc.CanAccessFile(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 file, and it is not a directory, we should have requested to just download that file
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := websocket.NewCaster()
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
		u = UserService.GetPublicUser()
	}

	fileId := fileTree.FileId(ctx.Param("fileId"))
	shareId := weblens.ShareId(ctx.Query("shareId"))

	file, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	acc := dataStore.NewAccessMeta(u)

	sh := ShareService.Get(shareId)
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

	u := UserService.Get(userCredentials.Username)
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
		if InstanceService.GetLocal().ServerRole() == types.Initialization {
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

	us, err := UserService.GetAll()
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, us)
}

func updateUserPassword(ctx *gin.Context) {
	reqUser := getUserFromCtx(ctx)
	if reqUser == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	updateUsername := weblens.Username(ctx.Param("username"))
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
		wlog.ShowErr(err)
		switch {
		case errors.Is(err.(error), types.ErrBadPassword):
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

	username := weblens.Username(ctx.Param("username"))
	u := UserService.Get(username)

	err = UserService.SetUserAdmin(u, update.Admin)
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
	username := weblens.Username(ctx.Param("username"))
	u := UserService.Get(username)
	err := UserService.ActivateUser(u)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := weblens.Username(ctx.Param("username"))
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

	users, err := UserService.SearchByUsername(filter)
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

	f, err := FileService.GetFileByIdAndRoot(shareInfo.FileId, "MEDIA")
	if f.GetShare() != nil {
		ctx.Status(http.StatusConflict)
		return
	}

	accessors := internal.Map(
		shareInfo.Users, func(un weblens.Username) *weblens.User {
			return UserService.Get(un)
		},
	)
	newShare := weblens.NewFileShare(f, u, accessors, shareInfo.Public, shareInfo.Wormhole)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	err = ShareService.Add(newShare)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"shareData": newShare})
}

func deleteShare(ctx *gin.Context) {
	shareId := weblens.ShareId(ctx.Param("shareId"))

	s := ShareService.Get(shareId)
	if s == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	err := ShareService.Del(s.GetShareId())
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func addUserToFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	shareID := weblens.ShareId(ctx.Param("shareId"))
	sh := ShareService.Get(shareID)

	if !AccessService.CanUserAccessShare(user, sh) {
		ctx.Status(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[userListBody](ctx)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	users := internal.Map(ub.Users, func(un weblens.Username) *weblens.User { return UserService.Get(un) })
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
	if !u.IsAdmin() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	newKey, err := AccessService.GenerateApiKey()
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

	keys, err := AccessService.GetAllKeys()
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
}

func deleteApiKey(ctx *gin.Context) {
	key := weblens.WeblensApiKey(ctx.Param("keyId"))
	keyInfo, err := AccessService.GetApiKeyById(key)
	if err != nil || keyInfo.Key == "" {
		wlog.ShowErr(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	err = AccessService.Del(key)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getFolderStats(ctx *gin.Context) {
	fileId := fileTree.FileId(ctx.Param("folderId"))
	rootFolder, err := FileService.GetFileByIdAndRoot(fileId, "MEDIA")
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
	if InstanceService.GetLocal().ServerRole() != types.Initialization {
		ctx.Status(http.StatusNotFound)
		return
	}

	si, err := readCtxBody[initServerBody](ctx)
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if si.Role == Core {
		localCore := weblens.New("", si.Name, si.CoreKey, BackupServer, true, si.CoreAddress)
		err := InstanceService.InitCore(localCore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		ctx.Status(http.StatusCreated)

	} else if si.Role == BackupServer {
		dbStore := types.SERV.StoreService

		if si.CoreAddress[len(si.CoreAddress)-1:] != "/" {
			si.CoreAddress += "/"
		}

		proxyStore := proxy.NewProxyStore(si.CoreAddress, si.CoreKey)
		proxyStore.Init(types.SERV.StoreService)
		types.SERV.SetStore(proxyStore)

		err = InstanceService.InitBackup(si.Name, si.CoreAddress, si.CoreKey, proxyStore)
		if err != nil {
			types.SERV.SetStore(dbStore)
			wlog.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}

		err = UserService.Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = FileService.GetJournal().Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		err = FileService.Init(proxyStore)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		hashCaster := websocket.NewCaster()
		err = FileService.InitMediaRoot(hashCaster)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		core := InstanceService.GetCore()
		err = websocket.WebsocketToCore(core)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		for types.SERV.ClientManager.GetClientByInstanceId(core.ServerId()) == nil {
			wlog.Info.Println("Waiting for core websocket to connect")
			time.Sleep(websocket.retryInterval)
		}

		types.SERV.TaskDispatcher.Backup(core.ServerId(), types.SERV.Caster)
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
	// if InstanceService.GetLocal().ServerRole() == types.Initialization {
	// 	ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
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
