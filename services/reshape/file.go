package reshape

import (
	"context"

	cover_model "github.com/ethanrous/weblens/models/cover"
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/structs"
)

type FileInfoOptions struct {
	IsPastFile         bool
	ModifiableOverride option.Option[bool]
	Perms              option.Option[share_model.Permissions]
	DontCheckShare     bool // If true, we won't check if the file has a share
}

func compileOptions(opts ...FileInfoOptions) FileInfoOptions {
	var compiled FileInfoOptions

	for _, opt := range opts {
		if opt.IsPastFile {
			compiled.IsPastFile = true
		}

		if modifiableOverride, hasOverride := opt.ModifiableOverride.Get(); hasOverride {
			compiled.ModifiableOverride = option.Of(modifiableOverride)
		}

		if perms, hasPerms := opt.Perms.Get(); hasPerms {
			compiled.Perms = option.Of(perms)
		}

		if opt.DontCheckShare {
			compiled.DontCheckShare = true
		}
	}

	return compiled
}

func checkModifyable(ctx context.Context, f *file_model.WeblensFileImpl, o FileInfoOptions) bool {
	if override, ok := o.ModifiableOverride.Get(); ok {
		return override
	}

	if o.IsPastFile || file_model.IsFileInTrash(f) || !f.Exists() {
		return false
	}

	if perms, ok := o.Perms.Get(); ok {
		if !perms.HasPermission(share_model.SharePermissionEdit) {
			return false
		}
	}

	// If the file is not in the trash and exists, it is modifiable
	return true
}

func WeblensFileToFileInfo(ctx context.Context, f *file_model.WeblensFileImpl, opts ...FileInfoOptions) (structs.FileInfo, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, f)
	if err != nil {
		return structs.FileInfo{}, err
	}

	o := compileOptions(opts...)

	children := f.GetChildren()

	childrenIds := make([]string, 0, len(children))
	for _, c := range children {
		childrenIds = append(childrenIds, c.ID())
	}

	var share *share_model.FileShare
	if !o.DontCheckShare {
		share, err = share_model.GetShareByFileId(ctx, f.ID())
		if err != nil && !db.IsNotFound(err) {
			return structs.FileInfo{}, err
		}
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

	modifiable := checkModifyable(ctx, f, o)

	var hasRestoreMedia bool
	if exists || !o.IsPastFile || f.IsDir() {
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
		PastFile:        o.IsPastFile,
		PastId:          f.GetPastId(),
		PortablePath:    portablePath.String(),
		ShareId:         shareId,
		Size:            size,
	}, nil
}
