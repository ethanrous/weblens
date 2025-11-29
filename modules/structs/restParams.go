package structs

import (
	"time"
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
	FileId       string   `json:"fileId"`
	Users        []string `json:"users"`
	Public       bool     `json:"public"`
	Wormhole     bool     `json:"wormhole"`
	TimelineOnly bool     `json:"timelineOnly"`
} // @name FileShareParams

type AlbumShareParams struct {
	AlbumId string   `json:"albumId"`
	Users   []string `json:"users"`
	Public  bool     `json:"public"`
} // @name AlbumShareParams

type DeleteKeyBody struct {
	Key string `json:"key"`
}

type DeleteRemoteBody struct {
	RemoteId string `json:"remoteId"`
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

type PermissionsParams struct {
	CanView     bool `json:"canView"`
	CanEdit     bool `json:"canEdit"`
	CanDownload bool `json:"canDownload"`
	CanDelete   bool `json:"canDelete"`
} // @name PermissionsParams

type AddUserParams struct {
	Username string `json:"username" validate:"required"`
	PermissionsParams
} // @name AddUserParams

type UpdateUsersPermissionsParams struct {
	AddUsers    map[string]PermissionsParams `json:"addUsers"`
	RemoveUsers map[string]PermissionsParams `json:"removeUsers"`
} // @name UpdateUsersPermissionsParams

type SharePublicityBody struct {
	Public bool `json:"public"`
}

type ScanBody struct {
	FolderId string `json:"folderId"`
	Filename string `json:"filename"`
}

type RestoreFilesParams struct {
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

type MediaBatchParams struct {
	Raw           bool     `json:"raw" query:"raw" example:"false" enums:"true,false"`
	Hidden        bool     `json:"hidden" query:"hidden" example:"false" enums:"true,false"`
	Sort          string   `json:"sort" query:"sort" example:"createDate" enums:"createDate"`
	SortDirection int      `json:"sortDirection" example:"1" enums:"1,-1"`
	Search        string   `json:"search" query:"search" example:""`
	Page          int      `json:"page" query:"page" example:"1"`
	Limit         int      `json:"limit" query:"limit" example:"20"`
	FolderIds     []string `json:"folderIds" query:"folderIds" example:"[fId1,fId2]"`
	MediaIds      []string `json:"mediaIds" query:"mediaIds" example:"[mId1,mId2]"`
} // @name MediaBatchParams
