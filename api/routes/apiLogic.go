package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

func getMediaBatch(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"

	albumFilter := []string{}
	json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)

	db := dataStore.NewDB()

	media, err := db.GetFilteredMedia(sort, ctx.GetString("username"), -1, albumFilter, raw)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": media})
}

func getProcessedMedia(ctx *gin.Context, q dataStore.Quality) {
	mediaId := ctx.Param("mediaId")

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
	bs, err := m.ReadDisplayable(q, pageNum)
	if err != nil {
		util.DisplayError(err)
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

func getMediaMeta(ctx *gin.Context) {
	mediaId := ctx.Param("mediaId")

	m, err := dataStore.MediaMapGet(mediaId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given Id not found"})
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func updateMedias(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read body of media update")

	var updateBody updateMediasBody
	err = json.Unmarshal(jsonData, &updateBody)
	util.FailOnError(err, "Failed to unmarshal body of media update")

	db := dataStore.NewDB()
	db.UpdateMediasById(updateBody.FileHashes, updateBody.Owner)

}

// Create new file upload task, and wait for data
func newFileUpload(ctx *gin.Context) {
	parentId := ctx.Query("parentFolderId")
	filename := ctx.Query("filename")
	parent := dataStore.FsTreeGet(parentId)
	if parent == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	children := parent.GetChildren()
	if slices.ContainsFunc(children, func(wf *dataStore.WeblensFile) bool { return wf.Filename() == filename }) {
		ctx.AbortWithStatus(http.StatusConflict)
		return
	}

	t := UploadTasker.WriteToFile(filename, parentId, Caster)
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

func newSharedFileUpload(ctx *gin.Context) {
	shareId := ctx.Query("shareId")
	parentFolderId := ctx.Query("parentFolderId")
	filename := ctx.Query("filename")

	share, err := dataStore.GetShare(shareId, dataStore.FileShare)

	if err == dataStore.ErrNoShare || !share.IsPublic() {
		ctx.Status(http.StatusNotFound)
	} else if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if parentFolderId == shareId {
		parentFolderId = share.GetContentId()
	}

	folder := dataStore.FsTreeGet(parentFolderId)
	if folder == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	children := folder.GetChildren()
	names := util.Map(children, func(f *dataStore.WeblensFile) string { return f.Filename() })
	slices.Sort(names)

	_, e := slices.BinarySearch(names, filename)
	var counter int = 1

	dotIndex := strings.LastIndex(filename, ".")
	nameOnly := filename[:dotIndex]
	ext := filename[dotIndex:]
	for e {
		tmpname := fmt.Sprintf("%s_%d%s", nameOnly, counter, ext)
		_, e = slices.BinarySearch(names, tmpname)
		if !e {
			filename = tmpname
			break
		}
		counter++
	}

	t := UploadTasker.WriteToFile(filename, parentFolderId, Caster)
	ctx.JSON(http.StatusCreated, gin.H{"uploadId": t.TaskId()})
}

// Add chunks of file to previously created task
func handleUploadChunk(ctx *gin.Context) {
	uploadId := ctx.Param("uploadId")
	t := dataProcess.GetTask(uploadId)
	if t == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No upload exists with given id"})
		return
	}
	// This request has come through, but might take time to read the body,
	// so we clear the timeout. The task must re-enable it when it goes to wait
	// for the next chunk
	t.ClearTimeout()

	stream := t.GetMeta().(dataProcess.WriteFileMeta).ChunkStream
	chunk, err := util.OracleReader(ctx.Request.Body, ctx.Request.ContentLength)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
	}
	chunkData := dataProcess.FileChunk{Chunk: chunk, ContentRange: ctx.GetHeader("Content-Range")}
	stream <- chunkData
}

func makeDir(ctx *gin.Context) {
	parentFolder := dataStore.FsTreeGet(ctx.Query("parentFolderId"))
	if parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}
	newDir, err := dataStore.MkDir(parentFolder, ctx.Query("folderName"))
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})

	Caster.PushFileCreate(newDir)
	Caster.Flush()
}

