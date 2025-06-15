package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/rs/zerolog"
)

// Format and write back directory information. Authorization checks should be done before this function.
func formatRespondFolderInfo(ctx context_service.RequestContext, dir *file_model.WeblensFileImpl) error {
	if dir == nil {
		ctx.Error(http.StatusNotFound, file_model.ErrFileNotFound)

		return nil
	}

	parentsInfo := []structs.FileInfo{}
	parent := dir.GetParent()

	owner, err := file_service.GetFileOwner(ctx, dir)
	if err != nil {
		return err
	}

	for parent != nil && !parent.GetPortablePath().IsRoot() && !owner.IsSystemUser() {
		var perms *share_model.Permissions
		if perms, err = auth.CanUserAccessFile(ctx, ctx.Requester, parent, ctx.Share); err != nil {
			break
		}

		parentInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, parent, reshape.FileInfoOptions{Perms: option.Of(*perms)})
		if err != nil {
			return err
		}

		parentsInfo = append(parentsInfo, parentInfo)

		parent = parent.GetParent()
	}

	perms, err := auth.CanUserAccessFile(ctx, ctx.Requester, dir, ctx.Share)
	if err != nil {
		ctx.Error(http.StatusForbidden, err)

		return err
	}

	children, err := ctx.FileService.GetChildren(ctx, dir)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return err
	}

	mediaFiles := append(children, dir)

	medias, err := getChildMedias(ctx, mediaFiles)
	if err != nil {
		return err
	}

	childInfos := make([]structs.FileInfo, 0, len(children))
	infoOpts := reshape.FileInfoOptions{Perms: option.Of(*perms)}

	for _, child := range children {
		if child == nil {
			continue
		}

		info, err := reshape.WeblensFileToFileInfo(ctx, child, infoOpts)
		if err != nil {
			return err
		}

		childInfos = append(childInfos, info)
	}

	selfInfo, err := reshape.WeblensFileToFileInfo(ctx, dir, infoOpts)
	if err != nil {
		return err
	}

	mediaInfos := make([]structs.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := structs.FolderInfoResponse{Self: selfInfo, Children: childInfos, Parents: parentsInfo, Medias: mediaInfos}
	ctx.JSON(http.StatusOK, packagedInfo)

	return nil
}

func formatRespondPastFolderInfo(ctx context_service.RequestContext, folder *file_model.WeblensFileImpl, pastTime time.Time) error {
	ctx.Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Getting past folder [%s] at time [%s]", folder.GetPortablePath(), pastTime)
	})

	pastFileInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, folder, reshape.FileInfoOptions{IsPastFile: true})
	if err != nil {
		return err
	}

	parentsInfo := []structs.FileInfo{}

	pastParent := folder.GetParent()
	for pastParent != nil && !pastParent.GetPortablePath().IsRoot() {
		parentInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, pastParent, reshape.FileInfoOptions{IsPastFile: true})
		if err != nil {
			return err
		}

		parentsInfo = append(parentsInfo, parentInfo)
		pastParent = pastParent.GetParent()
	}

	children := folder.GetChildren()
	childrenInfos := make([]structs.FileInfo, 0, len(children))

	for _, child := range children {
		if child.GetPortablePath().Filename() == file_model.UserTrashDirName {
			continue
		}

		childInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, child, reshape.FileInfoOptions{IsPastFile: true})
		if err != nil {
			return err
		}

		childrenInfos = append(childrenInfos, childInfo)
	}

	medias := []*media_model.Media{}

	for _, child := range children {
		m, err := media_model.GetMediaByContentId(ctx, child.GetContentId())

		if err == nil {
			medias = append(medias, m)
		}
	}

	mediaInfos := make([]structs.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := structs.FolderInfoResponse{
		Self:     pastFileInfo,
		Children: childrenInfos,
		Parents:  parentsInfo,
		Medias:   mediaInfos,
	}
	ctx.JSON(http.StatusOK, packagedInfo)

	return nil
}
