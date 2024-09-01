package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
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

	err = pack.FileService.MoveFiles(children, newDir, pack.Caster)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, gin.H{"error": safe.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.ID()})
}

// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir *fileTree.WeblensFileImpl, pastTime time.Time, ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	share, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		safeErr, code := werror.TrySafeErr(err)
		ctx.JSON(code, safeErr)
		return
	}

	selfData, err := formatFileSafe(dir, u, share, pack)
	if err != nil {
		safeErr, code := werror.TrySafeErr(err)
		ctx.JSON(code, safeErr)
		return
	}

	var filteredDirInfo []FileInfo
	if dir.IsDir() {
		if pastTime.Unix() > 0 {
			ctx.Status(http.StatusNotImplemented)
			return
			// filteredDirInfo, err =  pack.FileService.GetJournal().GetPastFolderInfo(dir, acc.GetTime())
			// if err != nil {
			// 	wlog.ErrTrace(err)
			// 	ctx.JSON(http.StatusInternalServerError, "Failed to get folder info")
			// 	return
			// }
		} else {
			for _, child := range dir.GetChildren() {
				childInfo, err := formatFileSafe(child, u, share, pack)
				if err != nil {
					safeErr, code := werror.TrySafeErr(err)
					ctx.JSON(code, safeErr)
					return
				}
				filteredDirInfo = append(filteredDirInfo, childInfo)
			}
		}
	}

	if filteredDirInfo == nil {
		filteredDirInfo = []FileInfo{}
	}

	var parentsInfo []FileInfo
	parent := dir.GetParent()
	for parent.ID() != "ROOT" && pack.AccessService.CanUserAccessFile(
		u, parent, share,
	) && !pack.FileService.GetFileOwner(parent).IsSystemUser() {
		parentInfo, err := formatFileSafe(parent, u, share, pack)
		if err != nil {
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	packagedInfo := gin.H{"self": selfData, "children": filteredDirInfo, "parents": parentsInfo}
	ctx.JSON(http.StatusOK, packagedInfo)

	for _, child := range dir.GetChildren() {
		if pack.MediaService.Get(models.ContentId(child.GetContentId())) == nil && pack.MediaService.IsFileDisplayable(child) {
			c := models.NewBufferedCaster(pack.ClientService)
			meta := models.ScanMeta{
				File:         dir,
				FileService:  pack.FileService,
				MediaService: pack.MediaService,
				TaskService:  pack.TaskService,
				TaskSubber:   pack.ClientService,

				Caster: c,
			}
			t, err := pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
			if err != nil {
				log.ShowErr(err)
				return
			}
			t.SetCleanup(
				func(t *task.Task) {
					if c.IsEnabled() {
						c.Close()
					}
				},
			)
			break
		}
	}
}

func getFolder(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	folderId := fileTree.FileId(ctx.Param("folderId"))
	dir, err := pack.FileService.GetFileSafe(folderId, u, sh)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	formatRespondFolderInfo(dir, time.UnixMilli(0), ctx)
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
	// 	time.Sleep(time.Millisecond*150 - time.Since(start))
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

	caster := models.NewBufferedCaster(pack.ClientService)
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

func getPastFolderInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	folderId := fileTree.FileId(ctx.Param("folderId"))
	folder, err := pack.FileService.GetFileSafe(folderId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	milliStr := ctx.Param("rewindTime")
	millis, err := strconv.ParseInt(milliStr, 10, 64)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	pastTime := time.UnixMilli(millis)

	formatRespondFolderInfo(folder, pastTime, ctx)
}

func getFolderHistory(ctx *gin.Context) {
	ctx.Status(http.StatusNotImplemented)
	return
	// var path *fileTree.WeblensFilepath
	// fileId := fileTree.FileId(ctx.Param("fileId"))
	// if fileId == "" {
	// 	pathString := ctx.Query("path")
	// 	if pathString == "" {
	// 		ctx.Status(http.StatusBadRequest)
	// 		return
	// 	}
	//
	// 	if !strings.HasSuffix(pathString, "/") {
	// 		pathString += "/"
	// 	}
	//
	// 	path = fileTree.ParsePortable(pathString)
	// } else {
	// 	file :=  pack.FileService.Get(fileId)
	// 	if file == nil {
	// 		ctx.Status(http.StatusNotFound)
	// 		return
	// 	}
	// 	path = file.GetPortablePath()
	// }
	//
	// actions, err :=  pack.FileService.GetJournal().GetActionsByPath(path)
	// if err != nil {
	// 	wlog.ShowErr(err)
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }
	//
	// slices.SortFunc(
	// 	actions, func(a, b types.FileAction) int {
	// 		return b.GetTimestamp().Compare(a.GetTimestamp())
	// 	},
	// )
	//
	// ctx.JSON(http.StatusOK, actions)
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
	ctx.Status(http.StatusNotImplemented)
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

	err = pack.FileService.MoveFiles(files, newParent, pack.Caster)
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
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shares, err := pack.ShareService.GetFileSharesWithUser(u)
	if err != nil {
		safeErr, code := werror.TrySafeErr(err)
		ctx.JSON(code, safeErr)
		return
	}

	var filesInfos = make([]FileInfo, 0)
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
		fInfo, err := formatFileSafe(f, u, share, pack)
		if err != nil {
			safeErr, code := werror.TrySafeErr(err)
			ctx.JSON(code, safeErr)
			return
		}
		filesInfos = append(filesInfos, fInfo)
	}

	ctx.JSON(http.StatusOK, filesInfos)
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
	fileId := fileTree.FileId(ctx.Param("fileId"))

	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	size, err := f.Size()
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
	}

	ctx.JSON(
		http.StatusOK,
		FileStat{Name: f.Filename(), Size: size, IsDir: f.IsDir(), ModTime: f.ModTime(), Exists: true},
	)
}

func getDirectoryContent(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	fileId := fileTree.FileId(ctx.Param("fileId"))
	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	children := internal.Map(
		f.GetChildren(), func(c *fileTree.WeblensFileImpl) FileStat {
			size, err := c.Size()
			if err != nil {
				log.ShowErr(err)
				return FileStat{}
			}
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
	folder, children, err := pack.FileService.PathToFile(searchPath, u, nil)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"children": children, "folder": folder})
}

func getFileDataFromPath(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	searchPath := ctx.Query("searchPath")

	folder, _, err := pack.FileService.PathToFile(searchPath, u, nil)
	if err != nil {
		if errors.Is(err, werror.ErrNoFile) {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Status(http.StatusInternalServerError)
		log.ShowErr(err)
		return
	}

	ctx.JSON(http.StatusOK, folder)
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
