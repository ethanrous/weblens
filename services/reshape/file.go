package reshape

import (
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services"
	"github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
)

func WeblensFileToFileInfo(ctx *context.RequestContext, f *file.WeblensFileImpl, isPastFile bool) (structs.FileInfo, error) {
	// Some fields are only needed if the file is the parent file of the request,
	// when the file is a child, these fields are not needed, and can be expensive to fetch,
	// so we conditionally ignore them.
	var children []string
	owner, err := pack.FileService.GetFileOwner(f)
	if err != nil {
		return structs.FileInfo{}, err
	}

	for _, c := range f.GetChildren() {
		children = append(children, c.ID())
	}

	share, err := pack.ShareService.GetFileShare(f.ID())
	if err != nil && !errors.Is(err, share_model.ErrShareNotFound) {
		return structs.FileInfo{}, err
	}
	var shareId models.ShareId
	if share != nil {
		shareId = share.ID()
	}

	if f.IsDir() && f.GetContentId() == "" {
		_, err = pack.FileService.GetFolderCover(f)
		if err != nil {
			return structs.FileInfo{}, err
		}
	}

	modifiable := !isPastFile && !pack.FileService.IsFileInTrash(f)

	var hasRestoreMedia bool
	if !isPastFile || f.IsDir() || f.Exists() {
		hasRestoreMedia = true
	} else {
		_, err := pack.FileService.GetFileByTree(f.ID(), services.UsersTreeKey)
		if err == nil {
			hasRestoreMedia = true
		} else {
			restoreTree, err := pack.FileService.GetFileTreeByName(services.RestoreTreeKey)
			if err != nil {
				return structs.FileInfo{}, err
			}

			_, err = restoreTree.GetRoot().GetChild(f.GetContentId())
			hasRestoreMedia = err == nil
		}
	}

	return structs.FileInfo{
		Id:              f.ID(),
		PortablePath:    f.GetPortablePath().ToPortable(),
		Filename:        f.Filename(),
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

		Owner:    owner.Username,
		Children: children,
	}, nil
}
