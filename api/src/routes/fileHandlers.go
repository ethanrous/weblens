package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/filetree"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func createFolder(ctx *gin.Context) {
	body, err := readCtxBody[createFolderBody](ctx)
	if err != nil {
		return
	}

	if body.NewFolderName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing body parameter 'new_folder_name'"})
		return
	}

	parentFolder := types.SERV.FileTree.Get(body.ParentFolderId)
	if parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}

	caster := NewBufferedCaster()
	defer caster.Close()

	event := history.NewFileEvent()

	newDir, err := types.SERV.FileTree.MkDir(parentFolder, body.NewFolderName, event, caster)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Children) != 0 {
		for _, fileId := range body.Children {
			child := types.SERV.FileTree.Get(fileId)
			err = types.SERV.FileTree.Move(child, newDir, "", false, event, caster)
			if err != nil {
				util.ShowErr(err)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(event)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.ID()})
}

// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir types.WeblensFile, acc types.AccessMeta, ctx *gin.Context) {
	selfData, err := dir.FormatFileInfo(acc)
	if err != nil {
		if errors.Is(err, dataStore.ErrNoFileAccess) {
			ctx.JSON(http.StatusNotFound, "Failed to get folder info")
			return
		}
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, "Failed to get folder info")
		return
	}

	var filteredDirInfo []types.FileInfo
	if dir.IsDir() {
		if acc.GetTime().Unix() > 0 {
			filteredDirInfo, err = types.SERV.FileTree.GetJournal().GetPastFolderInfo(dir, acc.GetTime())
			if err != nil {
				util.ErrTrace(err)
				ctx.JSON(http.StatusInternalServerError, "Failed to get folder info")
				return
			}
		} else {
			filteredDirInfo = dir.GetChildrenInfo(acc)
		}
	} else {
		filteredDirInfo = []types.FileInfo{}
	}

	var parentsInfo []types.FileInfo
	parent := dir.GetParent()
	for acc.CanAccessFile(parent) && parent != dir && (parent.Owner() != types.SERV.UserService.Get("WEBLENS")) {
		parentInfo, err := parent.FormatFileInfo(acc)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, "Failed to format parent file info")
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	packagedInfo := gin.H{"self": selfData, "children": filteredDirInfo, "parents": parentsInfo}
	ctx.JSON(http.StatusOK, packagedInfo)

	if slices.ContainsFunc(filteredDirInfo, func(i types.FileInfo) bool { return !i.Imported && i.Displayable }) {
		c := NewBufferedCaster()
		t := types.SERV.TaskDispatcher.ScanDirectory(dir, c)
		t.SetCleanup(
			func() {
				if c.IsEnabled() {
					c.Close()
				}
			},
		)
	}
}

