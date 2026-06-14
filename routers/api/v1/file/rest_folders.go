package file

import (
	"net/http"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/embed"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/rs/zerolog"
)

// ScanDir godoc
//
//	@ID	ScanFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Dispatch a folder scan
//	@Tags		Folder
//	@Param		folderID		path		string				true	"Folder ID"
//	@Param		shareID			query		string				false	"Share ID"
//	@Param		forceReindex	query		bool				false	"Force a full re-index, rebuilding media and embeddings"
//	@Success	200				{object}	wlstructs.TaskInfo	"Task Info"
//	@Failure	404
//	@Failure	500
//	@Router		/folder/{folderID}/scan [post]
func ScanDir(ctx context_service.RequestContext) {
	folder := ctx.File

	meta := job.IndexMeta{
		File:         folder,
		ForceReIndex: ctx.QueryBool("forceReindex"),
	}

	var jobName string
	if folder.IsDir() {
		jobName = job.ScanDirectoryTask
	} else {
		jobName = job.IndexFileTask
	}

	t, err := ctx.TaskService.DispatchJob(ctx, jobName, meta, nil)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Errorf("Failed to dispatch scan task: %w", err))

		return
	}

	taskInfo := reshape.TaskToTaskInfo(t)

	ctx.JSON(http.StatusOK, taskInfo)
}

// Format and write back directory information. Authorization checks should be done before this function.
func formatRespondFolderInfo(ctx context_service.RequestContext, dir *file_model.WeblensFileImpl) error {
	if dir == nil {
		ctx.Error(http.StatusNotFound, file_model.ErrFileNotFound)

		return nil
	}

	parentsInfo := []wlstructs.FileInfo{}
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

	medias, err := getChildMedias(ctx, children)
	if err != nil {
		return err
	}

	mediaMap := make(map[string]*media_model.Media, len(medias))
	for _, m := range medias {
		mediaMap[m.ContentID] = m
	}

	childInfos := make([]wlstructs.FileInfo, 0, len(children))
	infoOpts := reshape.FileInfoOptions{Perms: option.Of(*perms)}

	var scanTask *task.Task

	for _, child := range children {
		if child == nil {
			continue
		}

		info, err := reshape.WeblensFileToFileInfo(ctx, child, infoOpts)
		if err != nil {
			return err
		}

		mediaExists := false
		_, mediaExists = mediaMap[info.ContentID]

		if mediaExists {
			info.HasMedia = true
		}

		if scanTask == nil && !child.IsDir() {
			shouldLaunchIndex := !mediaExists
			if !shouldLaunchIndex && media_model.EmbedEligible(child.GetPortablePath().Ext()) {
				isEmbedded, err := embed.IsFileEmbedded(ctx, child)
				if err != nil {
					return err
				}

				shouldLaunchIndex = !isEmbedded
			}

			if shouldLaunchIndex {
				ctx.Log().Debug().Msgf("Dispatching scan task for parent folder [%s] since child [%s] is missing media or index", dir.GetPortablePath(), child.GetPortablePath())

				t, err := ctx.TaskService.DispatchJob(ctx, job.ScanDirectoryTask, job.IndexMeta{File: dir}, nil)
				if err != nil {
					ctx.Log().Error().Err(err).Msgf("Failed to dispatch scan task for file [%s]", dir.GetPortablePath())
				}

				scanTask = t
			}
		}

		childInfos = append(childInfos, info)
	}

	sortFileInfos(childInfos, ctx.Query("sortProp"), ctx.Query("sortOrder"), mediaMap)

	selfInfo, err := reshape.WeblensFileToFileInfo(ctx, dir, infoOpts)
	if err != nil {
		return err
	}

	mediaInfos := make([]wlstructs.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := wlstructs.FolderInfoResponse{Self: selfInfo, Children: childInfos, Parents: parentsInfo, Medias: mediaInfos}
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

	parentsInfo := []wlstructs.FileInfo{}

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
	childrenInfos := make([]wlstructs.FileInfo, 0, len(children))

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
		m, err := media_model.GetMediaByContentID(ctx, child.GetContentID())
		if err == nil {
			medias = append(medias, m)
		}
	}

	mediaInfos := make([]wlstructs.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := wlstructs.FolderInfoResponse{
		Self:     pastFileInfo,
		Children: childrenInfos,
		Parents:  parentsInfo,
		Medias:   mediaInfos,
	}
	ctx.JSON(http.StatusOK, packagedInfo)

	return nil
}
