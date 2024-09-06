package http

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
)

type loginBody struct {
	Username models.Username `json:"username"`
	Password string          `json:"password"`
}

type fileUpdateBody struct {
	NewName     string          `json:"newName"`
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
	MediaIds []models.ContentId `json:"mediaIds"`
}

type mediaTimeBody struct {
	AnchorId models.ContentId   `json:"anchorId"`
	NewTime  time.Time          `json:"newTime"`
	MediaIds []models.ContentId `json:"mediaIds"`
}

type newUserBody struct {
	Username     models.Username `json:"username"`
	Password     string          `json:"password"`
	Admin        bool            `json:"admin"`
	AutoActivate bool            `json:"autoActivate"`
}

type newFileBody struct {
	ParentFolderId fileTree.FileId `json:"parentFolderId"`
	NewFileName    string          `json:"newFileName"`
	FileSize       int64           `json:"fileSize"`
}

type newUploadBody struct {
	RootFolderId    fileTree.FileId `json:"rootFolderId"`
	ChunkSize       int64           `json:"chunkSize"`
	TotalUploadSize int64           `json:"totalUploadSize"`
}

type passwordUpdateBody struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword"`
}

type newShareBody struct {
	FileId   fileTree.FileId   `json:"fileId"`
	Users    []models.Username `json:"users"`
	Public   bool              `json:"public"`
	Wormhole bool              `json:"wormhole"`
}

type initServerBody struct {
	Name string            `json:"name"`
	Role models.ServerRole `json:"role"`

	Username    models.Username      `json:"username"`
	Password    string               `json:"password"`
	CoreAddress string               `json:"coreAddress"`
	CoreKey     models.WeblensApiKey `json:"coreKey"`
}

type newServerBody struct {
	Id       models.InstanceId    `json:"serverId"`
	Role     models.ServerRole    `json:"role"`
	Name     string               `json:"name"`
	UsingKey models.WeblensApiKey `json:"usingKey"`
}

type deleteKeyBody struct {
	Key models.WeblensApiKey `json:"key"`
}

type deleteRemoteBody struct {
	RemoteId models.InstanceId `json:"remoteId"`
}

type restoreBody struct {
	FileIds   []fileTree.FileId `json:"fileIds"`
	Timestamp int64             `json:"timestamp"`
}

type createFolderBody struct {
	ParentFolderId fileTree.FileId   `json:"parentFolderId"`
	NewFolderName  string            `json:"newFolderName"`
	Children       []fileTree.FileId `json:"children"`
}

type updateAlbumBody struct {
	AddMedia    []models.ContentId `json:"newMedia"`
	AddFolders  []fileTree.FileId  `json:"newFolders"`
	RemoveMedia []models.ContentId `json:"removeMedia"`
	Cover       models.ContentId   `json:"cover"`
	NewName     string             `json:"newName"`
	Users       []models.Username  `json:"users"`
	RemoveUsers []models.Username  `json:"removeUsers"`
}

type albumCreateBody struct {
	Name string `json:"name"`
}

type userListBody struct {
	AddUsers    []models.Username `json:"addUsers"`
	RemoveUsers []models.Username `json:"removeUsers"`
}

type sharePublicityBody struct {
	Public bool `json:"public"`
}

type scanBody struct {
	FolderId fileTree.FileId `json:"folderId"`
	Filename string          `json:"filename"`
}
