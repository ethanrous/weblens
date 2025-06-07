package user

const PublicUserName = "PUBLIC"

var publicUser = User{
	Username:  PublicUserName,
	UserPerms: UserPermissionPublic,
}

func GetPublicUser() *User {
	return &publicUser
}
