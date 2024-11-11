package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// GetFile godoc
//
//	@ID	GetFile
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get information about a file
//	@Tags		Files
//	@Produce	json
//	@Param		fileId	path		string			true	"File Id"
//	@Param		shareId	query		string			false	"Share Id"
//	@Success	200		{object}	rest.FileInfo	"File Info"
//	@Failure	401
//	@Failure	404
//	@Router		/files/{fileId} [get]
func getFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	fileId := ctx.Param("fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, sh)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := rest.WeblensFileToFileInfo(file, pack, false)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

// GetFileText godoc
//
//	@ID	GetFileText
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get the text of a text file
//	@Tags		Files
//	@Produce	plain
//	@Param		fileId	path		string	true	"File Id"
//	@Success	200		{string}	string	"File text"
//	@Failure	400
//	@Router		/files/{fileId}/text [get]
func getFileText(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}

	fileId := ctx.Param("fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	filename := file.Filename()

	dotIndex := strings.LastIndex(filename, ".")
	if filename[dotIndex:] != ".txt" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File is not a text file"})
		return
	}

	ctx.File(file.AbsPath())
}

// GetFileStats godoc
//
//	@ID	GetFileStats
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get the statistics of a file
//	@Tags		Files
//	@Produce	json
//	@Param		fileId	path	string	true	"File Id"
//	@Failure	400
//	@Failure	501
//	@Router		/files/{fileId}/stats [get]
func getFileStats(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := ctx.Param("fileId")

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

// DownloadFile godoc
//
//	@ID	DownloadFile
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Download a file
//	@Tags		Files
//	@Produce	octet-stream
//	@Param		fileId		path		string	true	"File Id"
//	@Param		shareId		query		string	false	"Share Id"
//	@Param		isTakeout	query		bool	false	"Is this a takeout file"	Enums(true, false)	default(false)
//	@Success	200			{string}	binary	"File content"
//	@Router		/files/{fileId}/download [get]
func downloadFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := ctx.Param("fileId")
	isTakeout := ctx.Query("isTakeout")

	var file *fileTree.WeblensFileImpl
	if isTakeout == "true" {
		var err error
		file, err = pack.FileService.GetZip(fileId)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
	} else {
		share, err := getShareFromCtx[*models.FileShare](ctx)
		if err != nil {
			return
		}

		file, err = pack.FileService.GetFileSafe(fileId, u, share)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
	}

	ctx.File(file.AbsPath())
}

// GetFolderHistory godoc
//
//	@ID			GetFolderHistory
//
//	@Security	SessionAuth
//
//	@Summary	Get actions of a folder at a given time
//	@Tags		Folder
//	@Param		fileId		path	string				true	"File Id"
//	@Param		timestamp	query	int					true	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{array}	fileTree.FileAction	"File actions"
//	@Failure	400
//	@Failure	500
//	@Router		/files/{fileId}/history [get]
func getFolderHistory(ctx *gin.Context) {
	pack := getServices(ctx)

	fileId := ctx.Param("fileId")
	if fileId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	pastTime := time.Now()

	milliStr := ctx.Query("timestamp")
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if err != nil {
			log.ShowErr(werror.WithStack(err))
			ctx.Status(http.StatusBadRequest)
			return
		}
		pastTime = time.UnixMilli(millis)
	}

	f, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(fileId, pastTime)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	actions, err := pack.FileService.GetJournalByTree("USERS").GetActionsByPath(f.GetPortablePath())
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}
	log.Trace.Printf("Found %d actions", len(actions))
	ctx.JSON(http.StatusOK, actions)
}

