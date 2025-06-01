package reshape

import (
	"context"

	cover_model "github.com/ethanrous/weblens/models/cover"
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/structs"
)

func WeblensFileToFileInfo(ctx context.Context, f *file_model.WeblensFileImpl, isPastFile bool) (structs.FileInfo, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, f)
	if err != nil {
		return structs.FileInfo{}, err
	}

	children := f.GetChildren()

	childrenIds := make([]string, 0, len(children))
	for _, c := range children {
		childrenIds = append(childrenIds, c.ID())
	}

	share, err := share_model.GetShareByFileId(ctx, f.ID())
	if err != nil && !db.IsNotFound(err) {
		return structs.FileInfo{}, err
	}

	shareId := ""
	if share != nil && !share.ID().IsZero() {
		shareId = share.ID().Hex()
	}

	contentId := f.GetContentId()

	if f.IsDir() && contentId == "" {
		// Check if the folder has a cover photo, and use that as the content id if it does
		cover, err := cover_model.GetCoverByFolderId(ctx, f.ID())
		if err != nil && !db.IsNotFound(err) {
			return structs.FileInfo{}, err
		} else if err == nil {
			contentId = cover.CoverPhotoId
		}
	}

	exists := f.Exists()

	modifiable := !isPastFile && !file_model.IsFileInTrash(f) && exists

	var hasRestoreMedia bool
	if exists || !isPastFile || f.IsDir() {
		hasRestoreMedia = true
	} else {
		// ctx.(context_mod.AppContexter).AppCtx().(context.AppContext).FileService.GetFileById(f.ID())
		// TODO: check if the file is in the restore media tree
		if err == nil {
			hasRestoreMedia = true
		} else {
			// restoreTree, err := pack.FileService.GetFileTreeByName(services.RestoreTreeKey)
			// if err != nil {
			// 	return structs.FileInfo{}, err
			// }
			//
			// _, err = restoreTree.GetRoot().GetChild(f.GetContentId())
			// hasRestoreMedia = err == nil
		}
	}

	size := int64(0)
	if exists {
		size = f.Size()
	}

	portablePath := f.GetPortablePath()
	// if !exists {
	// 	portablePath = fs.Filepath{}
	// }
	parentId := ""
	if f.GetParent() != nil {
		parentId = f.GetParent().ID()
	}

	return structs.FileInfo{
		Children:        childrenIds,
		ContentId:       contentId,
		HasRestoreMedia: hasRestoreMedia,
		Id:              f.ID(),
		IsDir:           f.IsDir(),
		ModTime:         f.ModTime().UnixMilli(),
		Modifiable:      modifiable,
		Owner:           ownerName,
		ParentId:        parentId,
		PastFile:        isPastFile,
		PastId:          f.GetPastId(),
		PortablePath:    portablePath.String(),
		ShareId:         shareId,
		Size:            size,
	}, nil
}
