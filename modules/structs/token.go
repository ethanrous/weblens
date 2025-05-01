package structs

type TokenInfo struct {
	Id          string `json:"id" validate:"required"`
	CreatedTime int64  `json:"createdTime" validate:"required" format:"int64"`
	LastUsed    int64  `json:"lastUsed" validate:"required" format:"int64"`
	Nickname    string `json:"nickname" validate:"required"`
	Owner       string `json:"owner" validate:"required"`
	RemoteUsing string `json:"remoteUsing" validate:"required"`
	CreatedBy   string `json:"createdBy" validate:"required"`
	Token       string `json:"token" validate:"required"`
} // @name TokenInfo

// type ApiKeyInfo struct {
// 	Id           string `json:"id" validate:"required"`
// 	Name         string `json:"name" validate:"required"`
// 	Key          string `json:"key" validate:"required"`
// 	Owner        string `json:"owner" validate:"required"`
// 	RemoteUsing  string `json:"remoteUsing" validate:"required"`
// 	CreatedBy    string `json:"createdBy" validate:"required"`
// 	CreatedTime  int64  `json:"createdTime" validate:"required"`
// 	LastUsedTime int64  `json:"lastUsedTime" validate:"required"`
// } // @name ApiKeyInfo
