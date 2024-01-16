package routes

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

type loginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getMediaBatch(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"

	albumFilter := []string{}
	json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)

	db := dataStore.NewDB(ctx.GetString("username"))

	media, err := db.GetFilteredMedia(sort, ctx.GetString("username"), -1, albumFilter, raw, false)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": media})
}

func getMediaItem(ctx *gin.Context) {
	fileHash := ctx.Param("filehash")
	_, err := b64.URLEncoding.DecodeString(fileHash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.DisplayError(err, "Given filehash ("+fileHash+") is not base64 encoded")
		return
	}

	includeMeta := util.BoolFromString(ctx.Query("meta"), true)
	includeThumbnail := util.BoolFromString(ctx.Query("thumbnail"), true)
	includeFullres := util.BoolFromString(ctx.Query("fullres"), true)

	if !(includeMeta || includeThumbnail || includeFullres) {
		// At least one option must be selected
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one of meta, thumbnail, or fullres must be selected"})
		return
	} else if includeFullres && (includeMeta || includeThumbnail) {
		// Full res must be the only option if selected
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "fullres should be the only option if selected"})
		return
	}

	db := dataStore.NewDB(ctx.GetString("username"))
	m := db.GetMedia(fileHash, includeThumbnail)

	filled, reason := m.IsFilledOut(!includeThumbnail)
	if !filled {
		util.Error.Printf("Failed to get [ %s ] from Database (missing %s)", fileHash, reason)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if includeFullres {
		fullresBytes, err := m.ReadFullres(db)
		util.FailOnError(err, "Failed to read fullres file")
		ctx.Writer.Write(fullresBytes)
		return

	} else if !includeMeta && includeThumbnail {
		thumbBytes, err := b64.StdEncoding.DecodeString(m.Thumbnail64)
		util.FailOnError(err, "Failed to decode thumb64 to bytes")

		ctx.Writer.Write(thumbBytes)
	} else {
		ctx.JSON(http.StatusOK, m)
	}

}

type updateMediaItemsBody struct {
	Owner      string   `json:"owner"`
	FileHashes []string `json:"fileHashes"`
}

func updateMediaItems(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read body of media item update")

	var updateBody updateMediaItemsBody
	err = json.Unmarshal(jsonData, &updateBody)
	util.FailOnError(err, "Failed to unmarshal body of media item update")

	db := dataStore.NewDB(ctx.GetString("username"))
	db.UpdateMediasByFilehash(updateBody.FileHashes, updateBody.Owner)

}

