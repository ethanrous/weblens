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

// FileInfoOptions configures how file information is computed and formatted.
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

func checkModifiable(_ context.Context, f *file_model.WeblensFileImpl, o FileInfoOptions) bool {
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

// WeblensFileToFileInfo converts a WeblensFileImpl to a FileInfo structure suitable for API responses.
func WeblensFileToFileInfo(ctx context.Context, f *file_model.WeblensFileImpl, opts ...FileInfoOptions) (structs.FileInfo, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, f)
	if err != nil {
		return structs.FileInfo{}, err
	}

	o := compileOptions(opts...)

	children := f.GetChildren()

	childrenIDs := make([]string, 0, len(children))
	for _, c := range children {
		childrenIDs = append(childrenIDs, c.ID())
	}

	var share *share_model.FileShare
	if !o.DontCheckShare {
		share, err = share_model.GetShareByFileID(ctx, f.ID())
		if err != nil && !db.IsNotFound(err) {
			return structs.FileInfo{}, err
		}
	}

	shareID := ""
	if share != nil && !share.ID().IsZero() {
		shareID = share.ID().Hex()
	}

	contentID := f.GetContentID()

	if f.IsDir() && contentID == "" {
		// Check if the folder has a cover photo, and use that as the content id if it does
		cover, err := cover_model.GetCoverByFolderID(ctx, f.ID())
		if err != nil && !db.IsNotFound(err) {
			return structs.FileInfo{}, err
		} else if err == nil {
			contentID = cover.CoverPhotoID
		}
	}

	exists := f.Exists()

	modifiable := checkModifiable(ctx, f, o)

	var hasRestoreMedia bool
	if exists || !o.IsPastFile || f.IsDir() {
		hasRestoreMedia = true
	}
	// ctx.(context_mod.AppContexter).AppCtx().(context.AppContext).FileService.GetFileByID(f.ID())
	// TODO: check if the file is in the restore media tree
	if err == nil {
		hasRestoreMedia = true
	}

	size := int64(0)
	if exists {
		size = f.Size()
	}

	portablePath := f.GetPortablePath()

	parentID := ""
	if f.GetParent() != nil {
		parentID = f.GetParent().ID()
	}

	return structs.FileInfo{
		Children:        childrenIDs,
		ContentID:       contentID,
		HasRestoreMedia: hasRestoreMedia,
		ID:              f.ID(),
		IsDir:           f.IsDir(),
		ModTime:         f.ModTime().UnixMilli(),
		Modifiable:      modifiable,
		Owner:           ownerName,
		ParentID:        parentID,
		PastFile:        o.IsPastFile,
		PastID:          f.GetPastID(),
		PortablePath:    portablePath.String(),
		ShareID:         shareID,
		Size:            size,
	}, nil
}
