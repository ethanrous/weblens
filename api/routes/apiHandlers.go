package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"

	"github.com/ethrousseau/weblens/api/util"

	"github.com/gin-gonic/gin"
)

func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
	if ctx.Request.Method == "GET" {
		err = ErrBodyNotAllowed
		return
	}
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		util.ShowErr(err)
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
		util.Warning.Println("Reading body with unknown content length")
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
		util.Warning.Println("Reading body with unknown content length")
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

func getMediaBatch(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"

	var albumFilter []types.AlbumId
	err := json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	media, err := media.GetFilteredMedia(user, sort, -1, albumFilter, raw)
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": media})
}

func getProcessedMedia(ctx *gin.Context, q types.Quality) {
	mediaId := types.ContentId(ctx.Param("mediaId"))

	var pageNum int
	var err error
	pageString := ctx.Query("page")
	if pageString != "" {
		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad page number"})
			return
		}
	}

	m := rc.MediaRepo.Get(mediaId)
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given ID not found"})
		return
	}
	bs, err := m.ReadDisplayable(q, rc.FileTree, pageNum)

	if errors.Is(err, dataStore.ErrNoCache) {
		util.Warning.Println("Did not find cache for media file")
		f := rc.FileTree.Get(m.GetFiles()[0])
		if f != nil {
			rc.TaskDispatcher.ScanDirectory(f.GetParent(), rc.Caster)
			ctx.Status(http.StatusNoContent)
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
			return
		}
	}
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
		return
	}

	_, err = ctx.Writer.Write(bs)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func getMediaThumbnail(ctx *gin.Context) {
	getProcessedMedia(ctx, dataStore.Thumbnail)
}

func getMediaFullres(ctx *gin.Context) {
	getProcessedMedia(ctx, dataStore.Fullres)
}

