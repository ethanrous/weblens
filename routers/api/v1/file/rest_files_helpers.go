package file

import (
	"net/http"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
)

func filesInfoFromFiles(ctx context_service.RequestContext, files []*file_model.WeblensFileImpl) (resp wlstructs.FilesInfo, err error) {
	fileInfos := make([]wlstructs.FileInfo, 0, len(files))
	filePermsMap := make(map[string]*share_model.Permissions)

	medias, err := getChildMedias(ctx, files)
	if err != nil {
		return resp, wlerrors.WrapStatus(http.StatusInternalServerError, err)
	}

	mediaMap := make(map[string]*media_model.Media, len(medias))
	for _, m := range medias {
		mediaMap[m.ContentID] = m
	}

	for _, f := range files {
		var parentPerms *share_model.Permissions

		parent := f.GetParent()

		if pp, ok := filePermsMap[parent.ID()]; ok {
			parentPerms = pp
		} else {
			parentPerms, err = auth.CanUserAccessFile(ctx, ctx.Requester, parent, ctx.Share)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msgf("Failed to check permissions for file ID: %s", f.ID())

				continue
			}
		}

		fileInfo, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, f, reshape.FileInfoOptions{Perms: option.Of(*parentPerms)})
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to convert file to FileInfo for file ID: %s", f.ID())

			continue
		}

		fileInfo.HasMedia = mediaMap[fileInfo.ContentID] != nil

		fileInfos = append(fileInfos, fileInfo)
	}

	sortFileInfos(fileInfos, ctx.Query("sortProp"), ctx.Query("sortOrder"), mediaMap)

	mediaInfos := make([]wlstructs.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	resp = wlstructs.FilesInfo{Files: fileInfos, Medias: mediaInfos}

	return resp, nil
}
