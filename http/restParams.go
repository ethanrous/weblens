package http

import (
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/types"
)

type loginBody struct {
	Username weblens.Username `json:"username"`
	Password string         `json:"password"`
}

type fileUpdateBody struct {
	NewName     string       `json:"newName"`
	NewParentId fileTree.FileId `json:"newParentId"`
}

type updateMany struct {
	Files       []fileTree.FileId `json:"fileIds"`
	NewParentId fileTree.FileId   `json:"newParentId"`
}

type takeoutFiles struct {
	FileIds []fileTree.FileId `json:"fileIds"`
}

type mediaIdsBody struct {
	MediaIds []weblens.ContentId `json:"mediaIds"`
}

type mediaTimeBody struct {
	AnchorId weblens.ContentId   `json:"anchorId"`
	NewTime  time.Time           `json:"newTime"`
	MediaIds []weblens.ContentId `json:"mediaIds"`
}

type newUserBody struct {
	Username weblens.Username `json:"username"`
	Password     string         `json:"password"`
	Admin        bool           `json:"admin"`
	AutoActivate bool           `json:"autoActivate"`
}

type newFileBody struct {
	ParentFolderId fileTree.FileId `json:"parentFolderId"`
	NewFileName    string       `json:"newFileName"`
	FileSize       int64        `json:"fileSize"`
}

type newUploadBody struct {
	RootFolderId fileTree.FileId `json:"rootFolderId"`
	ChunkSize       int64        `json:"chunkSize"`
	TotalUploadSize int64        `json:"totalUploadSize"`
}

type passwordUpdateBody struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword"`
}

type newShareBody struct {
	FileId fileTree.FileId    `json:"fileId"`
	Users  []weblens.Username `json:"users"`
	Public   bool             `json:"public"`
	Wormhole bool             `json:"wormhole"`
}

type initServerBody struct {
	Name string           `json:"name"`
	Role types.ServerRole `json:"role"`

	Username weblens.Username `json:"username"`
	Password    string              `json:"password"`
	CoreAddress string              `json:"coreAddress"`
	CoreKey     types.WeblensApiKey `json:"coreKey"`
}

type newServerBody struct {
	Id       types.InstanceId    `json:"serverId"`
	Role     types.ServerRole    `json:"role"`
	Name     string              `json:"name"`
	UsingKey types.WeblensApiKey `json:"usingKey"`
}

type deleteKeyBody struct {
	Key types.WeblensApiKey `json:"key"`
}

type deleteRemoteBody struct {
	RemoteId types.InstanceId `json:"remoteId"`
}

type restoreBody struct {
	FileIds []fileTree.FileId `json:"fileIds"`
	Timestamp int64          `json:"timestamp"`
}

type createFolderBody struct {
	ParentFolderId fileTree.FileId   `json:"parentFolderId"`
	NewFolderName  string         `json:"newFolderName"`
	Children       []fileTree.FileId `json:"children"`
}

type updateAlbumBody struct {
	AddMedia    []weblens.ContentId `json:"newMedia"`
	AddFolders  []fileTree.FileId   `json:"newFolders"`
	RemoveMedia []weblens.ContentId `json:"removeMedia"`
	Cover       weblens.ContentId   `json:"cover"`
	NewName     string              `json:"newName"`
	Users       []weblens.Username  `json:"users"`
	RemoveUsers []weblens.Username  `json:"removeUsers"`
}

type albumCreateBody struct {
	Name string `json:"name"`
}

type userListBody struct {
	Users []weblens.Username `json:"users"`
}

type scanBody struct {
	FolderId fileTree.FileId `json:"folderId"`
	Filename string          `json:"filename"`
}