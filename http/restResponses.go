package http

import (
	"errors"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
)

// FileInfo is a structure for safely sending file information to the client
type FileInfo struct {
	Id           fileTree.FileId   `json:"id"`
	PortablePath string            `json:"portablePath"`
	Filename     string            `json:"filename"`
	Size         int64             `json:"size"`
	IsDir        bool              `json:"isDir"`
	ModTime      int64             `json:"modifyTimestamp"`
	ParentId     fileTree.FileId   `json:"parentId"`
	Children     []fileTree.FileId `json:"childrenIds,omitempty"`
	ContentId    models.ContentId  `json:"contentId"`
	// PastFile     bool              `json:"pastFile"`
	Owner      models.Username `json:"owner"`
	ShareId    models.ShareId  `json:"shareId,omitempty"`
	Modifiable bool            `json:"modifiable"`
}

func WeblensFileToFileInfo(f *fileTree.WeblensFileImpl, pack *models.ServicePack, isParent bool) (FileInfo, error) {
	// Some fields are only needed if the file is the parent file of the request,
	// when the file is a child, these fields are not needed, and can be expensive to fetch,
	// so we conditionally ignore them.
	var owner models.Username
	var children []fileTree.FileId
	if isParent {
		owner = pack.FileService.GetFileOwner(f).GetUsername()
		for _, c := range f.GetChildren() {
			children = append(children, c.ID())
		}
	}

	share, err := pack.ShareService.GetFileShare(f)
	if err != nil && !errors.Is(err, werror.ErrNoShare) {
		return FileInfo{}, err
	}
	var shareId models.ShareId
	if share != nil {
		shareId = share.ID()
	}

	if f.IsDir() && f.GetContentId() == "" {
		_, err = pack.FileService.GetFolderCover(f)
		if err != nil {
			return FileInfo{}, err
		}
	}

	modifiable := !pack.FileService.IsFileInTrash(f)

	return FileInfo{
		Id:           f.ID(),
		PortablePath: f.GetPortablePath().ToPortable(),
		Size:         f.Size(),
		IsDir:        f.IsDir(),
		ModTime:      f.ModTime().UnixMilli(),
		ParentId:     f.GetParentId(),
		ContentId:    f.GetContentId(),
		ShareId:      shareId,
		Modifiable:   modifiable,

		Owner:    owner,
		Children: children,
	}, nil
}
