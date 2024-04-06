package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

// func getFileTreeInfo(ctx *gin.Context) {
// 	user := getUserFromCtx(ctx)
// 	if !user.IsAdmin() {
// 		return
// 	}
// 	util.Debug.Println("File tree size: ", dataStore.GetTreeSize())
// }

// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir types.WeblensFile, acc types.AccessMeta, ctx *gin.Context) {
	selfData, err := dir.FormatFileInfo(acc)
	if err != nil {
		if err == dataStore.ErrNoFileAccess {
			ctx.JSON(http.StatusNotFound, "Failed to get folder info")
			return
		}
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, "Failed to get folder info")
		return
	}

	var filteredDirInfo []types.FileInfo
	if dir.IsDir() {
		filteredDirInfo = dir.GetChildrenInfo(acc)
		filteredDirInfo = util.Filter(filteredDirInfo, func(t types.FileInfo) bool { return t.Id != "R" })
	}

	parentsInfo := []types.FileInfo{}
	parent := dir.GetParent()
	for dataStore.CanAccessFile(parent, acc) && parent != dir && (parent.Owner() != dataStore.WEBLENS_ROOT_USER) {
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

}

func getFolderInfo(ctx *gin.Context) {
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

	username := types.Username(ctx.GetString("username"))
	shareId := types.ShareId(ctx.Query("shareId"))
	acc := dataStore.NewAccessMeta(username).AddShareId(shareId, dataStore.FileShare)
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
	username := types.Username(ctx.GetString("username"))
	acc := dataStore.NewAccessMeta(username).SetRequestMode(dataStore.FileGet)
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

	username := types.Username(ctx.GetString("username"))
	acc := dataStore.NewAccessMeta(username)
	formatRespondFolderInfo(dir, acc, ctx)
}

func recursiveScanDir(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var scanInfo scanInfo
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
