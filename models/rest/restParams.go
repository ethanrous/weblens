package rest

import (
	"time"

	"github.com/ethanrous/weblens/models"
)

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name LoginBody

type UpdateFileParams struct {
	NewName     string `json:"newName"`
	NewParentId string `json:"newParentId"`
} // @name UpdateFileParams

type MoveFilesParams struct {
	NewParentId string   `json:"newParentId"`
	Files       []string `json:"fileIds"`
} // @name MoveFilesParams

type FilesListParams struct {
	FileIds []string `json:"fileIds"`
} // @name FilesListParams

type MediaIdsParams struct {
	MediaIds []string `json:"mediaIds"`
} // @name MediaIdsParams

type MediaTimeBody struct {
	AnchorId string    `json:"anchorId"`
	NewTime  time.Time `json:"newTime"`
	MediaIds []string  `json:"mediaIds"`
}

type NewUserParams struct {
	FullName     string `json:"fullName" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
	Admin        bool   `json:"admin"`
	AutoActivate bool   `json:"autoActivate"`
} // @name NewUserParams

type NewFileParams struct {
	ParentFolderId string `json:"parentFolderId"`
	NewFileName    string `json:"newFileName"`
	FileSize       int64  `json:"fileSize"`
	IsDir          bool   `json:"isDir"`
} // @name NewFileParams

type NewFilesParams struct {
	NewFiles []NewFileParams `json:"newFiles"`
} // @name NewFilesParams

type NewUploadParams struct {
	RootFolderId string `json:"rootFolderId"`
	ChunkSize    int64  `json:"chunkSize"`
} // @name NewUploadParams

type PasswordUpdateParams struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword" validate:"required"`
} // @name PasswordUpdateParams

type FileShareParams struct {
	FileId   string   `json:"fileId"`
	Users    []string `json:"users"`
	Public   bool     `json:"public"`
	Wormhole bool     `json:"wormhole"`
} // @name FileShareParams

type AlbumShareParams struct {
	AlbumId string   `json:"albumId"`
	Users   []string `json:"users"`
	Public  bool     `json:"public"`
} // @name AlbumShareParams

type InitServerParams struct {
	Name string            `json:"name"`
	Role models.ServerRole `json:"role"`

	Username    string               `json:"username"`
	Password    string               `json:"password"`
	FullName    string               `json:"fullName"`
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

type UpdateServerParams struct {
	Name        string               `json:"name"`
	CoreAddress string               `json:"coreAddress"`
	UsingKey    models.WeblensApiKey `json:"usingKey"`
} // @name UpdateServerParams

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
	FileIds   []string `json:"fileIds"`
	Timestamp int64    `json:"timestamp"`
}

type UploadRestoreFileBody struct {
	FileId string `json:"fileId"`
}

type CreateFolderBody struct {
	ParentFolderId string   `json:"parentFolderId" validate:"required"`
	NewFolderName  string   `json:"newFolderName" validate:"required"`
	Children       []string `json:"children" validate:"optional"`
} // @name CreateFolderBody

type UpdateAlbumParams struct {
	AddMedia    []string `json:"newMedia"`
	AddFolders  []string `json:"newFolders"`
	RemoveMedia []string `json:"removeMedia"`
	Cover       string   `json:"cover"`
	NewName     string   `json:"newName"`
	Users       []string `json:"users"`
	RemoveUsers []string `json:"removeUsers"`
} // @name UpdateAlbumParams

type CreateAlbumParams struct {
	Name string `json:"name"`
} // @name CreateAlbumParams

type UserListBody struct {
	AddUsers    []string `json:"addUsers"`
	RemoveUsers []string `json:"removeUsers"`
}

type SharePublicityBody struct {
	Public bool `json:"public"`
}

type ScanBody struct {
	FolderId string `json:"folderId"`
	Filename string `json:"filename"`
}

type RestoreFilesBody struct {
	NewParentId string   `json:"newParentId"`
	FileIds     []string `json:"fileIds"`
	Timestamp   int64    `json:"timestamp"`
} // @name RestoreFilesBody

type RestoreCoreParams struct {
	HostUrl  string `json:"restoreUrl"`
	ServerId string `json:"restoreId"`
} // @name RestoreCoreParams

type ApiKeyParams struct {
	Name string `json:"name" validate:"required"`
} // @name ApiKeyParams
