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
	"github.com/ethanrous/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func createFolder(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	body, err := readCtxBody[createFolderBody](ctx)
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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe.Error()})
		return
	}

	err = pack.FileService.MoveFiles(children, newDir, "USERS", pack.Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.ID()})
}

func formatRespondPastFolderInfo(folderId fileTree.FileId, pastTime time.Time, ctx *gin.Context) {
	pack := getServices(ctx)

	pastFile, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(folderId, pastTime)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
		return
	}

	var parentsInfo []*fileTree.WeblensFileImpl
	parentId := pastFile.GetParentId()
	if parentId == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find parent folder"})
		return
	}
	for parentId != "ROOT" {
		pastParent, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(parentId, pastTime)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, gin.H{"error": safe})
			return
		}

		parentsInfo = append(parentsInfo, pastParent)
		parentId = pastParent.GetParentId()
	}

	children, err := pack.FileService.GetJournalByTree("USERS").GetPastFolderChildren(pastFile, pastTime)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
		return
	}

	var medias []*models.Media
	for _, child := range children {
		m := pack.MediaService.Get(child.GetContentId())
		if m != nil {
			medias = append(medias, m)
		}
	}

	packagedInfo := gin.H{"self": pastFile, "children": children, "parents": parentsInfo, "medias": medias}
	ctx.JSON(http.StatusOK, packagedInfo)
}

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

// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir *fileTree.WeblensFileImpl, ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		safeErr, code := werror.TrySafeErr(err)
		ctx.JSON(code, safeErr)
		return
	}

	var parentsInfo []FileInfo
	parent := dir.GetParent()
	for parent.ID() != "ROOT" && pack.AccessService.CanUserAccessFile(
		u, parent, share,
	) && !pack.FileService.GetFileOwner(parent).IsSystemUser() {
		parentInfo, err := WeblensFileToFileInfo(parent, pack, false)
		if err != nil {
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	children := dir.GetChildren()

	mediaFiles := append(children, dir)
	medias, err := getChildMedias(pack, mediaFiles)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	childInfos := make([]FileInfo, 0, len(children))
	for _, child := range children {
		info, err := WeblensFileToFileInfo(child, pack, false)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
		childInfos = append(childInfos, info)
	}

	selfInfo, err := WeblensFileToFileInfo(dir, pack, true)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	packagedInfo := gin.H{"self": selfInfo, "children": childInfos, "parents": parentsInfo, "medias": medias}
	ctx.JSON(http.StatusOK, packagedInfo)
}

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
		if err != nil || millis < 0 {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	formatRespondFolderInfo(dir, ctx)
}

func getExternalDirs(ctx *gin.Context) {
	panic(werror.NotImplemented("getExternalDirs"))
	// externalRoot :=  pack.FileService.Get("EXTERNAL")
	// user := getUserFromCtx(ctx)
	// if user == nil {
	// 	ctx.Status(comm.StatusUnauthorized)
	// 	return
	// }
	// acc := dataStore.NewAccessMeta(user).SetRequestMode(dataStore.FileGet)
	// formatRespondFolderInfo(externalRoot, acc, ctx)
}

func getExternalFolderInfo(ctx *gin.Context) {
	panic(werror.NotImplemented("getExternalFolderInfo"))
	// start := time.Now()
	//
	// folderId := fileTree.FileId(ctx.Param("folderId"))
	// dir :=  pack.FileService.Get(folderId)
	// if dir == nil {
	// 	wlog.Debug.Println("Actually not found")
	// 	ctx.JSON(comm.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
	// 	return
	// }
	//
	// if dir.ID() == "" {
	// 	wlog.Error.Println("Blank file descriptor trying to get folder info")
	// 	time.Sleep(time.Millisecond*150 - time.Since(start))
	// 	ctx.Status(comm.StatusNotFound)
	// 	return
	// }
	//
	// user := getUserFromCtx(ctx)
	// if user == nil {
	// 	ctx.Status(comm.StatusUnauthorized)
	// 	return
	// }
	// acc := dataStore.NewAccessMeta(user)
	// formatRespondFolderInfo(dir, acc, ctx)
}

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
	var scanInfo scanBody
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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	actions, err := pack.FileService.GetJournalByTree("USERS").GetActionsByPath(f.GetPortablePath())
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}
	log.Trace.Printf("Found %d actions", len(actions))
	ctx.JSON(http.StatusOK, actions)
}

func restorePastFiles(ctx *gin.Context) {
	// body, err := readCtxBody[restoreBody](ctx)
	// if err != nil {
	// 	return
	// }
	//
	// t := time.UnixMilli(body.Timestamp)
	//
	// err = dataStore.RestoreFiles(body.FileIds, t, SERV.FileTree)
	// if err != nil {
	// 	util.ShowErr(err)
	// 	ctx.Status(comm.StatusInternalServerError)
	// 	return
	// }
	//
	// ctx.Status(comm.StatusOK)
	panic(werror.NotImplemented("restorePastFiles"))
	// ctx.Status(http.StatusNotImplemented)
}

func moveFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		return
	}

	filesData, err := readCtxBody[updateMany](ctx)
	if err != nil {
		return
	}

	newParent, err := pack.FileService.GetFileSafe(filesData.NewParentId, u, sh)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range filesData.Files {
		f, err := pack.FileService.GetFileSafe(fileId, u, sh)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}

		files = append(files, f)
	}

	err = pack.FileService.MoveFiles(files, newParent, "USERS", pack.Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func getSharedFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	log.Trace.Printf("Getting shared files for user %s", u.GetUsername())
	if u.IsPublicUser() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shares, err := pack.ShareService.GetFileSharesWithUser(u)
	if err != nil {
		safeErr, code := werror.TrySafeErr(err)
		ctx.JSON(code, safeErr)
		return
	}

	var children = make([]*fileTree.WeblensFileImpl, 0)
	for _, share := range shares {
		f, err := pack.FileService.GetFileSafe(share.FileId, u, share)
		if err != nil {
			if errors.Is(err, werror.ErrNoFile) {
				log.Warning.Println("Could not find file acompanying a file share")
				continue
			}
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		children = append(children, f)
	}

	childInfos := make([]FileInfo, 0, len(children))
	for _, child := range children {
		fInfo, err := WeblensFileToFileInfo(child, pack, false)
		if err != nil {
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		childInfos = append(childInfos, fInfo)
	}

	medias, err := getChildMedias(pack, children)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"children": childInfos, "medias": medias})
}

func getFileShare(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	shareId := models.ShareId(ctx.Param("shareId"))

	share := pack.ShareService.Get(shareId)
	if share == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	fileShare, ok := share.(*models.FileShare)
	if !ok {
		log.Warning.Printf(
			"%s tried to get share [%s] as a fileShare (is %s)", u.GetUsername(), shareId, share.GetShareType(),
		)
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, fileShare)
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

	files := internal.FilterMap(
		matches, func(match fuzzy.Rank) (*fileTree.WeblensFileImpl, bool) {
			// f, err := pack.FileService.GetFileSafe(children[match.OriginalIndex], u, nil)
			// if err != nil {
			// 	return nil, false
			// }
			// if f.ID() == u.HomeId || f.ID() == u.TrashId {
			// 	return nil, false
			// }
			return children[match.OriginalIndex], true
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"children": files, "folder": folder})
}

func getFileDataFromPath(ctx *gin.Context) {
	// pack := getServices(ctx)
	// u := getUserFromCtx(ctx)
	// searchPath := ctx.Query("searchPath")
	//
	// folder, _, err := pack.FileService.PathToFile(searchPath, u, nil)
	// if err != nil {
	// 	if errors.Is(err, werror.ErrNoFile) {
	// 		ctx.Status(http.StatusNotFound)
	// 		return
	// 	}
	// 	ctx.Status(http.StatusInternalServerError)
	// 	log.ShowErr(err)
	// 	return
	// }
	//
	// ctx.JSON(http.StatusOK, folder)
}

func searchByFilename(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	filenameSearch := ctx.Query("search")
	if filenameSearch == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	userHome, err := pack.FileService.GetFileSafe(u.HomeId, u, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var fileIds []fileTree.FileId
	var filenames []string
	_ = userHome.RecursiveMap(
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

	ctx.JSON(http.StatusOK, files)
}

func getFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}

	fileId := ctx.Param("fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	if file == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := formatFileSafe(file, u, nil, pack)
	if err != nil {
		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)
}

func getFileText(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		return
	}

	fileId := ctx.Param("fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe})
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

func updateFile(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileId := ctx.Param("fileId")
	updateInfo, err := readCtxBody[fileUpdateBody](ctx)
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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func trashFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	fileIds, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}
	u := getUserFromCtx(ctx)

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}

		if u != pack.FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You can't trash this file"})
			return
		}

		files = append(files, file)
	}
	err = pack.FileService.MoveFilesToTrash(files, u, nil, pack.Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func deleteFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileIds, err := readCtxBody[[]fileTree.FileId](ctx)
	if err != nil {
		return
	}

	if len(fileIds) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file ids provided"})
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to delete"})
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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func unTrashFiles(ctx *gin.Context) {
	pack := getServices(ctx)
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

	caster := models.NewSimpleCaster(pack.ClientService)
	defer caster.Close()

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if err != nil || u != pack.FileService.GetFileOwner(file) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file to untrash"})
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

func restoreFiles(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	body, err := readCtxBody[restoreFilesBody](ctx)
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
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	} else {
		newParent, err = pack.FileService.GetFileSafe(body.NewParentId, u, nil)
		if err != nil {
			safe, code := werror.TrySafeErr(err)
			ctx.JSON(code, safe)
			return
		}
	}

	err = pack.FileService.RestoreFiles(body.FileIds, newParent, restoreTime, pack.Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"newParentId": newParent.ID()})
}

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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	coverId, err := pack.FileService.GetFolderCover(folder)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	log.Trace.Printf("Cover id: [%s] [%s]", coverId, mediaId)

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
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	pack.Caster.PushFileUpdate(folder, m)

	ctx.Status(http.StatusOK)
}