func pubMakeDir(ctx *gin.Context) {
	share, err := dataStore.GetShare(ctx.Query("shareId"), dataStore.FileShare)

	if err != nil || !CanUserAccessShare(share, ctx.GetString("username")) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	shareFolder := dataStore.FsTreeGet(share.GetContentId())
	parentFolderId := ctx.Query("parentFolderId")

	if parentFolderId == share.GetShareId() {
		parentFolderId = share.GetContentId()
	}

	parentFolder := dataStore.FsTreeGet(parentFolderId)
	if shareFolder == nil || parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if !strings.HasPrefix(parentFolder.String(), shareFolder.String()) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Folder is not child of share folder"})
		return
	}

	newDir, err := dataStore.MkDir(parentFolder, ctx.Query("folderName"))
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})

	Caster.PushFileCreate(newDir)
}

func _getDirInfo(dir *dataStore.WeblensFile, ctx *gin.Context) {
	username := ctx.GetString("username")

	selfData, err := dir.FormatFileInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Failed to get folder info")
		return
	}

	var filteredDirInfo []dataStore.FileInfo
	if dir.IsDir() {
		filteredDirInfo = dir.GetChildrenInfo()
		filteredDirInfo = util.Filter(filteredDirInfo, func(t dataStore.FileInfo) bool { return t.Id != "R" })
	}

	parentsInfo := []dataStore.FileInfo{}
	parent := dir.GetParent()
	if parent == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find parent directory"})
		return
	}
	for parent.Id() != "0" && CanUserAccessFile(parent.Id(), username, "") {
		parentInfo, err := parent.FormatFileInfo()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Failed to format parent file info")
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	packagedInfo := gin.H{"self": selfData, "children": filteredDirInfo, "parents": parentsInfo}
	ctx.JSON(http.StatusOK, packagedInfo)

}

func getFolderInfo(ctx *gin.Context) {
	folderId := ctx.Param("folderId")

	dir := dataStore.FsTreeGet(folderId)
	if dir == nil {
		util.Debug.Println("Actually not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	if dir.Id() == "" {
		util.Error.Println("Blank file descriptor trying to get folder info")
		ctx.Status(http.StatusNotFound)
		return
	}

	if !CanUserAccessFile(dir.Id(), ctx.GetString("username"), "") {
		util.Debug.Println("Not auth")
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	if !dir.IsDir() {
		dir = dir.GetParent()
	}

	_getDirInfo(dir, ctx)
}

func getFolderMedia(ctx *gin.Context) {
	folderId := ctx.Param("folderId")
	medias := dataStore.RecursiveGetMedia(folderId)

	ctx.JSON(http.StatusOK, gin.H{"medias": medias})
}

func getUserTrashInfo(ctx *gin.Context) {
	trash := dataStore.GetUserTrashDir(ctx.GetString("username"))
	if trash == nil {
		util.Error.Println("Could not get trash directory for ", ctx.GetString("username"))
		return
	}
	_getDirInfo(trash, ctx)
}

func getFile(ctx *gin.Context) {
	file := dataStore.FsTreeGet(ctx.Param("fileId"))
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := file.FormatFileInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

func updateFile(ctx *gin.Context) {
	fileId := ctx.Query("fileId")
	newParentId := ctx.Query("newParentId")
	newFilename := ctx.Query("newFilename")

	if fileId == "" {
		err := fmt.Errorf("fileId is required to update file")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	file := dataStore.FsTreeGet(fileId)
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	// If the directory does not change, just assume this is a rename
	if newParentId == "" {
		newParentId = file.GetParent().Id()
	}

	t := dataProcess.GetGlobalQueue().MoveFile(fileId, newParentId, newFilename, Caster)
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
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var filesData updateMany
	json.Unmarshal(jsonData, &filesData)

	wq := dataProcess.NewWorkQueue()

	for _, fileId := range filesData.Files {
		wq.MoveFile(fileId, filesData.NewParentId, "", Caster)
	}
	wq.SignalAllQueued()
	wq.Wait(false)
	Caster.Flush()

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []string
	json.Unmarshal(bodyBytes, &fileIds)

	var failed []string

	trashDir := dataStore.GetUserTrashDir(ctx.GetString("username"))

	caster := NewBufferedCaster()
	caster.AutoflushEnable()

	for _, fileId := range fileIds {
		file := dataStore.FsTreeGet(fileId)
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		}

		inTrash := false
		parent := file.GetParent()
		for parent.Id() != "0" {
			if parent.Id() == trashDir.Id() {
				inTrash = true
				break
			}
			parent = parent.GetParent()
		}

		if inTrash {
			err := dataStore.PermenantlyDeleteFile(file, caster)
			if err != nil {
				util.DisplayError(err)
				failed = append(failed, fileId)
				continue
				// ctx.AbortWithStatus(http.StatusInternalServerError)
				// return
			}
			dataStore.Resize(file.GetParent(), caster)
		} else {
			oldFile := file.Copy()
			err := dataStore.MoveFileToTrash(file, caster)
			if err != nil {
				util.DisplayError(err)
				failed = append(failed, fileId)
				continue
				// ctx.AbortWithStatus(http.StatusInternalServerError)
				// return
			}
			dataStore.Resize(oldFile.GetParent(), caster)
			dataStore.Resize(file.GetParent(), caster)
		}
	}

	caster.Flush()

	if len(failed) != 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"failedIds": failed})
	}
	ctx.Status(http.StatusOK)
}

