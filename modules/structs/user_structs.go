package structs

type NewUserParams struct {
	FullName     string `json:"fullName" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
	Admin        bool   `json:"admin"`
	AutoActivate bool   `json:"autoActivate"`
} // @name NewUserParams

type LoginParams struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
} // @name LoginBody

type UserInfo struct {
	Activated       bool   `json:"activated" validate:"required"`
	FullName        string `json:"fullName" validate:"required"`
	HomeId          string `json:"homeId" validate:"required"`
	PermissionLevel int    `json:"permissionLevel" validate:"required"`
	Token           string `json:"token" omitEmpty:"true"`
	TrashId         string `json:"trashId" validate:"required"`
	Username        string `json:"username" validate:"required"`
	IsOnline        bool   `json:"isOnline"`
} // @name UserInfo

type UserInfoArchive struct {
	UserInfo

	Password string `json:"password" omitEmpty:"true"`
} // @name UserInfoArchive
