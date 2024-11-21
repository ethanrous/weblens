package rest

import (
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
)

type LoginBody struct {
	Username models.Username `json:"username"`
	Password string          `json:"password"`
} // @name LoginBody

type UpdateFileParams struct {
	NewName     string          `json:"newName"`
	NewParentId fileTree.FileId `json:"newParentId"`
} // @name UpdateFileParams

type MoveFilesParams struct {
	Files       []fileTree.FileId `json:"fileIds"`
	NewParentId fileTree.FileId   `json:"newParentId"`
} // @name MoveFilesParams

type FilesListParams struct {
	FileIds []fileTree.FileId `json:"fileIds"`
} // @name FilesListParams

type MediaIdsParams struct {
	MediaIds []models.ContentId `json:"mediaIds"`
} // @name MediaIdsParams

type MediaTimeBody struct {
	AnchorId models.ContentId   `json:"anchorId"`
	NewTime  time.Time          `json:"newTime"`
	MediaIds []models.ContentId `json:"mediaIds"`
}

type NewUserParams struct {
	Username     models.Username `json:"username" validate:"required"`
	Password     string          `json:"password" validate:"required"`
	Admin        bool            `json:"admin" validate:"required"`
	AutoActivate bool            `json:"autoActivate" validate:"required"`
} // @name NewUserParams

type NewFileParams struct {
	ParentFolderId fileTree.FileId `json:"parentFolderId"`
	NewFileName    string          `json:"newFileName"`
	FileSize       int64           `json:"fileSize"`
} // @name NewFileParams

type NewFilesParams struct {
	NewFiles []NewFileParams `json:"newFiles"`
} // @name NewFilesParams

type NewUploadParams struct {
	RootFolderId    fileTree.FileId `json:"rootFolderId"`
	ChunkSize       int64           `json:"chunkSize"`
	TotalUploadSize int64           `json:"totalUploadSize"`
} // @name NewUploadParams

type PasswordUpdateParams struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword" validate:"required"`
} // @name PasswordUpdateParams

type FileShareParams struct {
	FileId   fileTree.FileId   `json:"fileId"`
	Users    []models.Username `json:"users"`
	Public   bool              `json:"public"`
	Wormhole bool              `json:"wormhole"`
} // @name FileShareParams

type AlbumShareParams struct {
	AlbumId fileTree.FileId   `json:"albumId"`
	Users   []models.Username `json:"users"`
	Public  bool              `json:"public"`
} // @name AlbumShareParams

type InitServerParams struct {
	Name string            `json:"name"`
	Role models.ServerRole `json:"role"`

	Username    models.Username      `json:"username"`
	Password    string               `json:"password"`
	CoreAddress string               `json:"coreAddress"`
	CoreKey     models.WeblensApiKey `json:"coreKey"`
	RemoteId    models.InstanceId    `json:"remoteId"`

	// For restoring a server, remoind the core of its serverId and api key the remote last used
	LocalId      models.InstanceId `json:"localId"`
	UsingKeyInfo models.ApiKey     `json:"usingKeyInfo"`
}

type NewServerParams struct {
	Id          models.InstanceId    `json:"serverId"`
	Role        models.ServerRole    `json:"role"`
	Name        string               `json:"name"`
	CoreAddress string               `json:"coreAddress"`
	UsingKey    models.WeblensApiKey `json:"usingKey"`
} // @name NewServerParams

type NewCoreBody struct {
	CoreAddress string               `json:"coreAddress"`
	UsingKey    models.WeblensApiKey `json:"usingKey"`
}

type DeleteKeyBody struct {
	Key models.WeblensApiKey `json:"key"`
}

type DeleteRemoteBody struct {
	RemoteId models.InstanceId `json:"remoteId"`
}

type RestoreBody struct {
	FileIds   []fileTree.FileId `json:"fileIds"`
	Timestamp int64             `json:"timestamp"`
}

type UploadRestoreFileBody struct {
	FileId fileTree.FileId `json:"fileId"`
}

type CreateFolderBody struct {
	ParentFolderId fileTree.FileId   `json:"parentFolderId" validate:"required"`
	NewFolderName  string            `json:"newFolderName" validate:"required"`
	Children       []fileTree.FileId `json:"children" validate:"optional"`
} // @name CreateFolderBody

type UpdateAlbumParams struct {
	AddMedia    []models.ContentId `json:"newMedia"`
	AddFolders  []fileTree.FileId  `json:"newFolders"`
	RemoveMedia []models.ContentId `json:"removeMedia"`
	Cover       models.ContentId   `json:"cover"`
	NewName     string             `json:"newName"`
	Users       []models.Username  `json:"users"`
	RemoveUsers []models.Username  `json:"removeUsers"`
} // @name UpdateAlbumParams

type CreateAlbumParams struct {
	Name string `json:"name"`
} // @name CreateAlbumParams

type UserListBody struct {
	AddUsers    []models.Username `json:"addUsers"`
	RemoveUsers []models.Username `json:"removeUsers"`
}

type SharePublicityBody struct {
	Public bool `json:"public"`
}

type ScanBody struct {
	FolderId fileTree.FileId `json:"folderId"`
	Filename string          `json:"filename"`
}

type RestoreFilesBody struct {
	FileIds     []fileTree.FileId `json:"fileIds"`
	NewParentId fileTree.FileId   `json:"newParentId"`
	Timestamp   int64             `json:"timestamp"`
} // @name RestoreFilesBody

type RestoreCoreParams struct {
	HostUrl  string `json:"restoreUrl"`
	ServerId string `json:"restoreId"`
} // @name RestoreCoreParams