// SearchByFilename godoc
//
//	@ID			SearchByFilename
//
//	@Security	SessionAuth
//
//	@Summary	Search for files by filename
//	@Tags		Files
//
//	@Param		search			query	string			true	"Filename to search for"
//	@Param		baseFolderId	query	string			false	"The folder to search in, defaults to the user's home folder"
//	@Success	200				{array}	rest.FileInfo	"File Info"
//	@Failure	400
//	@Failure	401
//	@Failure	500
//	@Router		/files/search [get]
func searchByFilename(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	filenameSearch := ctx.Query("search")
	if filenameSearch == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	baseFolderId := ctx.Query("baseFolderId")
	if baseFolderId == "" {
		baseFolderId = u.HomeId
	}

	baseFolder, err := pack.FileService.GetFileSafe(baseFolderId, u, nil)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	if !baseFolder.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	var fileIds []fileTree.FileId
	var filenames []string
	_ = baseFolder.RecursiveMap(
		func(f *fileTree.WeblensFileImpl) error {
			fileIds = append(fileIds, f.ID())
			filenames = append(filenames, f.Filename())
			return nil
		},
	)

	matches := fuzzy.RankFindFold(filenameSearch, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			return a.Distance - b.Distance
		},
	)

	files := internal.FilterMap(
		matches, func(match fuzzy.Rank) (*fileTree.WeblensFileImpl, bool) {
			f, err := pack.FileService.GetFileSafe(fileIds[match.OriginalIndex], u, nil)
			if err != nil {
				return nil, false
			}
			if f.ID() == u.HomeId || f.ID() == u.TrashId {
				return nil, false
			}
			return f, true
		},
	)

	var fileInfos []rest.FileInfo
	for _, file := range files {
		f, err := rest.WeblensFileToFileInfo(file, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		fileInfos = append(fileInfos, f)
	}

	ctx.JSON(http.StatusOK, fileInfos)
}

// CreateFolder godoc
//
//	@ID	CreateFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Create a new folder
//	@Tags		Folder
//	@Accept		json
//	@Produce	json
//	@Param		request	body		rest.CreateFolderBody	true	"New folder body"
//	@Param		shareId	query		string					false	"Share Id"
//	@Success	200		{object}	rest.FileInfo			"File Info"
//	@Router		/folder [post]
func createFolder(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	body, err := readCtxBody[rest.CreateFolderBody](ctx)
	if err != nil {
		return
	}

	if body.NewFolderName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing body parameter 'newFolderName'"})
		return
	}

	parentFolder, err := pack.FileService.GetFileSafe(body.ParentFolderId, u, nil)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}

	var children []*fileTree.WeblensFileImpl
	if len(body.Children) != 0 {
		for _, fileId := range body.Children {
			child, err := pack.FileService.GetFileSafe(fileId, u, nil)
			if err != nil {
				log.ShowErr(err)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			children = append(children, child)
		}
	}

	newDir, err := pack.FileService.CreateFolder(parentFolder, body.NewFolderName, pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	err = pack.FileService.MoveFiles(children, newDir, "USERS", pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	newDirInfo, err := rest.WeblensFileToFileInfo(newDir, pack, false)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.JSON(http.StatusCreated, newDirInfo)
}

// GetFolder godoc
//
//	@ID	GetFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get a folder
//	@Tags		Folder
//	@Accept		json
//	@Produce	json
//	@Param		folderId	path		string	true	"Folder Id"
//	@Param		shareId		query		string	false	"Share Id"
//	@Param		timestamp	query		int		false	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{object}	rest.FolderInfoResponse"Folder Info"
//	@Router		/folder/{folderId} [get]
func getFolder(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	milliStr := ctx.Query("timestamp")
	date := time.UnixMilli(0)
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		} else if millis < 0 {
			ctx.Status(http.StatusBadRequest)
			return
		}

		date = time.UnixMilli(millis)
	}

	folderId := ctx.Param("folderId")
	if folderId == "" {
		log.Trace.Println("No folder id provided")
		ctx.Status(http.StatusBadRequest)
		return
	}

	if date.Unix() != 0 {
		formatRespondPastFolderInfo(folderId, date, ctx)
		return
	}

	dir, err := pack.FileService.GetFileSafe(folderId, u, sh)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	formatRespondFolderInfo(dir, ctx)
}

