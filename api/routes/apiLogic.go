package routes

import (
	"encoding/json"
	"fmt"
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

func readRequestBody(ctx *gin.Context, obj any) (err error) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, obj)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is not in expected JSON format"})
		return
	}

	return
}

func getUserFromCtx(ctx *gin.Context) types.User {
	return dataStore.GetUser(types.Username(ctx.GetString("username")))
}

/* ================ */

func getMediaBatch(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"

	albumFilter := []types.AlbumId{}
	json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)

	db := dataStore.NewDB()

	media, err := db.GetFilteredMedia(sort, types.Username(ctx.GetString("username")), -1, albumFilter, raw)
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
	if !dataStore.CanUserAccessShare(share, types.Username(ctx.GetString("username"))) {
		ctx.Status(http.StatusNotFound)
		return
	}

	newUploadTask(ctx)
}

// Create new file upload task, and wait for data
func newUploadTask(ctx *gin.Context) {
	var upInfo newUploadInfo
	readRequestBody(ctx, &upInfo)
	c := NewBufferedCaster()
	c.Enable()
	t := UploadTasker.WriteToFile(upInfo.RootFolderId, upInfo.ChunkSize, upInfo.TotalUploadSize, c)
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newFileUpload(ctx *gin.Context) {
	uploadTaskId := types.TaskId(ctx.Param("uploadId"))
	var newFInfo newFileInfo
	readRequestBody(ctx, &newFInfo)

	handleNewFile(uploadTaskId, newFInfo, ctx)
}

func handleNewFile(uploadTaskId types.TaskId, newFInfo newFileInfo, ctx *gin.Context) {
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

	// newF.AddTask(uTask)
	uTask.AddChunkToStream(newF.Id(), []byte{}, "0-0/"+fmt.Sprint(newFInfo.FileSize))

	ctx.JSON(http.StatusCreated, gin.H{"fileId": newF.Id()})
}

func newSharedFileUpload(ctx *gin.Context) {
	shareId := types.ShareId(ctx.Param("shareId"))

	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if !dataStore.CanUserAccessShare(share, types.Username(ctx.GetString("username"))) {
		ctx.Status(http.StatusNotFound)
		return
	}

	var newFInfo newFileInfo
	readRequestBody(ctx, &newFInfo)

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
	newDir, err := dataStore.MkDir(parentFolder, ctx.Query("folderName"))
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})

	Caster.PushFileCreate(newDir)
	Caster.Flush()
}

func pubMakeDir(ctx *gin.Context) {
	share, err := dataStore.GetShare(types.ShareId(ctx.Param("shareId")), dataStore.FileShare)
	if err != nil || !dataStore.CanUserAccessShare(share, types.Username(ctx.GetString("username"))) {
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


	acc := dataStore.NewAccessMeta(user.GetUsername())
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
	dir.RecursiveMap(func(w types.WeblensFile) {
		if r.MatchString(w.Filename()) {
			if w.Filename() == ".user_trash" {
				return
			}
			files = append(files, w)
		}
	})

	filesData := util.Map(files, func(w types.WeblensFile) types.FileInfo {
		d, err := w.FormatFileInfo(acc)
		util.ErrTrace(err)
		return d
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesData})
}

func getUserTrashInfo(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		return
	}

	trash := dataStore.GetUserTrashDir(user.GetUsername())
	if trash == nil {
		util.Error.Println("Could not get trash directory for ", user.GetUsername())
		return
	}
	acc := dataStore.NewAccessMeta(user.GetUsername())
	formatRespondFolderInfo(trash, acc, ctx)
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

	acc := dataStore.NewAccessMeta(user.GetUsername())
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
	var updateInfo fileUpdateInfo
	readRequestBody(ctx, &updateInfo)

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

	t := dataProcess.GetGlobalQueue().MoveFile(fileId, types.FileId(updateInfo.NewParentId), updateInfo.NewName, Caster)
	t.Wait()
	Caster.Flush()

	if t.ReadError() != nil {
		util.Error.Println(t.ReadError())
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func updateFiles(ctx *gin.Context) {
	var filesData updateMany
	err := readRequestBody(ctx, &filesData)
	if err != nil {
		return
	}

	tp := dataProcess.NewTaskPool(false, nil)

	for _, fileId := range filesData.Files {
		tp.MoveFile(fileId, filesData.NewParentId, "", Caster)
	}
	tp.SignalAllQueued()
	tp.Wait(false)
	Caster.Flush()

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	var fileIds []types.FileId
	err := readRequestBody(ctx, &fileIds)
	if err != nil {
		return
	}
	user := getUserFromCtx(ctx)

	caster := NewBufferedCaster()
	caster.AutoflushEnable()

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

	caster.Flush()

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func deleteFiles(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	var fileIds []types.FileId
	err := readRequestBody(ctx, &fileIds)
	if err != nil {
		return
	}

	var failed []types.FileId

	caster := NewBufferedCaster()
	caster.AutoflushEnable()

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

		err := dataStore.PermenantlyDeleteFile(file, caster)
		if err != nil {
			util.ErrTrace(err)
			failed = append(failed, fileId)
			continue
		}
		dataStore.ResizeUp(file.GetParent(), caster)
	}

	caster.Flush()

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func unTrashFiles(ctx *gin.Context) {
	user := dataStore.GetUser(types.Username(ctx.GetString("username")))
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []types.FileId
	json.Unmarshal(bodyBytes, &fileIds)

	caster := NewBufferedCaster()
	caster.AutoflushEnable()

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

	caster.Flush()

	if len(failed) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"failedIds": failed})
		return
	}
	ctx.Status(http.StatusOK)
}