func streamVideo(ctx *gin.Context) {
	fileHash := ctx.Param("filehash")
	_, err := b64.URLEncoding.DecodeString(fileHash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.DisplayError(err, "Given filehash ("+fileHash+") is not base64 encoded")
		return
	}

	includeThumbnail := false

	db := dataStore.NewDB(ctx.GetString("username"))
	m := db.GetMedia(fileHash, includeThumbnail)

	filled, reason := m.IsFilledOut(!includeThumbnail)
	if !filled {
		util.Error.Printf("Failed to get [ %s ] from Database (missing %s)", fileHash, reason)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-i", dataStore.GuaranteeAbsolutePath(m.Filepath), "-c:v", "h264_videotoolbox", "-f", "h264", "pipe:1")
	// stdout, err := cmd.StdoutPipe()
	// util.FailOnError(err, "Failed to get ffmpeg stdout pipe")

	// err = cmd.Start()
	// if err != nil {
	// 	util.FailOnError(err, "Failed to start ffmpeg")
	// }
	// // buf := new (bytes.Buffer)
	// // _, err = buf.ReadFrom(stdout)

	//	util.FailOnError(err, "Failed to run ffmpeg to get video thumbnail")
	// file, err := os.Open(dataStore.GuaranteeAbsolutePath(m.Filepath, ctx.GetString("username")))

	util.FailOnError(err, "Failed to open fullres stream file")

	ctx.Writer.Header().Add("Connection", "keep-alive")

	//writtenBytes, err := io.Copy(ctx.Writer, stdout)
	// writtenBytes, err := io.Copy(ctx.Writer, file)
	// util.FailOnError(err, fmt.Sprintf("Failed to write video stream to response writer (wrote %d bytes)", writtenBytes))

	//cmd.Wait()
}

func uploadItem(file *dataStore.WeblensFileDescriptor, item64, uploaderName string) error {
	if !file.Exists() {
		conflictErr := fmt.Errorf("file (%s) does not exist for writing", file)
		return conflictErr
	}

	index := strings.Index(item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item64[index+1:])
	if err != nil {
		return err
	}

	err = file.Write(itemBytes)
	if err != nil {
		return err
	}

	return nil
}

type uploadedFile struct {
	File64         string `json:"file64"`
	FileName       string `json:"fileName"`
	ParentFolderId string `json:"parentFolderId"`
}

func handleChunkedFileUpload(ctx *gin.Context) {
	chunk64 := ctx.Request.FormValue("chunk")
	filename := ctx.Request.FormValue("filename")
	uploadId := ctx.Request.FormValue("uploadId")

	contentRangeHeader := ctx.GetHeader("Content-Range")
	rangeAndSize := strings.Split(contentRangeHeader, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	rangeMin, err := strconv.Atoi(rangeParts[0])
	if err != nil {
		util.Debug.Println(rangeParts[0])
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing range in Content-Range header"})
		return
	}
	rangeMax, err := strconv.Atoi(rangeParts[1])
	if err != nil {
		util.Debug.Println(rangeParts[1])
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing range in Content-Range header"})
		return
	}
	chunkSize := rangeMax - rangeMin
	if chunkSize == 0 {
		util.Error.Println("Chunk is empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Chunk is empty"})
		return
	}

	fileSize, err := strconv.Atoi(rangeAndSize[1])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing file size in Content-Range header"})
		return
	}

	var tmpFile *dataStore.WeblensFileDescriptor

	if uploadId == "" {
		tmpDir := dataStore.GetTmpDir()
		finalParent := dataStore.FsTreeGet(ctx.Request.FormValue("parentFolderId"))
		if finalParent == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not get parent directory"})
			return
		}

		tmpFilename := util.HashOfString(8, finalParent.GetParent().Id()+filename+time.Now().String())
		tmpFile, err = dataStore.Touch(tmpDir, tmpFilename, true)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not touch temp file"})
			return
		}
		uploadId = tmpFile.Id()
	} else {
		tmpFile = dataStore.FsTreeGet(uploadId)
	}

	f, err := os.OpenFile(tmpFile.String(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating file"})
		return
	}

	commaPos := strings.Index(chunk64, ",")
	if commaPos != -1 {
		chunk64 = chunk64[commaPos+1:]
	}
	chunkBytes, err := b64.StdEncoding.DecodeString(chunk64)
	if err != nil {
		util.DisplayError(err)
		illegalByte, byteErr := strconv.Atoi(strings.Replace(err.Error(), "illegal base64 data at input byte ", "", 1))
		if byteErr == nil {
			util.Debug.Println("Illegal byte:", string(chunk64[illegalByte]))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding chunk"})
		return
	}

	chunkBuffer := bytes.NewBuffer(chunkBytes)

	bsWritten, err := io.Copy(f, chunkBuffer)
	if err != nil || bsWritten == 0 {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error writing to tmp file"})
		return
	}

	f.Close()
	if rangeMax >= fileSize-1 {
		parentDir := dataStore.FsTreeGet(ctx.Request.FormValue("parentFolderId"))
		err = dataStore.FsTreeMove(tmpFile, parentDir, filename, false)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move tmp file to permenant location"})
			return
		}

		ctx.Status(http.StatusCreated)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "uploading", "uploadId": uploadId})
}