// SetFolderCover godoc
//
//	@ID			SetFolderCover
//
//	@Security	SessionAuth
//
//	@Summary	Set the cover image of a folder
//	@Tags		Folder
//	@Param		folderId	path	string	true	"Folder Id"
//	@Param		mediaId		query	string	true	"Media Id"
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/folder/{folderId}/cover [patch]
func setFolderCover(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	folderId := ctx.Param("folderId")
	mediaId := ctx.Query("mediaId")

	if folderId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	folder, err := pack.FileService.GetFileSafe(folderId, u, nil)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	coverId, err := pack.FileService.GetFolderCover(folder)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	if coverId == "" && mediaId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	} else if coverId == mediaId {
		ctx.Status(http.StatusOK)
		return
	}

	var m *models.Media
	if mediaId != "" {
		m = pack.MediaService.Get(mediaId)
		if m == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
	}

	err = pack.FileService.SetFolderCover(folderId, mediaId)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	pack.Caster.PushFileUpdate(folder, m)

	ctx.Status(http.StatusOK)
}

func getExternalDirs(ctx *gin.Context) {
	panic(werror.NotImplemented("getExternalDirs"))
}

func getExternalFolderInfo(ctx *gin.Context) {
	panic(werror.NotImplemented("getExternalFolderInfo"))
}

// ScanFolder godoc
//
//	@ID	ScanFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Dispatch a folder scan
//	@Tags		Folder
//	@Param		shareId	query	string			false	"Share Id"
//	@Param		request	body	rest.ScanBody	true	"Scan parameters"
//	@Success	200
//	@Failure	404
//	@Failure	500
//	@Router		/admin/folder/scan [post]
func scanDir(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var scanInfo rest.ScanBody
	err = json.Unmarshal(body, &scanInfo)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	dir, err := pack.FileService.GetFileSafe(scanInfo.FolderId, u, sh)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	caster := models.NewSimpleCaster(pack.ClientService)
	meta := models.ScanMeta{
		File:         dir,
		FileService:  pack.FileService,
		MediaService: pack.MediaService,
		TaskSubber:   pack.ClientService,
		TaskService:  pack.TaskService,
		Caster:       caster,
	}
	t, err := pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	t.SetCleanup(
		func(t *task.Task) {
			if caster.IsEnabled() {
				caster.Close()
			}
		},
	)

	ctx.Status(http.StatusOK)
}

