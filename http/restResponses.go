package http

import (
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
)

// FileInfo is a structure for safely sending file information to the client
type FileInfo struct {
	Id fileTree.FileId `json:"id"`
	/* If the content of the file can be displayed visually.
	Say the file is a jpg, mov, arw, etc. and not a zip,
	txt, doc, directory etc. */
	Displayable bool `json:"displayable"`

	IsDir        bool              `json:"isDir"`
	Modifiable   bool              `json:"modifiable"`
	Size         int64             `json:"size"`
	ModTime      int64             `json:"modifyTimestamp"`
	Filename     string            `json:"filename"`
	ParentId     fileTree.FileId   `json:"parentId"`
	MediaData    *models.Media     `json:"mediaData,omitempty"`
	Owner        models.Username   `json:"owner"`
	PortablePath string            `json:"portablePath"`
	ShareId      models.ShareId    `json:"shareId,omitempty"`
	Children     []fileTree.FileId `json:"children"`
	PastFile     bool              `json:"pastFile,omitempty"`
}