func uploadFile(ctx *gin.Context) {
	if strings.Contains(ctx.GetHeader("Content-Type"), "multipart/form-data") {
		handleChunkedFileUpload(ctx)
		return
	}

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body on file upload"})
		return
	}

	var fileMeta uploadedFile
	err = json.Unmarshal(jsonData, &fileMeta)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	parentFolder := dataStore.FsTreeGet(fileMeta.ParentFolderId)
	if parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find parent folder"})
		return
	}
	file, err := dataStore.Touch(parentFolder, fileMeta.FileName, false)

	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create file"})
		return
	}

	username := ctx.GetString("username")
	err = uploadItem(file, fileMeta.File64, username)
	if err != nil {
		util.DisplayError(err, "Failed to upload item")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload item"})
		return
	}

	dataStore.FsTreeInsert(file, parentFolder.Id())

	ctx.Status(http.StatusCreated)
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

	Caster.PushItemCreate(newDir)
}

func createUserHomeDir(username string) {
	homeDir, err := dataStore.Touch(dataStore.GetMediaDir(), username, true)
	util.DisplayError(err)

	_, err = dataStore.Touch(homeDir, ".user_trash", true)
	util.DisplayError(err)
}

func _getDirInfo(dir *dataStore.WeblensFileDescriptor, ctx *gin.Context) {
	username := ctx.GetString("username")

	selfData, _ := dir.FormatFileInfo()

	filteredDirInfo := dir.GetChildrenInfo()
	filteredDirInfo = util.Filter(filteredDirInfo, func(t dataStore.FileInfo) bool { return (t.Id != "R" && t.Filename != ".user_trash") })

	parentsInfo := []dataStore.FileInfo{}
	parent := dir.GetParent()
	if parent == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find parent directory"})
		return
	}
	for parent.Id() != "0" && parent.UserCanAccess(username) {
		parentInfo, err := parent.FormatFileInfo()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Failed to format parent file info")
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

	if !dir.UserCanAccess(ctx.GetString("username")) {
		util.Debug.Println("Not auth")
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	_getDirInfo(dir, ctx)
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
	if file.Err() != nil {
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
	}

	// If the directory does not change, just assume this is a rename
	if newParentId == "" {
		newParentId = file.GetParent().Id()
	}

	t := dataProcess.NewTask("move_file", dataProcess.MoveMeta{FileId: fileId, DestinationFolderId: newParentId, NewFilename: newFilename})
	dataProcess.QueueGlobalTask(t)

	t.Wait()

	if t.Err() != nil {
		util.Error.Println(t.Err())
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

type updateMany struct {
	Files       []string `json:"fileIds"`
	NewParentId string   `json:"newParentId"`
}

func updateFiles(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var filesData updateMany
	json.Unmarshal(jsonData, &filesData)

	wq := dataProcess.NewWorkQueue()

	// arbFile := dataStore.FsTreeGet(filesData.Files[0])
	// preParent := arbFile.GetParent()

	for _, fileId := range filesData.Files {
		t := dataProcess.NewTask("move_file", dataProcess.MoveMeta{FileId: fileId, DestinationFolderId: filesData.NewParentId, NewFilename: ""})
		wq.QueueTask(t)
	}
	wq.SignalAllQueued()
	wq.Wait()

	// postParent := arbFile.GetParent()

	// Caster.PushItemUpdate(preParent, postParent)
	ctx.Status(http.StatusOK)
}

func trashFile(ctx *gin.Context) {
	file := dataStore.FsTreeGet(ctx.Query("fileId"))
	trashDir := dataStore.GetUserTrashDir(ctx.GetString("username"))

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
		dataStore.PermenantlyDeleteFile(file)
	} else {
		err := dataStore.MoveFileToTrash(file)
		if err != nil {
			util.DisplayError(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

type takeoutItems struct {
	FileIds []string `json:"fileIds"`
}

func createTakeout(ctx *gin.Context) {
	username := ctx.GetString("username")

	var items takeoutItems
	bodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}

	err = json.Unmarshal(bodyData, &items)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request body not in expected format"})
		return
	}

	files := util.Map(items.FileIds, func(fileId string) *dataStore.WeblensFileDescriptor { return dataStore.FsTreeGet(fileId) })
	for _, file := range files {
		if file == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find at least one file"})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 item, and it is not a directory, we should have requested to just download that file
		util.Warning.Println("Creating zip file with only 1 non-dir item")
	}

	t := dataProcess.NewTask("create_zip", dataProcess.ZipMetadata{Files: files, Username: username})
	dataProcess.QueueGlobalTask(t)

	if t.Completed {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": t.Result("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": t.TaskId})
	}
}

func downloadFile(ctx *gin.Context) {
	file := dataStore.FsTreeGet(ctx.Query("fileId"))
	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	ctx.File(file.String())
}

type tokenReturn struct {
	Token string `json:"token"`
}

type newUserInfo struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Admin        bool   `json:"admin"`
	AutoActivate bool   `json:"autoActivate"`
}

func createUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var userInfo newUserInfo
	json.Unmarshal(jsonData, &userInfo)

	db := dataStore.NewDB(ctx.GetString("username"))
	db.GetUser(userInfo.Username)

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

	db := dataStore.NewDB(ctx.GetString("username"))
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
	db := dataStore.NewDB(ctx.GetString("username"))
	authList := strings.Split(ctx.Request.Header["Authorization"][0], ",")

	user, err := db.GetUser(authList[0])
	util.FailOnError(err, "Failed to get user info")
	user.Password = ""

	var empty []string
	user.Tokens = empty

	ctx.JSON(http.StatusOK, gin.H{"username": user.Username, "homeFolderId": dataStore.GetUserHomeDir(user.Username).Id(), "trashFolderId": dataStore.GetUserTrashDir(user.Username).Id(), "admin": user.Admin})

}

func getUsers(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
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

	db := dataStore.NewDB(ctx.GetString("username"))
	db.ActivateUser(userToUpdate.Username)
	createUserHomeDir(userToUpdate.Username)

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := ctx.Param("username") // User to delete username
	homeDir := dataStore.GetUserHomeDir(username)
	dataStore.MoveFileToTrash(homeDir)

	db := dataStore.NewDB(ctx.GetString("username")) // Admin username
	db.DeleteUser(username)

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
	db.ClearCache()
	dataProcess.FlushCompleteTasks()
}

func searchUsers(ctx *gin.Context) {
	filter := ctx.Query("filter")
	if len(filter) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autcomplete must contain at least 2 characters"})
		return
	}

	db := dataStore.NewDB(ctx.GetString("username"))
	users := db.SearchUsers(filter)
	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

type fileShare struct {
	Files []string `json:"files"`
	Users []string `json:"users"`
}

func shareFiles(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	var shareData fileShare
	json.Unmarshal(jsonData, &shareData)
	db := dataStore.NewDB(ctx.GetString("username"))

	if len(shareData.Files) == 0 || len(shareData.Users) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot share files with either 0 files or 0 users"})
		return
	}

	shareFiles := util.Map(shareData.Files, func(fileId string) *dataStore.WeblensFileDescriptor { return dataStore.FsTreeGet(fileId) })
	err = db.ShareFiles(shareFiles, shareData.Users)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed writing to database"})
		return
	}

	ctx.Status(http.StatusOK)
}

func getSharedFiles(ctx *gin.Context) {
	username := ctx.GetString("username")
	db := dataStore.NewDB(username)
	sharedList := db.GetSharedWith(username)

	filesInfos := util.Map(sharedList, func(file *dataStore.WeblensFileDescriptor) dataStore.FileInfo {
		fileInfo, _ := file.FormatFileInfo()
		return fileInfo
	})

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}

func getFileTreeInfo(ctx *gin.Context) {
	util.Debug.Println("File tree size: ", dataStore.GetTreeSize())
}