func getMediaTypes(ctx *gin.Context) {
	typeMap := rc.MediaRepo.TypeService()
	ctx.JSON(http.StatusOK, typeMap)
}

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	upInfo, err := readCtxBody[newUploadBody](ctx)
	if err != nil {
		return
	}
	c := NewBufferedCaster()
	t := rc.TaskDispatcher.WriteToFile(upInfo.RootFolderId, upInfo.ChunkSize, upInfo.TotalUploadSize, c)
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
	uTask := rc.TaskDispatcher.GetWorkerPool().GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Status(http.StatusNotFound)
		return
	}

	parent := rc.FileTree.Get(newFInfo.ParentFolderId)
	if parent == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	newName := dataStore.MakeUniqueChildName(parent, newFInfo.NewFileName)

	newF, err := rc.FileTree.Touch(parent, newName, true, nil, rc.Caster)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = uTask.NewFileInStream(newF, newFInfo.FileSize)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newF.ID()})
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	uploadId := types.TaskId(ctx.Param("uploadId"))

	t := rc.TaskDispatcher.GetWorkerPool().GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := types.FileId(ctx.Param("fileId"))

	// We are about to read from the client, which could take a while.
	// Since we actually got this request, we know the client is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := util.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
	if err != nil {
		util.ShowErr(err)
		err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		if err != nil {
			util.ShowErr(err)
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
	}

	err = t.AddChunkToStream(fileId, chunk, ctx.GetHeader("Content-Range"))
	if err != nil {
		util.ShowErr(err)
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
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	folders := util.Map(folderIds, func(fId types.FileId) types.WeblensFile {
		return rc.FileTree.Get(fId)
	})

	ctx.JSON(http.StatusOK, gin.H{"medias": dataStore.RecursiveGetMedia(folders...)})
}

func searchFolder(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	folderId := types.FileId(ctx.Param("folderId"))
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir := rc.FileTree.Get(folderId)
	if dir == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !dir.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search must performed on a folder, not a regular file"})
		return
	}

	acc := dataStore.NewAccessMeta(user, rc.FileTree)
	if !dataStore.CanAccessFile(dir, acc) {
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
	err = dir.RecursiveMap(func(w types.WeblensFile) error {
		if r.MatchString(w.Filename()) {
			if w.Filename() == ".user_trash" {
				return nil
			}
			files = append(files, w)
		}
		return nil
	})
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	filesData := util.Map(files, func(w types.WeblensFile) types.FileInfo {
		d, err := w.FormatFileInfo(acc, rc.MediaRepo)
		util.ErrTrace(err)
		return d
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesData})
}

func getFile(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		return
	}

	fileId := types.FileId(ctx.Param("fileId"))
	file := rc.FileTree.Get(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	acc := dataStore.NewAccessMeta(user, rc.FileTree)
	formattedInfo, err := file.FormatFileInfo(acc, rc.MediaRepo)
	if err != nil {
		util.ErrTrace(err)
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

	file := rc.FileTree.Get(fileId)
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
	t := rc.TaskDispatcher.MoveFile(fileId, updateInfo.NewParentId, updateInfo.NewName, caster)
	t.Wait()

	if t.ReadError() != nil {
		util.Error.Println(t.ReadError())
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	fileIds, err := readCtxBody[[]types.FileId](ctx)
	if err != nil {
		return
	}
	user := getUserFromCtx(ctx)

	caster := NewBufferedCaster()
	defer caster.Close()

	var failed []types.FileId

	acc := dataStore.NewAccessMeta(user, rc.FileTree)
	shareId := types.ShareId(ctx.Query("shareId"))
	if shareId != "" {
		share, err := dataStore.GetShare(shareId, dataStore.FileShare, rc.FileTree)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		err = acc.AddShare(share)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	for _, fileId := range fileIds {
		file := rc.FileTree.Get(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to trash"})
			return
		}
		if user != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot trash file"})
			return
		}
		// oldFile := file.Copy()
		err := dataStore.MoveFileToTrash(file, acc, caster)
		if err != nil {
			util.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
		// err = dataStore.ResizeUp(oldFile.GetParent(), caster)
		// if err != nil {
		// 	util.ShowErr(err)
		// 	ctx.Status(http.StatusNotFound)
		// 	return
		// }
		// err = dataStore.ResizeUp(file.GetParent(), caster)
		// if err != nil {
		// 	util.ShowErr(err)
		// 	ctx.Status(http.StatusNotFound)
		// 	return
		// }

	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func deleteFiles(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	fileIds, err := readCtxBody[[]types.FileId](ctx)
	if err != nil {
		return
	}
	var failed []types.FileId

	caster := NewBufferedCaster()
	defer caster.Close()

	for _, fileId := range fileIds {
		file := rc.FileTree.Get(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		}
		if user != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot delete file"})
			return
		}

		err := dataStore.PermanentlyDeleteFile(file, caster)
		if err != nil {
			util.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}

		// err = dataStore.ResizeUp(file.GetParent(), caster)
		// if err != nil {
		// 	util.ShowErr(err)
		// 	return
		// }
	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func unTrashFiles(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []types.FileId
	err = json.Unmarshal(bodyBytes, &fileIds)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		util.ShowErr(err)
		return
	}

	caster := NewBufferedCaster()
	defer caster.Close()

	var failed []types.FileId

	for _, fileId := range fileIds {
		file := rc.FileTree.Get(fileId)
		if file == nil {
			util.ErrTrace(dataStore.ErrNoFile)
			failed = append(failed, fileId)
			continue
		}
		if user != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot un-trash file"})
			return
		}

		err = dataStore.ReturnFileFromTrash(file, caster)
		util.ErrTrace(err)
	}

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func createTakeout(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	shareId := types.ShareId(ctx.Query("shareId"))

	takeoutRequest, err := readCtxBody[takeoutFiles](ctx)
	if err != nil {
		return
	}
	if len(takeoutRequest.FileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot takeout 0 files"})
		return
	}

	files := util.Map(takeoutRequest.FileIds, func(fileId types.FileId) types.WeblensFile { return rc.FileTree.Get(fileId) })
	for _, file := range files {
		file.GetAbsPath() // Make sure directories have trailing slash

		acc := dataStore.NewAccessMeta(user, rc.FileTree).AddShareId(shareId, dataStore.FileShare)
		if file == types.WeblensFile(nil) || !dataStore.CanAccessFile(file, acc) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 file, and it is not a directory, we should have requested to just download that file
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := NewCaster()
	t := rc.TaskDispatcher.CreateZip(files, user.GetUsername(), shareId, rc.FileTree, caster)

	completed, status := t.Status()
	if completed && status == dataProcess.TaskSuccess {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId()})
	}
}

func downloadFile(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	fileId := types.FileId(ctx.Query("fileId"))
	shareId := types.ShareId(ctx.Param("shareId"))

	file := rc.FileTree.Get(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	acc := dataStore.NewAccessMeta(user, rc.FileTree).AddShareId(shareId, dataStore.FileShare)
	if !dataStore.CanAccessFile(file, acc) {
		util.Debug.Println("No auth")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	ctx.File(file.GetAbsPath())
}

func createUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var userInfo newUserBody
	err = json.Unmarshal(jsonData, &userInfo)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = dataStore.CreateUser(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate, rc.FileTree)
	if err != nil {
		if errors.Is(err, dataStore.ErrUserAlreadyExists) {
			ctx.Status(http.StatusConflict)
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}

func loginUser(ctx *gin.Context) {
	userCredentials, err := readCtxBody[loginBody](ctx)
	if err != nil {
		return
	}

	u := dataStore.GetUser(userCredentials.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if dataStore.CheckLogin(u, userCredentials.Password) {
		util.Info.Printf("Valid login for [%s]\n", userCredentials.Username)

		if token := u.GetToken(); token == "" {
			ctx.Status(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, tokenReturn{Token: token})
		}
	} else {
		ctx.AbortWithStatus(http.StatusNotFound)
	}

}

func getUserInfo(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		si := dataStore.GetServerInfo()
		if si == nil {
			ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
			return
		}
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, user)

}

func getUsers(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Store.GetUsers())
}

func updateUserPassword(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
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
	err = dataStore.UpdatePassword(user.GetUsername(), passUpd.OldPass, passUpd.NewPass)
	if err != nil {
		util.ShowErr(err)
		switch {
		case errors.Is(err.(error), dataStore.ErrBadPassword):
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

	err = dataStore.UpdateAdmin(username, update.Admin)
	if err != nil {
		if errors.Is(err, dataStore.ErrNoUser) {
			ctx.Status(http.StatusNotFound)
			return
		}
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func activateUser(ctx *gin.Context) {
	username := types.Username(ctx.Param("username"))

	if err := dataStore.ActivateUser(username, rc.FileTree); err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	// User to delete username
	// *cannot* use getUserFromCtx() here because that
	// will grab the user making the request, not the
	// username from the Param  \/
	user := dataStore.GetUser(types.Username(ctx.Param("username")))
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user with given username does not exist"})
		return
	}
	dataStore.DeleteUser(user)

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	ctx.Status(http.StatusNotImplemented)
	return

	user := getUserFromCtx(ctx)
	if !user.IsAdmin() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	// media.ClearCache()
}

func searchUsers(ctx *gin.Context) {
	filter := ctx.Query("filter")
	if len(filter) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autocomplete must contain at least 2 characters"})
		return
	}

	db := dataStore.NewDB()
	users := db.SearchUsers(filter)
	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

func getSharedFiles(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shares := dataStore.GetSharedWithUser(user)

	acc := dataStore.NewAccessMeta(user, rc.FileTree)
	filesInfos := util.Map(shares, func(sh types.Share) types.FileInfo {
		err := acc.AddShare(sh)
		acc.SetUsingShare(sh)
		if err != nil {
			util.ShowErr(err)
		}
		f := rc.FileTree.Get(types.FileId(sh.GetContentId()))
		fileInfo, err := f.FormatFileInfo(acc, rc.MediaRepo)
		if err != nil {
			util.ShowErr(err)
		}
		return fileInfo
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}

func createFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shareInfo, err := readCtxBody[newShareBody](ctx)
	if err != nil {
		return
	}
	if len(shareInfo.Users) != 0 && shareInfo.Public {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot create public share and specify users"})
		return
	}

	f := rc.FileTree.Get(shareInfo.FileIds[0])
	newShare, err := dataStore.CreateFileShare(f, user.GetUsername(), shareInfo.Users, shareInfo.Public, shareInfo.Wormhole)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"shareData": newShare})
}

func deleteShare(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))

	s, err := dataStore.GetShare(shareId, dataStore.FileShare, rc.FileTree)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	err = dataStore.DeleteShare(s, rc.FileTree)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func updateFileShare(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))
	share, err := dataStore.GetShare(shareId, dataStore.FileShare, rc.FileTree)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var shareInfo newShareBody
	err = json.Unmarshal(body, &shareInfo)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	share.SetAccessors(shareInfo.Users)
	share.SetPublic(shareInfo.Public)
	err = dataStore.UpdateFileShare(share, rc.FileTree)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)

}

func getFilesShares(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	file := rc.FileTree.Get(fileId)
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, file.GetShares())
}

func getFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)

	shareId := types.ShareId(ctx.Param("shareId"))
	share, err := dataStore.GetShare(shareId, dataStore.FileShare, rc.FileTree)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	acc := dataStore.NewAccessMeta(user, rc.FileTree)
	err = acc.AddShare(share)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, share)
}

func newApiKey(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	acc := dataStore.NewAccessMeta(user, rc.FileTree).SetRequestMode(dataStore.ApiKeyCreate)
	newKey, err := dataStore.GenerateApiKey(acc)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"key": newKey})
}

