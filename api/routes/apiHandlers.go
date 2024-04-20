package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"

	"github.com/ethrousseau/weblens/api/util"

	"github.com/gin-gonic/gin"
)

func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
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

// func readRequestBody(ctx *gin.Context, obj any) (err error) {
// 	jsonData, err := io.ReadAll(ctx.Request.Body)
// 	if err != nil {
// 		util.ShowErr(err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
// 		return
// 	}
// 	err = json.Unmarshal(jsonData, obj)
// 	if err != nil {
// 		util.ShowErr(err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is not in expected JSON format"})
// 		return
// 	}

// 	return
// }

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

	albumFilter := []types.AlbumId{}
	err := json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	media, err := dataStore.GetFilteredMedia(user, sort, -1, albumFilter, raw)
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": media})
}

func getProcessedMedia(ctx *gin.Context, q types.Quality) {
	mediaId := types.MediaId(ctx.Param("mediaId"))

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

	m, err := dataStore.MediaMapGet(mediaId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given Id not found"})
		return
	}
	bs, err := m.ReadDisplayable(types.Quality(q), pageNum)
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fullres image"})
		return
	}
	ctx.Writer.Write(bs)
}

func getMediaThumbnail(ctx *gin.Context) {
	getProcessedMedia(ctx, dataStore.Thumbnail)
}

func getMediaFullres(ctx *gin.Context) {
	getProcessedMedia(ctx, dataStore.Fullres)
}

func getMediaTypes(ctx *gin.Context) {
	typeMap := dataStore.GetMediaTypeMap()
	// for _, t := range typeMap {

	// }
	ctx.JSON(http.StatusOK, typeMap)
}

func getMediaMeta(ctx *gin.Context) {
	mediaId := types.MediaId(ctx.Param("mediaId"))

	m, err := dataStore.MediaMapGet(mediaId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given Id not found"})
		return
	}

	ctx.JSON(http.StatusOK, m)
}

// func updateMedias(ctx *gin.Context) {
// 	jsonData, err := io.ReadAll(ctx.Request.Body)
// 	util.FailOnError(err, "Failed to read body of media update")

// 	var updateBody updateMediasBody
// 	err = json.Unmarshal(jsonData, &updateBody)
// 	util.FailOnError(err, "Failed to unmarshal body of media update")

// 	db := dataStore.NewDB()
// 	db.UpdateMediasById(updateBody.MediaIdes, updateBody.Owner)
// }

func newSharedUploadTask(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil {
		if err != dataStore.ErrNoShare {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		} else {
			ctx.Status(http.StatusNotFound)
			return
		}
	}
	if !dataStore.CanUserAccessShare(share, types.Username(user.GetUsername())) {
		ctx.Status(http.StatusNotFound)
		return
	}

	newUploadTask(ctx)
}

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	upInfo, err := readCtxBody[newUploadBody](ctx)
	if err != nil {
		return
	}
	c := NewBufferedCaster()
	t := UploadTasker.WriteToFile(upInfo.RootFolderId, upInfo.ChunkSize, upInfo.TotalUploadSize, c)
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
	uTask := dataProcess.GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	completed, _ := uTask.Status()
	if completed {
		ctx.Status(http.StatusNotFound)
		return
	}

	parent := dataStore.FsTreeGet(newFInfo.ParentFolderId)
	if parent == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	children := parent.GetChildren()
	if slices.ContainsFunc(children, func(wf types.WeblensFile) bool { return wf.Filename() == newFInfo.NewFileName }) {
		ctx.AbortWithStatus(http.StatusConflict)
		return
	}

	newF, err := dataStore.Touch(parent, newFInfo.NewFileName, true, VoidCaster)
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

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newF.Id()})
}

func newSharedFileUpload(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if !dataStore.CanUserAccessShare(share, types.Username(user.GetUsername())) {
		ctx.Status(http.StatusNotFound)
		return
	}

	newFInfo, err := readCtxBody[newFileBody](ctx)
	if err != nil {
		return
	}

	uploadTaskId := types.TaskId(ctx.Param("uploadId"))
	handleNewFile(uploadTaskId, newFInfo, ctx)
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	uploadId := types.TaskId(ctx.Param("uploadId"))
	t := dataProcess.GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}

	fileId := types.FileId(ctx.Param("fileId"))
	// f := dataStore.FsTreeGet(fileId)
	// if f == nil {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"error": "No file exists with given id"})
	// 	return
	// }

	// This request has come through, but might take time to read the body,
	// so we clear the timeout. The task must re-enable it when it goes to wait
	// for the next chunk
	// t.ClearTimeout()

	// start := time.Now()
	chunk, err := util.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
	if err != nil {
		util.ErrTrace(err)
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

func makeDir(ctx *gin.Context) {
	pfId := types.FileId(ctx.Query("parentFolderId"))
	parentFolder := dataStore.FsTreeGet(pfId)
	if parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}

	caster := NewBufferedCaster()
	defer caster.Close()
	newDir, err := dataStore.MkDir(parentFolder, ctx.Query("folderName"), caster)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})
}

