package structs

// NewUserParams contains parameters for creating a new user.
type NewUserParams struct {
	FullName     string `json:"fullName" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
	Admin        bool   `json:"admin"`
	AutoActivate bool   `json:"autoActivate"`
} // @name NewUserParams

// LoginParams contains parameters for user login.
type LoginParams struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
} // @name LoginBody

// UserInfo represents user information for API responses.
type UserInfo struct {
	Activated       bool   `json:"activated" validate:"required"`
	FullName        string `json:"fullName" validate:"required"`
	HomeID          string `json:"homeID" validate:"required"`
	PermissionLevel int    `json:"permissionLevel" validate:"required"`
	Token           string `json:"token" omitEmpty:"true"`
	TrashID         string `json:"trashID" validate:"required"`
	Username        string `json:"username" validate:"required"`
	IsOnline        bool   `json:"isOnline"`
	UpdatedAt       int64  `json:"updatedAt" validate:"required" swaggertype:"integer" format:"int64"`
} // @name UserInfo

// UserInfoArchive extends UserInfo with password for backup/restore operations.
type UserInfoArchive struct {
	UserInfo

	Password string `json:"password" omitEmpty:"true"`
} // @name UserInfoArchive
