package structs

// FileInfo is a structure for safely sending file information to the client.
type FileInfo struct {
	Id              string   `json:"id"`
	PortablePath    string   `json:"portablePath"`
	Filename        string   `json:"filename"`
	ParentId        string   `json:"parentId"`
	ContentId       string   `json:"contentId"`
	Owner           string   `json:"owner"`
	ShareId         string   `json:"shareId,omitempty"`
	PastId          string   `json:"currentId"`
	Children        []string `json:"childrenIds"`
	Size            int64    `json:"size"`
	ModTime         int64    `json:"modifyTimestamp"`
	IsDir           bool     `json:"isDir"`
	Modifiable      bool     `json:"modifiable"`
	PastFile        bool     `json:"pastFile"`
	HasRestoreMedia bool     `json:"hasRestoreMedia"`
} // @name FileInfo

type FolderInfoResponse struct {
	Children []FileInfo  `json:"children"`
	Parents  []FileInfo  `json:"parents"`
	Medias   []MediaInfo `json:"medias"`
	Self     FileInfo    `json:"self"`
} // @name FolderInfo

type NewUploadInfo struct {
	UploadId string `json:"uploadId"`
} // @name NewUploadInfo

type NewFileInfo struct {
	FileId string `json:"fileId"`
} // @name NewFileInfo

type NewFilesInfo struct {
	FileIds []string `json:"fileIds"`
} // @name NewFilesInfo

type RestoreFilesInfo struct {
	NewParentId string `json:"newParentId"`
} //	@name	RestoreFilesInfo
