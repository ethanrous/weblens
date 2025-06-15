package structs

type ShareInfo struct {
	ShareId     string                     `json:"shareId"`
	FileId      string                     `json:"fileId"`
	ShareName   string                     `json:"shareName"`
	Owner       string                     `json:"owner"`
	ShareType   string                     `json:"shareType"`
	Accessors   []UserInfo                 `json:"accessors"`
	Permissions map[string]PermissionsInfo `json:"permissions"`
	Expires     int64                      `json:"expires"`
	Updated     int64                      `json:"updated"`
	Public      bool                       `json:"public"`
	Wormhole    bool                       `json:"wormhole"`
	Enabled     bool                       `json:"enabled"`
} // @name ShareInfo

type PermissionsInfo struct {
	CanEdit     bool `json:"canEdit"`
	CanDownload bool `json:"canDownload"`
	CanDelete   bool `json:"canDelete"`
} // @name PermissionsInfo
