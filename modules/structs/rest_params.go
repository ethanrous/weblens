package structs

import (
	"time"
)

// LoginBody represents the request body for user login.
type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name LoginBody

// UpdateFileParams represents parameters for updating a file's name or parent folder.
type UpdateFileParams struct {
	NewName     string `json:"newName"`
	NewParentID string `json:"newParentID"`
} // @name UpdateFileParams

// MoveFilesParams represents parameters for moving multiple files to a new parent folder.
type MoveFilesParams struct {
	NewParentID string   `json:"newParentID"`
	Files       []string `json:"fileIDs"`
} // @name MoveFilesParams

// FilesListParams represents a list of file IDs for batch operations.
type FilesListParams struct {
	FileIDs []string `json:"fileIDs"`
} // @name FilesListParams

// MediaIDsParams represents a list of media IDs for batch operations.
type MediaIDsParams struct {
	MediaIDs []string `json:"mediaIDs"`
} // @name MediaIDsParams

// MediaTimeBody represents parameters for adjusting media timestamps.
type MediaTimeBody struct {
	AnchorID string    `json:"anchorID"`
	NewTime  time.Time `json:"newTime"`
	MediaIDs []string  `json:"mediaIDs"`
}

// NewFileParams represents parameters for creating a new file or folder.
type NewFileParams struct {
	ParentFolderID string `json:"parentFolderID"`
	NewFileName    string `json:"newFileName"`
	FileSize       int64  `json:"fileSize"`
	IsDir          bool   `json:"isDir"`
} // @name NewFileParams

// NewFilesParams represents parameters for creating multiple files or folders.
type NewFilesParams struct {
	NewFiles []NewFileParams `json:"newFiles"`
} // @name NewFilesParams

// NewUploadParams represents parameters for initiating a file upload.
type NewUploadParams struct {
	RootFolderID string `json:"rootFolderID"`
	ChunkSize    int64  `json:"chunkSize"`
} // @name NewUploadParams

// PasswordUpdateParams represents parameters for updating a user's password.
type PasswordUpdateParams struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword" validate:"required"`
} // @name PasswordUpdateParams

// FileShareParams represents parameters for creating a file share.
type FileShareParams struct {
	FileID       string   `json:"fileID"`
	Users        []string `json:"users"`
	Public       bool     `json:"public"`
	Wormhole     bool     `json:"wormhole"`
	TimelineOnly bool     `json:"timelineOnly"`
} // @name FileShareParams

// AlbumShareParams represents parameters for sharing an album with users.
type AlbumShareParams struct {
	AlbumID string   `json:"albumID"`
	Users   []string `json:"users"`
	Public  bool     `json:"public"`
} // @name AlbumShareParams

// DeleteKeyBody represents the request body for deleting an API key.
type DeleteKeyBody struct {
	Key string `json:"key"`
}

// DeleteRemoteBody represents the request body for deleting a remote connection.
type DeleteRemoteBody struct {
	RemoteID string `json:"remoteID"`
}

// RestoreBody represents parameters for restoring files from a backup.
type RestoreBody struct {
	FileIDs   []string `json:"fileIDs"`
	Timestamp int64    `json:"timestamp"`
}

// UploadRestoreFileBody represents parameters for uploading a file during restore.
type UploadRestoreFileBody struct {
	FileID string `json:"fileID"`
}

// CreateFolderBody represents parameters for creating a new folder.
type CreateFolderBody struct {
	ParentFolderID string   `json:"parentFolderID" validate:"required"`
	NewFolderName  string   `json:"newFolderName" validate:"required"`
	Children       []string `json:"children" validate:"optional"`
} // @name CreateFolderBody

// UpdateAlbumParams represents parameters for updating an album's content and settings.
type UpdateAlbumParams struct {
	AddMedia    []string `json:"newMedia"`
	AddFolders  []string `json:"newFolders"`
	RemoveMedia []string `json:"removeMedia"`
	Cover       string   `json:"cover"`
	NewName     string   `json:"newName"`
	Users       []string `json:"users"`
	RemoveUsers []string `json:"removeUsers"`
} // @name UpdateAlbumParams

// CreateAlbumParams represents parameters for creating a new album.
type CreateAlbumParams struct {
	Name string `json:"name"`
} // @name CreateAlbumParams

// PermissionsParams represents permission settings for a user or share.
type PermissionsParams struct {
	CanView     bool `json:"canView"`
	CanEdit     bool `json:"canEdit"`
	CanDownload bool `json:"canDownload"`
	CanDelete   bool `json:"canDelete"`
} // @name PermissionsParams

// AddUserParams represents parameters for adding a user to a share with specific permissions.
type AddUserParams struct {
	PermissionsParams

	Username string `json:"username" validate:"required"`
} // @name AddUserParams

// UpdateUsersPermissionsParams represents parameters for updating user permissions on a share.
type UpdateUsersPermissionsParams struct {
	AddUsers    map[string]PermissionsParams `json:"addUsers"`
	RemoveUsers map[string]PermissionsParams `json:"removeUsers"`
} // @name UpdateUsersPermissionsParams

// SharePublicityBody represents parameters for setting a share's public visibility.
type SharePublicityBody struct {
	Public bool `json:"public"`
}

// ScanBody represents parameters for scanning a folder or file.
type ScanBody struct {
	FolderID string `json:"folderID"`
	Filename string `json:"filename"`
}

// RestoreFilesParams represents parameters for restoring files from backup.
type RestoreFilesParams struct {
	NewParentID string   `json:"newParentID"`
	FileIDs     []string `json:"fileIDs"`
	Timestamp   int64    `json:"timestamp"`
} // @name RestoreFilesBody

// RestoreCoreParams represents parameters for restoring core server configuration.
type RestoreCoreParams struct {
	HostURL  string `json:"restoreUrl"`
	ServerID string `json:"restoreID"`
} // @name RestoreCoreParams

// APIKeyParams represents parameters for creating an API key.
type APIKeyParams struct {
	Name string `json:"name" validate:"required"`
} // @name APIKeyParams

// MediaBatchParams represents parameters for retrieving a batch of media items.
type MediaBatchParams struct {
	Raw           bool     `json:"raw" query:"raw" example:"false" enums:"true,false"`
	Hidden        bool     `json:"hidden" query:"hidden" example:"false" enums:"true,false"`
	Sort          string   `json:"sort" query:"sort" example:"createDate" enums:"createDate"`
	SortDirection int      `json:"sortDirection" example:"1" enums:"1,-1"`
	Search        string   `json:"search" query:"search" example:""`
	Page          int      `json:"page" query:"page" example:"1"`
	Limit         int      `json:"limit" query:"limit" example:"20"`
	FolderIDs     []string `json:"folderIDs" query:"folderIDs" example:"[fID1,fID2]"`
	MediaIDs      []string `json:"mediaIDs" query:"mediaIDs" example:"[mID1,mID2]"`
} // @name MediaBatchParams
