package wlstructs

// ShareInfo represents file share information for API responses.
type ShareInfo struct {
	ShareID      string                     `json:"shareID"`
	FileID       string                     `json:"fileID"`
	ShareName    string                     `json:"shareName"`
	Owner        string                     `json:"owner"`
	ShareType    string                     `json:"shareType"`
	Accessors    []UserInfo                 `json:"accessors"`
	Permissions  map[string]PermissionsInfo `json:"permissions"`
	Expires      int64                      `json:"expires" swaggertype:"integer" format:"int64"`
	Updated      int64                      `json:"updated" swaggertype:"integer" format:"int64"`
	Public       bool                       `json:"public"`
	Wormhole     bool                       `json:"wormhole"`
	TimelineOnly bool                       `json:"timelineOnly"`
	Enabled      bool                       `json:"enabled"`
} // @name ShareInfo

// PermissionsInfo represents permission settings for API responses.
type PermissionsInfo struct {
	CanView     bool `json:"canView"`
	CanEdit     bool `json:"canEdit"`
	CanDownload bool `json:"canDownload"`
	CanDelete   bool `json:"canDelete"`
} // @name PermissionsInfo