// GetSharedFiles godoc
//
//	@ID			GetSharedFiles
//
//	@Security	SessionAuth
//
//	@Summary	Get files shared with the logged in user
//	@Tags		Files
//	@Success	200	{object}	rest.FolderInfoResponse"An object containing all the files shared with the user"
//	@Failure	404
//	@Failure	500
//	@Router		/files/shared [get]
func getSharedFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	log.Trace.Printf("Getting shared files for user %s", u.GetUsername())

	shares, err := pack.ShareService.GetFileSharesWithUser(u)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var children = make([]*fileTree.WeblensFileImpl, 0)
	for _, share := range shares {
		f, err := pack.FileService.GetFileSafe(share.FileId, u, share)
		if err != nil {
			if errors.Is(err, werror.ErrNoFile) {
				log.Error.Println("Could not find file acompanying a file share")
				continue
			}
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		children = append(children, f)
	}

	childInfos := make([]rest.FileInfo, 0, len(children))
	for _, child := range children {
		fInfo, err := rest.WeblensFileToFileInfo(child, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		childInfos = append(childInfos, fInfo)
	}

	medias, err := getChildMedias(pack, children)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	res := rest.FolderInfoResponse{Children: childInfos, Medias: medias}

	ctx.JSON(http.StatusOK, res)
}

// CreateTakeout godoc
//
//	@ID	CreateTakeout
//
//	@Security
//	@Security		SessionAuth
//
//	@Summary		Create a zip file
//	@Description	Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
//	@Tags			Files
//	@Param			shareId	query		string					false	"Share Id"
//	@Param			request	body		rest.FilesListParams	true	"File Ids"
//	@Success		200		{object}	rest.TakeoutInfo		"Zip Takeout Info"
//	@Success		202		{object}	rest.TakeoutInfo		"Task Dispatch Info"
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/takeout [post]
func createTakeout(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	takeoutRequest, err := readCtxBody[rest.FilesListParams](ctx)
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
		if werror.SafeErrorAndExit(err, ctx) {
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
		res := rest.TakeoutInfo{TakeoutId: t.GetResult("takeoutId").(string), Single: false, Filename: t.GetResult("filename").(string)}
		ctx.JSON(http.StatusOK, res)
	} else {
		res := rest.TakeoutInfo{TaskId: t.TaskId()}
		ctx.JSON(http.StatusAccepted, res)
	}
}

func getFileStat(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileId := ctx.Param("fileId")

	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	size := f.Size()

	ctx.JSON(
		http.StatusOK,
		FileStat{Name: f.Filename(), Size: size, IsDir: f.IsDir(), ModTime: f.ModTime(), Exists: true},
	)
}

func getDirectoryContent(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileId := ctx.Param("fileId")
	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	children := internal.Map(
		f.GetChildren(), func(c *fileTree.WeblensFileImpl) FileStat {
			size := c.Size()
			return FileStat{Name: c.Filename(), Size: size, IsDir: c.IsDir(), ModTime: c.ModTime(), Exists: true}
		},
	)

	ctx.JSON(http.StatusOK, children)
}

// AutocompletePath godoc
//
//	@ID			AutocompletePath
//
//	@Security	SessionAuth
//
//	@Summary	Get path completion suggestions
//	@Tags		Files
//	@Param		searchPath	query		string	true	"Search path"
//	@Success	200			{object}	rest.FolderInfoResponse"Path info"
//	@Failure	500
//	@Router		/files/autocomplete [get]
func autocompletePath(ctx *gin.Context) {
	pack := getServices(ctx)
	searchPath := ctx.Query("searchPath")
	if len(searchPath) == 0 {
		ctx.Status(http.StatusBadRequest)
		return
	}

	u := getUserFromCtx(ctx)

	lastSlashI := strings.LastIndex(searchPath, "/")
	childName := searchPath[lastSlashI+1:]
	searchPath = searchPath[:lastSlashI] + "/"

	folder, err := pack.FileService.UserPathToFile(searchPath, u)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	children := folder.GetChildren()
	if folder.GetParentId() == "ROOT" {
		trashIndex := slices.IndexFunc(children, func(f *fileTree.WeblensFileImpl) bool {
			return f.ID() == u.TrashId
		})
		children = internal.Banish(children, trashIndex)
	}

	var filenames []string
	for _, child := range children {
		filenames = append(filenames, child.Filename())
	}

	matches := fuzzy.RankFindFold(childName, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			diff := a.Distance - b.Distance
			if diff == 0 {
				return strings.Compare(a.Target, b.Target)
			} else {
				return diff
			}
		},
	)

	var childInfos []rest.FileInfo
	for _, match := range matches {
		f := children[match.OriginalIndex]

		childInfo, err := rest.WeblensFileToFileInfo(f, pack, true)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		childInfos = append(childInfos, childInfo)
	}

	selfInfo, err := rest.WeblensFileToFileInfo(folder, pack, true)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ret := rest.FolderInfoResponse{Children: childInfos, Self: selfInfo}
	ctx.JSON(http.StatusOK, ret)
}