func getApiKeys(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	acc := dataStore.NewAccessMeta(user, rc.FileTree).SetRequestMode(dataStore.ApiKeyGet)
	keys, err := dataStore.GetApiKeys(acc)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
}

func deleteApiKey(ctx *gin.Context) {
	keyBody, err := readCtxBody[deleteKeyBody](ctx)
	if err != nil {
		return
	}
	dataStore.DeleteApiKey(keyBody.Key)
}

func getFolderStats(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("folderId"))
	rootFolder := rc.FileTree.Get(fileId)
	if rootFolder == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !rootFolder.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	t := rc.TaskDispatcher.GatherFsStats(rootFolder, rc.Caster)
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
	if dataStore.GetOwner() != nil && dataStore.GetServerInfo().ServerRole() != types.Initialization {
		ctx.Status(http.StatusNotFound)
		return
	}

	si, err := readCtxBody[initServer](ctx)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	if si.Role == types.Core {
		err := dataStore.InitServerCore(si.Name, si.Username, si.Password, rc.FileTree)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		ctx.Status(http.StatusCreated)

	} else if si.Role == types.Backup {
		rq := NewRequester()
		err := dataStore.InitServerForBackup(si.Name, si.CoreAddress, si.CoreKey, rq)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
	}

	ctx.Status(http.StatusCreated)

	store := GetStore()
	err = store.LoadUsers(rc.FileTree)
	if err != nil {
		util.ShowErr(err)
		return
	}

	util.Info.Println("Initialization succeeded. Restarting router...")
	go DoRoutes()
}

func getServerInfo(ctx *gin.Context) {
	si := dataStore.GetServerInfo()
	if si.ServerRole() == types.Initialization {
		ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"info": si})
}
