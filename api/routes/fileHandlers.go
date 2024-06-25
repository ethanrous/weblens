package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
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

	newDir, err := types.SERV.FileTree.MkDir(parentFolder, body.NewFolderName, caster)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Children) != 0 {
		for _, fileId := range body.Children {
			child := types.SERV.FileTree.Get(fileId)
			err = types.SERV.FileTree.Move(child, newDir, "", false, caster)
			if err != nil {
				util.ShowErr(err)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
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
			ctx.Status(http.StatusNotImplemented)
			return
			// filteredDirInfo, err = dataStore.GetPastFileInfo(dir, acc)
			// if err != nil {
			// 	util.ShowErr(err)
			// 	ctx.Status(http.StatusInternalServerError)
			// 	return
			// }
		} else {
			filteredDirInfo = dir.GetChildrenInfo(acc)
		}
	}

	var parentsInfo []types.FileInfo
	parent := dir.GetParent()
	for acc.CanAccessFile(parent) && parent != dir && (parent.Owner() != dataStore.WeblensRootUser) {
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
				c.Close()
			},
		)
	}

}

func getFolder(ctx *gin.Context) {
	start := time.Now()
	user := getUserFromCtx(ctx)
	// if user == nil {
	//	ctx.Status(http.StatusUnauthorized)
	//	return
	// }

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

	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree)
	if sh != nil {
		err = acc.AddShare(sh)
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

	if !dir.IsDir() {
		dir = dir.GetParent()
	}

	formatRespondFolderInfo(dir, acc, ctx)
}

func getExternalDirs(ctx *gin.Context) {
	externalRoot := types.SERV.FileTree.Get("EXTERNAL")
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree).SetRequestMode(dataStore.FileGet)
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
	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree)
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
	milliStr := ctx.Query("before")
	user := getUserFromCtx(ctx)

	millis, err := strconv.ParseInt(milliStr, 10, 64)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	before := time.UnixMilli(millis)
	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree).SetTime(before)

	folder := types.SERV.FileTree.Get(folderId)
	formatRespondFolderInfo(folder, acc, ctx)
}

func getFileHistory(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	file := types.SERV.FileTree.Get(fileId)

	actions, err := types.SERV.Database.GetActionsByPath(file.GetPortablePath())
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, actions)
	// ctx.Status(http.StatusNotImplemented)
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
