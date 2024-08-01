package routes

import (
	"time"

	"github.com/ethrousseau/weblens/api/types"
)

type loginBody struct {
	Username types.Username `json:"username"`
	Password string         `json:"password"`
}

type fileUpdateBody struct {
	NewName     string       `json:"newName"`
	NewParentId types.FileId `json:"newParentId"`
}

type updateMany struct {
	Files       []types.FileId `json:"fileIds"`
	NewParentId types.FileId   `json:"newParentId"`
}

type takeoutFiles struct {
	FileIds []types.FileId `json:"fileIds"`
}

type mediaIdsBody struct {
	MediaIds []types.ContentId `json:"mediaIds"`
}

type mediaTimeBody struct {
	AnchorId types.ContentId   `json:"anchorId"`
	NewTime  time.Time         `json:"newTime"`
	MediaIds []types.ContentId `json:"mediaIds"`
}

type newUserBody struct {
	Username     types.Username `json:"username"`
	Password     string         `json:"password"`
	Admin        bool           `json:"admin"`
	AutoActivate bool           `json:"autoActivate"`
}

type newFileBody struct {
	ParentFolderId types.FileId `json:"parentFolderId"`
	NewFileName    string       `json:"newFileName"`
	FileSize       int64        `json:"fileSize"`
}

type newUploadBody struct {
	RootFolderId    types.FileId `json:"rootFolderId"`
	ChunkSize       int64        `json:"chunkSize"`
	TotalUploadSize int64        `json:"totalUploadSize"`
}

type passwordUpdateBody struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword"`
}

type newShareBody struct {
	FileId   types.FileId     `json:"fileId"`
	Users    []types.Username `json:"users"`
	Public   bool             `json:"public"`
	Wormhole bool             `json:"wormhole"`
}

type initServerBody struct {
	Name string           `json:"name"`
	Role types.ServerRole `json:"role"`

	Username    types.Username      `json:"username"`
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
	FileIds   []types.FileId `json:"fileIds"`
	Timestamp int64          `json:"timestamp"`
}

type createFolderBody struct {
	ParentFolderId types.FileId   `json:"parentFolderId"`
	NewFolderName  string         `json:"newFolderName"`
	Children       []types.FileId `json:"children"`
}

type updateAlbumBody struct {
	AddMedia    []types.ContentId `json:"newMedia"`
	AddFolders  []types.FileId    `json:"newFolders"`
	RemoveMedia []types.ContentId `json:"removeMedia"`
	Cover       types.ContentId   `json:"cover"`
	NewName     string            `json:"newName"`
	Users       []types.Username  `json:"users"`
	RemoveUsers []types.Username  `json:"removeUsers"`
}

type albumCreateBody struct {
	Name string `json:"name"`
}

type userListBody struct {
	Users []types.Username `json:"users"`
}