// RestoreFiles godoc
//
//	@ID			RestoreFiles
//
//	@Security	SessionAuth
//
//	@Summary	Restore files from some time in the past
//	@Tags		Files History
//	@Accept		json
//	@Produce	json
//	@Param		request	body		rest.RestoreFilesBody				true	"Restore files request body"
//	@Success	200		{object}	http.restoreFiles.restoreFilesInfo	"Restore files info"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/files/restore [post]
func restoreFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	body, err := readCtxBody[rest.RestoreFilesBody](ctx)
	if err != nil {
		return
	}

	if body.Timestamp == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing body parameter 'timestamp'"})
		return
	}
	restoreTime := time.UnixMilli(body.Timestamp)

	lt := pack.FileService.GetJournalByTree("USERS").Get(body.NewParentId)
	if lt == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find new parent"})
		return
	}

	var newParent *fileTree.WeblensFileImpl
	if lt.GetLatestAction().GetActionType() == fileTree.FileDelete {
		newParent, err = pack.FileService.GetFileSafe(u.HomeId, u, nil)

		// this should never error, but you never know
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	} else {
		newParent, err = pack.FileService.GetFileSafe(body.NewParentId, u, nil)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	}

	err = pack.FileService.RestoreFiles(body.FileIds, newParent, restoreTime, pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	type restoreFilesInfo struct {
		NewParentId string `json:"newParentId"`
	} // @name RestoreFilesInfo
	res := restoreFilesInfo{NewParentId: newParent.ID()}

	ctx.JSON(http.StatusOK, res)
}

// UpdateFile godoc
//
//	@ID	UpdateFile
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Update a File
//	@Tags		Files
//	@Accept		json
//	@Param		fileId	path	string					true	"File Id"
//	@Param		request	body	rest.UpdateFileParams	true	"Update file request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/{fileId} [patch]
func updateFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileId := ctx.Param("fileId")
	updateInfo, err := readCtxBody[rest.UpdateFileParams](ctx)
	if err != nil {
		return
	}

	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if pack.FileService.IsFileInTrash(file) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot rename file in trash"})
		return
	}

	err = pack.FileService.RenameFile(file, updateInfo.NewName, pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
}

// MoveFiles godoc
//
//	@ID	MoveFiles
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Move a list of files to a new parent folder
//	@Tags		Files
//	@Param		request	body	rest.MoveFilesParams	true	"Move files request body"
//	@Param		shareId	query	string					false	"Share Id"
//	@Success	200
//	@Failure	404
//	@Failure	500
//	@Router		/files [patch]
func moveFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	filesData, err := readCtxBody[rest.MoveFilesParams](ctx)
	if err != nil {
		return
	}

	newParent, err := pack.FileService.GetFileSafe(filesData.NewParentId, u, sh)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var files []*fileTree.WeblensFileImpl
	parentId := ""
	for _, fileId := range filesData.Files {
		f, err := pack.FileService.GetFileSafe(fileId, u, sh)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		if parentId == "" {
			parentId = f.GetParentId()
		} else if parentId != f.GetParentId() {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "All files must have the same parent"})
			return
		}

		files = append(files, f)
	}

	err = pack.FileService.MoveFiles(files, newParent, "USERS", pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
}

// TrashFiles godoc
//
//	@ID			TrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files to the trash
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"Trash files request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/trash [patch]
func trashFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	params, err := readCtxBody[rest.FilesListParams](ctx)
	if err != nil {
		return
	}
	fileIds := params.FileIds
	u := getUserFromCtx(ctx)

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		if u != pack.FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You can't trash this file"})
			return
		}

		files = append(files, file)
	}
	err = pack.FileService.MoveFilesToTrash(files, u, nil, pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
}

// UnTrashFiles godoc
//
//	@ID			UnTrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files out of the trash, restoring them to where they were before
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"UnTrash files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files/untrash [patch]
func unTrashFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	params, err := readCtxBody[rest.FilesListParams](ctx)
	if err != nil {
		return
	}
	fileIds := params.FileIds

	caster := models.NewSimpleCaster(pack.ClientService)
	defer caster.Close()

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if err != nil || u != pack.FileService.GetFileOwner(file) {
			ctx.Status(http.StatusNotFound)
			return
		}
		files = append(files, file)
	}

	err = pack.FileService.ReturnFilesFromTrash(files, caster)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