func createTakeout(ctx *gin.Context) {
	username := ctx.GetString("username")
	shareId := ctx.Query("shareId")

	var takeoutRequest takeoutFiles
	bodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}

	err = json.Unmarshal(bodyData, &takeoutRequest)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request body not in expected format"})
		return
	}

	if len(takeoutRequest.FileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot takeout 0 files"})
		return
	}

	files := util.Map(takeoutRequest.FileIds, func(fileId string) *dataStore.WeblensFile { return dataStore.FsTreeGet(fileId) })
	for _, file := range files {
		_ = file.String() // Make sure directories have trailing slash

		if file == nil || !CanUserAccessFile(file.Id(), username, shareId) {
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
	if !CanUserAccessFile(ctx.Query("fileId"), ctx.GetString("username"), ctx.Query("shareId")) {
		util.Debug.Println("No auth")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	file := dataStore.FsTreeGet(ctx.Query("fileId"))
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	ctx.File(file.String())
}

func createUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var userInfo newUserInfo
	json.Unmarshal(jsonData, &userInfo)

	db := dataStore.NewDB()
	_, err = db.GetUser(userInfo.Username)
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	db.CreateUser(userInfo.Username, userInfo.Password, userInfo.Admin)

	ctx.Status(http.StatusCreated)
}

func loginUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var usrCreds loginInfo
	json.Unmarshal(jsonData, &usrCreds)

	db := dataStore.NewDB()
	if db.CheckLogin(usrCreds.Username, usrCreds.Password) {
		util.Info.Printf("Valid login for [%s]\n", usrCreds.Username)
		user, err := db.GetUser(usrCreds.Username)
		util.FailOnError(err, "Failed to get user to log in")

		if len(user.Tokens) != 0 {
			var ret tokenReturn = tokenReturn{Token: user.Tokens[0]}
			ctx.JSON(http.StatusOK, ret)
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		tokenString, err := token.SignedString([]byte("key"))
		if err != nil {
			util.Error.Println(err)
		}

		var ret tokenReturn = tokenReturn{Token: tokenString}
		db.AddTokenToUser(usrCreds.Username, tokenString)
		ctx.JSON(http.StatusOK, ret)
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

}

func getUserInfo(ctx *gin.Context) {
	db := dataStore.NewDB()
	authList := strings.Split(ctx.Request.Header["Authorization"][0], ",")

	user, err := db.GetUser(authList[0])
	util.FailOnError(err, "Failed to get user info")
	user.Password = ""

	var empty []string
	user.Tokens = empty

	ctx.JSON(http.StatusOK, gin.H{"username": user.Username, "homeFolderId": dataStore.GetUserHomeDir(user.Username).Id(), "trashFolderId": dataStore.GetUserTrashDir(user.Username).Id(), "admin": user.Admin})

}

func getUsers(ctx *gin.Context) {
	db := dataStore.NewDB()
	users := db.GetUsers()
	ctx.JSON(http.StatusOK, users)
}

func updateUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read request body to update user")

	var userToUpdate struct {
		Username string `json:"username"`
	}
	json.Unmarshal(jsonData, &userToUpdate)

	db := dataStore.NewDB()
	err = dataStore.CreateUserHomeDir(userToUpdate.Username)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	db.ActivateUser(userToUpdate.Username)

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := ctx.Param("username") // User to delete username
	homeDir := dataStore.GetUserHomeDir(username)
	dataStore.MoveFileToTrash(homeDir)

	db := dataStore.NewDB() // Admin username
	db.DeleteUser(username)

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	db := dataStore.NewDB()
	db.FlushRedis()
	dataProcess.FlushCompleteTasks()

	cacheFiles := dataStore.GetCacheDir().GetChildren()
	util.Each(cacheFiles, func(wf *dataStore.WeblensFile) { util.DisplayError(dataStore.PermenantlyDeleteFile(wf)) })
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
	username := ctx.GetString("username")
	db := dataStore.NewDB()
	sharedList := db.GetSharedWith(username)

	filesInfos := util.Map(sharedList, func(file *dataStore.WeblensFile) dataStore.FileInfo {
		fileInfo, _ := file.FormatFileInfo()
		return fileInfo
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}

func getFileTreeInfo(ctx *gin.Context) {
	util.Debug.Println("File tree size: ", dataStore.GetTreeSize())
}

func cleanupMedias(ctx *gin.Context) {
	dataStore.CleanOrphanedMedias()
}

func createFileShare(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var shareInfo newShareInfo
	err = json.Unmarshal(body, &shareInfo)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if len(shareInfo.Users) != 0 && shareInfo.Public {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot create public share and specify users"})
		return
	}

	f := dataStore.FsTreeGet(shareInfo.FileIds[0])
	newShare, err := dataStore.CreateFileShare(f, ctx.GetString("username"), shareInfo.Users, shareInfo.Public, shareInfo.Wormhole)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"shareData": newShare})
	Caster.Flush()
}

func deleteShare(ctx *gin.Context) {
	shareId := ctx.Param("shareId")

	s, err := dataStore.GetShare(shareId, dataStore.FileShare)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	err = dataStore.DeleteShare(s)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	Caster.Flush()

	ctx.Status(http.StatusOK)
}

func updateFileShare(ctx *gin.Context) {
	share, err := dataStore.GetShare(ctx.Param("shareId"), dataStore.FileShare)
	if err != nil {
		util.DisplayError(err)
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
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	share.AddAccessors(shareInfo.Users)
	share.SetPublic(shareInfo.Public)
	err = dataStore.UpdateFileShare(share)
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)

}

func getFilesShares(ctx *gin.Context) {
	file := dataStore.FsTreeGet(ctx.Param("fileId"))
	if file == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, file.GetShares())
}

func getFileShare(ctx *gin.Context) {
	share, err := dataStore.GetShare(ctx.Param("shareId"), dataStore.FileShare)
	if err != nil || !CanUserAccessShare(share, ctx.GetString("username")) {
		if err != dataStore.ErrNoShare {
			util.DisplayError(err)
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file share"})
		return
	}

	ctx.JSON(http.StatusOK, share)
}
