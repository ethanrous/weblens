package structs

// FileInfo is a structure for safely sending file information to the client.
type FileInfo struct {
	ID              string   `json:"id"`
	PortablePath    string   `json:"portablePath"`
	ParentID        string   `json:"parentID"`
	ContentID       string   `json:"contentID"`
	Owner           string   `json:"owner"`
	ShareID         string   `json:"shareID,omitempty"`
	Children        []string `json:"childrenIds"`
	Size            int64    `json:"size" swaggertype:"integer" format:"int64"`
	ModTime         int64    `json:"modifyTimestamp" swaggertype:"integer" format:"int64"`
	IsDir           bool     `json:"isDir"`
	Modifiable      bool     `json:"modifiable"`
	PastFile        bool     `json:"pastFile"`
	HasRestoreMedia bool     `json:"hasRestoreMedia"`
} // @name FileInfo

// FolderInfoResponse represents the complete information about a folder including its children, parents, media, and self reference.
type FolderInfoResponse struct {
	Children []FileInfo  `json:"children"`
	Parents  []FileInfo  `json:"parents"`
	Medias   []MediaInfo `json:"medias"`
	Self     FileInfo    `json:"self"`
} // @name FolderInfo

// NewUploadInfo represents the response containing a new upload identifier.
type NewUploadInfo struct {
	UploadID string `json:"uploadID"`
} // @name NewUploadInfo

// NewFileInfo represents the response containing a newly created file identifier.
type NewFileInfo struct {
	FileID string `json:"fileID"`
} // @name NewFileInfo

// NewFilesInfo represents the response containing multiple newly created file identifiers.
type NewFilesInfo struct {
	FileIDs []string `json:"fileIDs"`
} // @name NewFilesInfo

// RestoreFilesInfo represents the response containing the new parent identifier after restoring files.
type RestoreFilesInfo struct {
	NewParentID string `json:"newParentID"`
} //	@name	RestoreFilesInfo