// DeleteFiles godoc
//
//	@ID			DeleteFiles
//
//	@Security	SessionAuth
//
//	@Summary	Delete Files "permanently"
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"Delete files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files [delete]
func deleteFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	params, err := readCtxBody[rest.FilesListParams](ctx)
	if err != nil {
		return
	}
	fileIds := params.FileIds

	if len(fileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file ids provided"})
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		} else if u != pack.FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
			return
		} else if !pack.FileService.IsFileInTrash(file) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete file not in trash"})
			return
		}
		files = append(files, file)
	}

	err = pack.FileService.DeleteFiles(files, "USERS", pack.Caster)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
}

// Helper Function
func getChildMedias(pack *models.ServicePack, children []*fileTree.WeblensFileImpl) ([]*models.Media, error) {
	var medias []*models.Media
	for _, child := range children {
		var m *models.Media
		contentId := child.GetContentId()
		if child.IsDir() && contentId == "" {
			coverId, err := pack.FileService.GetFolderCover(child)
			if err != nil {
				return nil, err
			}

			log.Trace.Printf("Cover id: %s", coverId)

			if coverId != "" {
				child.SetContentId(coverId)
				contentId = coverId
			}
		}

		m = pack.MediaService.Get(contentId)

		if m != nil {
			medias = append(medias, m)
		}
	}

	return medias, nil
}

// Helper Function
// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir *fileTree.WeblensFileImpl, ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	share, err := getShareFromCtx[*models.FileShare](ctx)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var parentsInfo []rest.FileInfo
	parent := dir.GetParent()
	for parent.ID() != "ROOT" && pack.AccessService.CanUserAccessFile(u, parent, share) && !pack.FileService.GetFileOwner(parent).IsSystemUser() {
		parentInfo, err := rest.WeblensFileToFileInfo(parent, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	children := dir.GetChildren()

	mediaFiles := append(children, dir)
	medias, err := getChildMedias(pack, mediaFiles)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	childInfos := make([]rest.FileInfo, 0, len(children))
	for _, child := range children {
		info, err := rest.WeblensFileToFileInfo(child, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		childInfos = append(childInfos, info)
	}

	selfInfo, err := rest.WeblensFileToFileInfo(dir, pack, true)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	packagedInfo := rest.FolderInfoResponse{Self: selfInfo, Children: childInfos, Parents: parentsInfo, Medias: medias}
	ctx.JSON(http.StatusOK, packagedInfo)
}

// Helper Function
func formatRespondPastFolderInfo(folderId fileTree.FileId, pastTime time.Time, ctx *gin.Context) {
	pack := getServices(ctx)

	pastFile, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(folderId, pastTime)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}
	pastFileInfo, err := rest.WeblensFileToFileInfo(pastFile, pack, false)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var parentsInfo []rest.FileInfo
	parentId := pastFile.GetParentId()
	if parentId == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find parent folder"})
		return
	}
	for parentId != "ROOT" {
		pastParent, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(parentId, pastTime)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		parentInfo, err := rest.WeblensFileToFileInfo(pastParent, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		parentsInfo = append(parentsInfo, parentInfo)
		parentId = pastParent.GetParentId()
	}

	children, err := pack.FileService.GetJournalByTree("USERS").GetPastFolderChildren(pastFile, pastTime)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var childrenInfos []rest.FileInfo
	for _, child := range children {
		childInfo, err := rest.WeblensFileToFileInfo(child, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		childrenInfos = append(childrenInfos, childInfo)
	}

	var medias []*models.Media
	for _, child := range children {
		m := pack.MediaService.Get(child.GetContentId())
		if m != nil {
			medias = append(medias, m)
		}
	}

	packagedInfo := rest.FolderInfoResponse{Self: pastFileInfo, Children: childrenInfos, Parents: parentsInfo, Medias: medias}
	ctx.JSON(http.StatusOK, packagedInfo)
}
