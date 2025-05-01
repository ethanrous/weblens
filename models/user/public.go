package user

var publicUser = User{
	Username:  "PUBLIC",
	UserPerms: UserPermissionPublic,
}

func GetPublicUser() *User {
	return &publicUser
}
