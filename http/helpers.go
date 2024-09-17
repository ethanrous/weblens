package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func getServices(ctx *gin.Context) *models.ServicePack {
	srv, ok := ctx.Get("services")
	if !ok {
		return nil
	}
	return srv.(*models.ServicePack)
}

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

func getInstanceFromCtx(ctx *gin.Context) *models.Instance {
	server, ok := ctx.Get("server")
	if !ok {
		return nil
	}
	return server.(*models.Instance)
}

func getShareFromCtx[T models.Share](ctx *gin.Context) (T, error) {
	pack := getServices(ctx)

	shareId := models.ShareId(ctx.Query("shareId"))
	if shareId == "" {
		shareId = models.ShareId(ctx.Param("shareId"))
	}
	var empty T
	if shareId == "" {
		return empty, nil
	}

	sh := pack.ShareService.Get(shareId)
	tsh, ok := sh.(T)
	if sh != nil && ok {
		return tsh, nil
	}

	err := errors.New("Could not find valid share")
	ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	return empty, err
}

func formatFileSafe(
	f *fileTree.WeblensFileImpl, accessor *models.User, share *models.FileShare, pack *models.ServicePack,
) (
	formattedInfo FileInfo,
	err error,
) {
	if f == nil {
		return formattedInfo, werror.WithStack(errors.New("cannot get file info of nil wf"))
	}

	if !pack.AccessService.CanUserAccessFile(accessor, f, share) {
		err = werror.ErrNoFileAccess
		return
	}

	var size int64
	size = f.Size()

	var parentId fileTree.FileId
	owner := pack.FileService.GetFileOwner(f)
	if f.GetParentId() != "ROOT" && pack.AccessService.CanUserAccessFile(accessor, f.GetParent(), share) {
		parentId = f.GetParent().ID()
	}

	tmpF := f
	var pathBits []string
	for tmpF != nil && tmpF.ID() != "ROOT" && pack.AccessService.CanUserAccessFile(
		accessor, tmpF, share,
	) {
		if tmpF.GetParent() == pack.FileService.GetMediaRoot() {
			pathBits = append(pathBits, "HOME")
			break
		} else if share != nil && tmpF.ID() == share.GetItemId() {
			pathBits = append(pathBits, "SHARE")
			break
		} else if pack.FileService.IsFileInTrash(tmpF) {
			pathBits = append(pathBits, "TRASH")
			break
		}
		pathBits = append(pathBits, tmpF.Filename())
		tmpF = tmpF.GetParent()
	}
	slices.Reverse(pathBits)

	fShare, _ := pack.ShareService.GetFileShare(f)
	var shareId models.ShareId
	if fShare != nil {
		shareId = fShare.ID()
	}

	formattedInfo = FileInfo{
		Id:          f.ID(),
		Displayable: pack.MediaService.IsFileDisplayable(f),
		IsDir:       f.IsDir(),
		Modifiable: !pack.FileService.IsFileInTrash(f) &&
			owner == accessor &&
			pack.FileService.GetFileOwner(f) != pack.UserService.GetRootUser() &&
			pack.InstanceService.GetLocal().GetRole() != models.BackupServer,
		Size:         size,
		ModTime:      f.ModTime().UnixMilli(),
		Filename:     f.Filename(),
		ParentId:     parentId,
		Owner:        owner.GetUsername(),
		PortablePath: f.GetPortablePath().ToPortable(),
		MediaData:    pack.MediaService.Get(f.GetContentId()),
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
	ModTime time.Time `json:"modifyTimestamp"`
	Exists  bool      `json:"exists"`
}
