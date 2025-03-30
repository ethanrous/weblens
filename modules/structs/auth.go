package structs

type ApiKeyInfo struct {
	Id           string `json:"id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Key          string `json:"key" validate:"required"`
	Owner        string `json:"owner" validate:"required"`
	RemoteUsing  string `json:"remoteUsing" validate:"required"`
	CreatedBy    string `json:"createdBy" validate:"required"`
	CreatedTime  int64  `json:"createdTime" validate:"required"`
	LastUsedTime int64  `json:"lastUsedTime" validate:"required"`
} // @name ApiKeyInfo