func pubMakeDir(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	share, err := dataStore.GetShare(types.ShareId(ctx.Param("shareId")), dataStore.FileShare)
	if err != nil || !dataStore.CanUserAccessShare(share, types.Username(user.GetUsername())) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	shareFolder := dataStore.FsTreeGet(types.FileId(share.GetContentId()))
	parentFolderId := types.FileId(ctx.Query("parentFolderId"))

	if parentFolderId.String() == share.GetShareId().String() {
		parentFolderId = types.FileId(share.GetContentId())
	}

	parentFolder := dataStore.FsTreeGet(parentFolderId)
	if shareFolder == nil || parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if !strings.HasPrefix(parentFolder.GetAbsPath(), shareFolder.GetAbsPath()) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Folder is not child of share folder"})
		return
	}

	newDir, err := dataStore.MkDir(parentFolder, ctx.Query("folderName"))
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})

	Caster.PushFileCreate(newDir)
}

func getFolderMedia(ctx *gin.Context) {
	folderId := types.FileId(ctx.Param("folderId"))
	medias := dataStore.RecursiveGetMedia(folderId)

	ctx.JSON(http.StatusOK, gin.H{"medias": medias})
}

func searchFolder(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	folderId := types.FileId(ctx.Param("folderId"))
	searchStr := ctx.Query("search")
	filterStr := ctx.Query("filter")

	dir := dataStore.FsTreeGet(folderId)
	if dir == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !dir.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search must performed on a folder, not a regular file"})
		return
	}

	acc := dataStore.NewAccessMeta(user)
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

	files := []types.WeblensFile{}
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
		d, err := w.FormatFileInfo(acc)
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
	file := dataStore.FsTreeGet(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	acc := dataStore.NewAccessMeta(user)
	formattedInfo, err := file.FormatFileInfo(acc)
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

	file := dataStore.FsTreeGet(fileId)
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
		updateInfo.NewParentId = file.GetParent().Id()
	}

	caster := NewBufferedCaster()
	defer caster.Close()
	t := dataProcess.GetGlobalQueue().MoveFile(fileId, types.FileId(updateInfo.NewParentId), updateInfo.NewName, caster)
	t.Wait()

	if t.ReadError() != nil {
		util.Error.Println(t.ReadError())
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func updateFiles(ctx *gin.Context) {
	filesData, err := readCtxBody[updateMany](ctx)
	if err != nil {
		return
	}

	tp := dataProcess.NewTaskPool(false, nil)

	caster := NewBufferedCaster()
	defer caster.Close()

	for _, fileId := range filesData.Files {
		tp.MoveFile(fileId, filesData.NewParentId, "", caster)
	}
	tp.SignalAllQueued()
	tp.Wait(false)

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

	for _, fileId := range fileIds {
		file := dataStore.FsTreeGet(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to trash"})
			return
		}
		if user != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot trash file"})
			return
		}
		oldFile := file.Copy()
		err := dataStore.MoveFileToTrash(file, caster)
		if err != nil {
			util.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
		dataStore.ResizeUp(oldFile.GetParent(), caster)
		dataStore.ResizeUp(file.GetParent(), caster)

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
		file := dataStore.FsTreeGet(fileId)
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
		dataStore.ResizeUp(file.GetParent(), caster)
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
	json.Unmarshal(bodyBytes, &fileIds)

	caster := NewBufferedCaster()
	defer caster.Close()

	var failed []types.FileId

	for _, fileId := range fileIds {
		file := dataStore.FsTreeGet(fileId)
		if file == nil {
			util.ErrTrace(dataStore.ErrNoFile)
			failed = append(failed, fileId)
			continue
		}
		if user != file.Owner() {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Non-owner cannot untrash file"})
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

	files := util.Map(takeoutRequest.FileIds, func(fileId types.FileId) types.WeblensFile { return dataStore.FsTreeGet(fileId) })
	for _, file := range files {
		_ = file.GetAbsPath() // Make sure directories have trailing slash

		acc := dataStore.NewAccessMeta(user).AddShareId(shareId, dataStore.FileShare)
		if file == nil || !dataStore.CanAccessFile(file, acc) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 file, and it is not a directory, we should have requested to just download that file
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := NewCaster(user.GetUsername())
	caster.Enable()
	t := dataProcess.GetGlobalQueue().CreateZip(files, user.GetUsername(), shareId, caster)

	completed, _ := t.Status()
	if completed {
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

	file := dataStore.FsTreeGet(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	acc := dataStore.NewAccessMeta(user).AddShareId(shareId, dataStore.FileShare)
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
	json.Unmarshal(jsonData, &userInfo)

	err = dataStore.CreateUser(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate)
	if err != nil {
		if err == dataStore.ErrUserAlreadyExists {
			ctx.Status(http.StatusConflict)
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}

func loginUser(ctx *gin.Context) {
	usrCreds, err := readCtxBody[loginBody](ctx)
	if err != nil {
		return
	}

	u := dataStore.GetUser(usrCreds.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if dataStore.CheckLogin(u, usrCreds.Password) {
		util.Info.Printf("Valid login for [%s]\n", usrCreds.Username)

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
		switch err {
		case dataStore.ErrBadPassword:
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
		if err == dataStore.ErrNoUser {
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
	username, err := readCtxBody[types.Username](ctx)
	if err != nil {
		return
	}
	db := dataStore.NewDB()
	_, err = dataStore.CreateUserHomeDir(username) //TODO
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	db.ActivateUser(username)

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
	db := dataStore.NewDB()
	db.FlushRedis()
	dataProcess.FlushCompleteTasks()

	cacheFiles := dataStore.GetCacheDir().GetChildren()
	util.Each(cacheFiles, func(wf types.WeblensFile) { util.ErrTrace(dataStore.PermanentlyDeleteFile(wf)) })
}

func searchUsers(ctx *gin.Context) {
	filter := ctx.Query("filter")
	if len(filter) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autcomplete must contain at least 2 characters"})
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
	db := dataStore.NewDB()
	shares := db.GetSharedWith(user.GetUsername())

	acc := dataStore.NewAccessMeta(user)
	filesInfos := util.Map(shares, func(sh types.Share) types.FileInfo {
		acc.AddShare(sh)
		f := dataStore.FsTreeGet(types.FileId(sh.GetContentId()))
		fileInfo, _ := f.FormatFileInfo(acc)
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

	f := dataStore.FsTreeGet(shareInfo.FileIds[0])
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

	s, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	err = dataStore.DeleteShare(s)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func updateFileShare(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))
	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
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
	err = dataStore.UpdateFileShare(share)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)

}

func getFilesShares(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	file := dataStore.FsTreeGet(fileId)
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, file.GetShares())
}

func getFileShare(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shareId := types.ShareId(ctx.Param("shareId"))
	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil || !dataStore.CanUserAccessShare(share, user.GetUsername()) {
		if err != nil && err != dataStore.ErrNoShare {
			util.ErrTrace(err)
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file share"})
		return
	}

	ctx.JSON(http.StatusOK, share)
}

func newApiKey(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	acc := dataStore.NewAccessMeta(user).SetRequestMode(dataStore.ApiKeyCreate)
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
	acc := dataStore.NewAccessMeta(user).SetRequestMode(dataStore.ApiKeyGet)
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
	rootFolder := dataStore.FsTreeGet(fileId)
	if rootFolder == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !rootFolder.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	t := dataProcess.GetGlobalQueue().GatherFsStats(rootFolder, VoidCaster)
	t.Wait()
	res := t.GetResult("sizesByExtension")

	ctx.JSON(http.StatusOK, res)
}

func getRandomMedias(ctx *gin.Context) {
	numStr := ctx.Query("count")
	numPhotos, err := strconv.Atoi(numStr)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	media := dataStore.GetRandomMedia(numPhotos)
	ctx.JSON(http.StatusOK, gin.H{"medias": media})
}

func initializeServer(ctx *gin.Context) {
	// Can't init server if already initialized
	if dataStore.GetOwner() != nil && dataStore.GetServerInfo() != nil {
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
		err := dataStore.InitServerCore(si.Name, si.Username, si.Password)
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
	store.LoadUsers()

	util.Debug.Println("Initilization succeeded. Restarting router...")
	go DoRoutes()

}

func getServerInfo(ctx *gin.Context) {
	si := dataStore.GetServerInfo()
	if si == nil {
		ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"info": si})
}