func createTakeout(ctx *gin.Context) {
	username := types.Username(ctx.GetString("username"))
	shareId := types.ShareId(ctx.Query("shareId"))

	var takeoutRequest takeoutFiles
	err := readRequestBody(ctx, &takeoutRequest)
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

		acc := dataStore.NewAccessMeta(username).AddShareId(shareId, dataStore.FileShare)
		if file == nil || !dataStore.CanAccessFile(file, acc) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 file, and it is not a directory, we should have requested to just download that file
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Single non-directory file should not be zipped"})
		return
	}

	caster := NewCaster(username)
	caster.Enable()
	t := dataProcess.GetGlobalQueue().CreateZip(files, username, shareId, caster)

	completed, _ := t.Status()
	if completed {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId()})
	}
}

func downloadFile(ctx *gin.Context) {
	fileId := types.FileId(ctx.Query("fileId"))
	username := types.Username(ctx.GetString("username"))
	shareId := types.ShareId(ctx.Param("shareId"))

	file := dataStore.FsTreeGet(fileId)
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	acc := dataStore.NewAccessMeta(username).AddShareId(shareId, dataStore.FileShare)
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

	var userInfo newUserInfo
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
	var usrCreds loginInfo
	err := readRequestBody(ctx, &usrCreds)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	u := dataStore.GetUser(usrCreds.Username)

	if dataStore.CheckLogin(u, usrCreds.Password) {
		util.Info.Printf("Valid login for [%s]\n", usrCreds.Username)

		if token := u.GetToken(); token == "" {
			ctx.Status(http.StatusInternalServerError)
		} else {
			ctx.JSON(http.StatusOK, tokenReturn{Token: token})
		}
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

}

func getUserInfo(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, user)

}

func getUsers(ctx *gin.Context) {
	users := dataStore.GetUsers()
	ctx.JSON(http.StatusOK, users)
}

func updateUserPassword(ctx *gin.Context) {
	var passUpd passwordUpdateInfo
	if err := readRequestBody(ctx, &passUpd); err != nil {
		return
	}

	if passUpd.OldPass == "" || passUpd.NewPass == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both oldPassword and newPassword fields are required"})
		return
	}
	err := dataStore.UpdatePassword(types.Username(ctx.GetString("username")), passUpd.OldPass, passUpd.NewPass)
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
	var update newUserInfo
	readRequestBody(ctx, &update)

	username := types.Username(ctx.Param("username"))

	err := dataStore.UpdateAdmin(username, update.Admin)
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
	var username types.Username
	err := readRequestBody(ctx, &username)
	if err != nil {
		return
	}

	db := dataStore.NewDB()
	err = dataStore.CreateUserHomeDir(username)
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
	util.Each(cacheFiles, func(wf types.WeblensFile) { util.ErrTrace(dataStore.PermenantlyDeleteFile(wf)) })
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
	username := types.Username(ctx.GetString("username"))
	db := dataStore.NewDB()
	shares := db.GetSharedWith(username)

	acc := dataStore.NewAccessMeta(username)
	filesInfos := util.Map(shares, func(sh types.Share) types.FileInfo {
		acc.AddShare(sh)
		f := dataStore.FsTreeGet(types.FileId(sh.GetContentId()))
		fileInfo, _ := f.FormatFileInfo(acc)
		return fileInfo
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}

func cleanupMedias(ctx *gin.Context) {
	dataStore.CleanOrphanedMedias()
}

func createFileShare(ctx *gin.Context) {
	var shareInfo newShareInfo
	err := readRequestBody(ctx, &shareInfo)
	if err != nil {
		return
	}

	if len(shareInfo.Users) != 0 && shareInfo.Public {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot create public share and specify users"})
		return
	}

	f := dataStore.FsTreeGet(shareInfo.FileIds[0])
	newShare, err := dataStore.CreateFileShare(f, types.Username(ctx.GetString("username")), shareInfo.Users, shareInfo.Public, shareInfo.Wormhole)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"shareData": newShare})
	Caster.Flush()
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
	Caster.Flush()

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
	var shareInfo newShareInfo
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
	shareId := types.ShareId(ctx.Param("shareId"))
	share, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil || !dataStore.CanUserAccessShare(share, types.Username(ctx.GetString("username"))) {
		if err != nil && err != dataStore.ErrNoShare {
			util.ErrTrace(err)
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file share"})
		return
	}

	ctx.JSON(http.StatusOK, share)
}

func newApiKey(ctx *gin.Context) {
	username := types.Username(ctx.GetString("username"))
	acc := dataStore.NewAccessMeta(username).SetRequestMode(dataStore.ApiKeyCreate)
	newKey, err := dataStore.GenerateApiKey(acc)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"key": newKey})
}

func getApiKeys(ctx *gin.Context) {
	acc := dataStore.NewAccessMeta(types.Username(ctx.GetString("username"))).SetRequestMode(dataStore.ApiKeyGet)
	keys, err := dataStore.GetApiKeys(acc)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"keys": keys})
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
