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

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
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

	parentFolder := dataStore.FsTreeGet(body.ParentFolderId)
	if parentFolder == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}

	caster := NewBufferedCaster()
	defer caster.Close()

	newDir, err := dataStore.MkDir(parentFolder, body.NewFolderName, caster)
	if err != nil {
		util.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Children) != 0 {
		for _, fileId := range body.Children {
			child := dataStore.FsTreeGet(fileId)
			err = dataStore.FsTreeMove(child, newDir, "", false, caster)
			if err != nil {
				util.ShowErr(err)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id()})
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
			filteredDirInfo, err = dataStore.GetPastFileInfo(dir, acc)
			if err != nil {
				util.ShowErr(err)
				ctx.Status(http.StatusInternalServerError)
				return
			}
		} else {
			filteredDirInfo = dir.GetChildrenInfo(acc)
		}
	}

	parentsInfo := []types.FileInfo{}
	parent := dir.GetParent()
	for dataStore.CanAccessFile(parent, acc) && parent != dir && (parent.Owner() != dataStore.WeblensRootUser) {
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

	if slices.ContainsFunc(filteredDirInfo, func(i types.FileInfo) bool { return !i.Imported }) {
		c := NewBufferedCaster()
		t := dataProcess.GetGlobalQueue().ScanDirectory(dir, false, true, c)
		t.SetCleanup(func() {
			c.Close()
		})
	}

}

func getFolder(ctx *gin.Context) {
	start := time.Now()
	user := getUserFromCtx(ctx)
	//if user == nil {
	//	ctx.Status(http.StatusUnauthorized)
	//	return
	//}

	folderId := types.FileId(ctx.Param("folderId"))
	dir := dataStore.FsTreeGet(folderId)
	if dir == nil {
		util.Debug.Println("Actually not found")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	if dir.Id() == "" {
		util.Error.Println("Blank file descriptor trying to get folder info")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.Status(http.StatusNotFound)
		return
	}

	shareId := types.ShareId(ctx.Query("shareId"))
	var share types.Share
	var err error
	if shareId != "" {
		share, err = dataStore.GetShare(shareId, dataStore.FileShare)
		if err != nil || share == nil {
			util.ShowErr(err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}
	}

	acc := dataStore.NewAccessMeta(user)
	if share != nil {
		err = acc.AddShare(share)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	if !dataStore.CanAccessFile(dir, acc) {
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
	externalRoot := dataStore.GetExternalDir()
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
	dir := dataStore.FsTreeGet(folderId)
	if dir == nil {
		util.Debug.Println("Actually not found")
		time.Sleep(time.Millisecond*150 - time.Since(start))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("failed to find folder with id \"%s\"", folderId)})
		return
	}

	if dir.Id() == "" {
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

	dir := dataStore.FsTreeGet(scanInfo.FolderId)
	if dir == nil {
		return
	}
	dataProcess.GetGlobalQueue().ScanDirectory(dir, true, true, Caster)

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
	acc := dataStore.NewAccessMeta(user).SetTime(before)

	folder := dataStore.FsTreeGet(folderId)
	formatRespondFolderInfo(folder, acc, ctx)
}

func getFileHistory(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	events, err := dataStore.GetFileHistory(fileId)
	if err != nil {
		if errors.Is(err, dataStore.ErrNoFile) {
			ctx.Status(http.StatusNotFound)
		} else {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"events": events})
}

func restorePastFiles(ctx *gin.Context) {
	body, err := readCtxBody[restoreBody](ctx)
	if err != nil {
		return
	}

	t := time.UnixMilli(body.Timestamp)

	err = dataStore.RestoreFiles(body.FileIds, t)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
