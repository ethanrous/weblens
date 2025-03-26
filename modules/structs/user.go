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
	Username        string `json:"username"`
	FullName        string `json:"fullName"`
	HomeId          string `json:"homeId"`
	TrashId         string `json:"trashId"`
	Token           string `json:"token" omitEmpty:"true"`
	HomeSize        int64  `json:"homeSize"`
	TrashSize       int64  `json:"trashSize"`
	PermissionLevel int    `json:"permissionLevel"`
} // @name UserInfo

type UserInfoArchive struct {
	UserInfo
	Password  string `json:"password" omitEmpty:"true"`
	Activated bool   `json:"activated"`
} // @name UserInfoArchive
