package comm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
	if ctx.Request.Method == "GET" {
		err = errors.New("trying to get body of get request")
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is not in expected JSON format"})
		return
	}

	return
}

func readRespBody[T any](resp *http.Response) (obj T, err error) {
	var bodyB []byte
	if resp.ContentLength == 0 {
		return obj, werror.ErrNoBody
	} else if resp.ContentLength == -1 {
		log.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyB, &obj)
	return
}

func readRespBodyRaw(resp *http.Response) (bodyB []byte, err error) {
	if resp.ContentLength == 0 {
		return nil, werror.ErrNoBody
	} else if resp.ContentLength == -1 {
		log.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	return
}

func getUserFromCtx(ctx *gin.Context) *models.User {
	user, ok := ctx.Get("user")
	if !ok {
		return nil
	}
	return user.(*models.User)
}

func getRemoteFromCtx(ctx *gin.Context) *models.Instance {
	instance, ok := ctx.Get("instance")
	if !ok {
		return nil
	}
	return instance.(*models.Instance)
}

func getShareFromCtx[T models.Share](ctx *gin.Context) (T, error) {
	shareId := models.ShareId(ctx.Query("shareId"))
	if shareId == "" {
		shareId = models.ShareId(ctx.Param("shareId"))
	}
	var empty T
	if shareId == "" {
		return empty, nil
	}

	sh := ShareService.Get(shareId)
	tsh, ok := sh.(T)
	if sh != nil && ok {
		return tsh, nil
	}

	err := errors.New("Could not find valid share")
	ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	return empty, err
}

func formatFileSafe(f *fileTree.WeblensFileImpl, accessor *models.User, share *models.FileShare) (
	formattedInfo FileInfo,
	err error,
) {
	if f == nil {
		return formattedInfo, werror.WithStack(errors.New("cannot get file info of nil wf"))
	}

	if !AccessService.CanUserAccessFile(accessor, f, share) {
		err = werror.ErrNoFileAccess
		return
	}

	var size int64
	size, err = f.Size()
	if err != nil {
		log.ShowErr(err, fmt.Sprintf("Failed to get file size of [ %s (ID: %s) ]", f.GetAbsPath(), f.ID()))
		return
	}

	var parentId fileTree.FileId
	owner := FileService.GetFileOwner(f)
	if f.GetParentId() != "ROOT" && AccessService.CanUserAccessFile(accessor, f.GetParent(), share) {
		parentId = f.GetParent().ID()
	}

	tmpF := f
	var pathBits []string
	for tmpF != nil && tmpF.ID() != "ROOT" && AccessService.CanUserAccessFile(
		accessor, tmpF, share,
	) {
		if tmpF.GetParent() == FileService.GetMediaRoot() {
			pathBits = append(pathBits, "HOME")
			break
		} else if share != nil && tmpF.ID() == fileTree.FileId(share.GetItemId()) {
			pathBits = append(pathBits, "SHARE")
			break
		} else if FileService.IsFileInTrash(tmpF) {
			pathBits = append(pathBits, "TRASH")
			break
		}
		pathBits = append(pathBits, tmpF.Filename())
		tmpF = tmpF.GetParent()
	}
	slices.Reverse(pathBits)
	pathString := strings.Join(pathBits, "/")

	var shareId models.ShareId
	if share != nil {
		shareId = share.GetShareId()
	}

	formattedInfo = FileInfo{
		Id:          f.ID(),
		Displayable: MediaService.IsFileDisplayable(f),
		IsDir:       f.IsDir(),
		Modifiable: !FileService.IsFileInTrash(f) &&
			owner == accessor &&
			FileService.GetFileOwner(f) != UserService.GetRootUser() &&
			InstanceService.GetLocal().ServerRole() != models.BackupServer,
		Size:         size,
		ModTime:      f.ModTime().UnixMilli(),
		Filename:     f.Filename(),
		ParentId:     parentId,
		Owner:        owner.GetUsername(),
		PathFromHome: pathString,
		MediaData:    MediaService.Get(models.ContentId(f.GetContentId())),
		ShareId:      shareId,
		Children: internal.Map(
			f.GetChildren(), func(wf *fileTree.WeblensFileImpl) fileTree.FileId { return wf.ID() },
		),
	}

	return formattedInfo, nil
}

type FileStat struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Exists  bool      `json:"exists"`
}
