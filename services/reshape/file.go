package reshape

import (
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/structs"
	file_service "github.com/ethanrous/weblens/services/file"
)

func WeblensFileToFileInfo(ctx context_mod.ContextZ, f *file_model.WeblensFileImpl, isPastFile bool) (structs.FileInfo, error) {

	// Some fields are only needed if the file is the parent file of the request,
	// when the file is a child, these fields are not needed, and can be expensive to fetch,
	// so we conditionally ignore them.
	var children []string

	ownerName, err := file_service.GetFileOwnerName(ctx, f)
	if err != nil {
		return structs.FileInfo{}, err
	}

	for _, c := range f.GetChildren() {
		children = append(children, c.ID())
	}

	share, err := share_model.GetShareByFileId(ctx, f.ID())
	if err != nil && !db.IsNotFound(err) {
		return structs.FileInfo{}, err
	}
	var shareId string
	if share != nil {
		shareId = share.ID()
	}

	if f.IsDir() && f.GetContentId() == "" {
		// TODO: check if the folder has a cover
		// cover, err := cover_model.GetCoverByFolderId(ctx, f.ID())
		// if err != nil {
		// 	return structs.FileInfo{}, err
		// }

		// cover.CoverPhotoId
	}

	modifiable := !isPastFile && !file_service.IsFileInTrash(f)

	var hasRestoreMedia bool
	if !isPastFile || f.IsDir() || f.Exists() {
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

	return structs.FileInfo{
		Id:              f.ID(),
		PortablePath:    f.GetPortablePath().ToPortable(),
		Filename:        f.GetPortablePath().Filename(),
		Size:            f.Size(),
		IsDir:           f.IsDir(),
		ModTime:         f.ModTime().UnixMilli(),
		ParentId:        f.GetParentId(),
		ContentId:       f.GetContentId(),
		ShareId:         shareId,
		Modifiable:      modifiable,
		PastFile:        isPastFile,
		HasRestoreMedia: hasRestoreMedia,
		PastId:          f.GetPastId(),

		Owner:    ownerName,
		Children: children,
	}, nil
}