func getFolder(ctx *gin.Context) {
	start := time.Now()
	user := getUserFromCtx(ctx)

	folderId := types.FileId(ctx.Param("folderId"))
	dir := types.SERV.FileTree.Get(folderId)
	if dir == nil {
		util.Debug.Println("Actually not found")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	shareId := types.ShareId(ctx.Query("shareId"))
	var sh types.Share
	var err error
	if shareId != "" {
		sh = types.SERV.ShareService.Get(shareId)
		if sh == nil {
			util.ShowErr(err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}
	}

	acc := dataStore.NewAccessMeta(user)
	if sh != nil {
		err = acc.AddShare(sh)
		if errors.Is(err, types.ErrUserNotAuthorized) {
			ctx.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	if !acc.CanAccessFile(dir) {
		util.Debug.Println("Not auth")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	// if !dir.IsDir() {
	// 	dir = dir.GetParent()
	// }

	formatRespondFolderInfo(dir, acc, ctx)
}

func getExternalDirs(ctx *gin.Context) {
	externalRoot := types.SERV.FileTree.Get("EXTERNAL")
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	acc := dataStore.NewAccessMeta(user).SetRequestMode(dataStore.FileGet)
	formatRespondFolderInfo(externalRoot, acc, ctx)
}

func getExternalFolderInfo(ctx *gin.Context) {
	start := time.Now()

	folderId := types.FileId(ctx.Param("folderId"))
	dir := types.SERV.FileTree.Get(folderId)
	if dir == nil {
		util.Debug.Println("Actually not found")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	if dir.ID() == "" {
		util.Error.Println("Blank file descriptor trying to get folder info")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.Status(http.StatusNotFound)
		return
	}

	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	acc := dataStore.NewAccessMeta(user)
	formatRespondFolderInfo(dir, acc, ctx)
}

func recursiveScanDir(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var scanInfo scanBody
	err = json.Unmarshal(body, &scanInfo)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	dir := types.SERV.FileTree.Get(scanInfo.FolderId)
	if dir == nil {
		return
	}
	types.SERV.TaskDispatcher.ScanDirectory(dir, types.SERV.Caster)

	ctx.Status(http.StatusOK)
}

func getPastFolderInfo(ctx *gin.Context) {
	folderId := types.FileId(ctx.Param("folderId"))
	milliStr := ctx.Param("rewindTime")
	user := getUserFromCtx(ctx)

	millis, err := strconv.ParseInt(milliStr, 10, 64)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	before := time.UnixMilli(millis)
	acc := dataStore.NewAccessMeta(user).SetTime(before)

	folder := types.SERV.FileTree.Get(folderId)
	formatRespondFolderInfo(folder, acc, ctx)
}

func getFolderHistory(ctx *gin.Context) {
	var path types.WeblensFilepath
	fileId := types.FileId(ctx.Param("fileId"))
	if fileId == "" {
		pathString := ctx.Query("path")
		if pathString == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}

		if !strings.HasSuffix(pathString, "/") {
			pathString += "/"
		}

		path = filetree.FilepathFromPortable(pathString)
	} else {
		file := types.SERV.FileTree.Get(fileId)
		if file == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		path = file.GetPortablePath()
	}

	actions, err := types.SERV.FileTree.GetJournal().GetActionsByPath(path)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	slices.SortFunc(
		actions, func(a, b types.FileAction) int {
			return b.GetTimestamp().Compare(a.GetTimestamp())
		},
	)

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
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }
	//
	// ctx.Status(http.StatusOK)

	ctx.Status(http.StatusNotImplemented)
}

func moveFiles(ctx *gin.Context) {
	filesData, err := readCtxBody[updateMany](ctx)
	if err != nil {
		return
	}

	tp := types.SERV.TaskDispatcher.GetWorkerPool().NewTaskPool(false, nil)

	caster := NewBufferedCaster()
	fileEvent := history.NewFileEvent()
	defer caster.Close()

	for _, fileId := range filesData.Files {
		tp.MoveFile(fileId, filesData.NewParentId, "", fileEvent, caster)
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(fileEvent)
	if err != nil {
		util.ErrTrace(err)
	}

	tp.SignalAllQueued()
	tp.Wait(false)

	ctx.Status(http.StatusOK)
}

func getSharedFiles(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	shares, err := types.SERV.ShareService.GetSharedWithUser(u)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	acc := dataStore.NewAccessMeta(u)
	filesInfos := util.FilterMap(
		shares, func(sh types.Share) (types.FileInfo, bool) {
			err = acc.AddShare(sh)
			acc.SetUsingShare(sh)
			if err != nil {
				util.ShowErr(err)
			}
			f := types.SERV.FileTree.Get(types.FileId(sh.GetItemId()))
			if f == nil {
				util.Error.Println("Cannot find file when getting shared files", sh.GetItemId())
				return types.FileInfo{}, false
			}
			fileInfo, err := f.FormatFileInfo(acc)
			if err != nil {
				util.ShowErr(err)
			}
			return fileInfo, true
		},
	)

	util.Debug.Printf("Got %d shared files for %s", len(filesInfos), u.GetUsername())

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}

func getFileShare(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	shareId := types.ShareId(ctx.Param("shareId"))

	sh := types.SERV.ShareService.Get(shareId)
	if sh == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	acc := dataStore.NewAccessMeta(u)
	err := acc.AddShare(sh)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, sh)
}

func getFileStat(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := types.SERV.FileTree.Get(fileId)
	if f == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	// err := f.LoadStat()
	// if err != nil {
	// 	ctx.JSON(http.StatusOK, types.FileStat{Exists: false})
	// }

	size, err := f.Size()
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
	}

	ctx.JSON(
		http.StatusOK,
		types.FileStat{Name: f.Filename(), Size: size, IsDir: f.IsDir(), ModTime: f.ModTime(), Exists: true},
	)
}

func getDirectoryContent(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := types.SERV.FileTree.Get(fileId)
	if f == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	children := util.Map(
		f.GetChildren(), func(c types.WeblensFile) types.FileStat {
			size, err := c.Size()
			if err != nil {
				util.ShowErr(err)
				return types.FileStat{}
			}
			return types.FileStat{Name: c.Filename(), Size: size, IsDir: c.IsDir(), ModTime: c.ModTime(), Exists: true}
		},
	)

	ctx.JSON(http.StatusOK, children)
}

func autocompletePath(ctx *gin.Context) {
	searchPath := ctx.Query("searchPath")

	lastSlashIndex := strings.LastIndex(searchPath, "/")
	if lastSlashIndex == -1 {
		if !strings.HasSuffix(searchPath, "/") {
			searchPath += "/"
		}
		lastSlashIndex = len(searchPath) - 1
	}
	prefix := searchPath[:lastSlashIndex+1]
	folderId := types.SERV.FileTree.GenerateFileId(filetree.FilepathFromPortable(prefix).ToAbsPath() + "/")

	folder := types.SERV.FileTree.Get(folderId)
	if folder == nil {
		ctx.JSON(http.StatusOK, gin.H{"children": []string{}, "folder": nil})
		return
	}

	postFix := searchPath[lastSlashIndex+1:]
	children := util.FilterMap(
		folder.GetChildren(), func(c types.WeblensFile) (types.WeblensFile, bool) {
			return c, strings.HasPrefix(c.Filename(), postFix)
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"children": children, "folder": folder})
}

func getFileDataFromPath(ctx *gin.Context) {
	searchPath := ctx.Query("searchPath")
	folderId := types.SERV.FileTree.GenerateFileId(filetree.FilepathFromPortable(searchPath).ToAbsPath() + "/")

	folder := types.SERV.FileTree.Get(folderId)
	if folder == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, folder)
}
