package user

const PublicUserName = "PUBLIC"
const UnknownUserName = "UNKNOWN"

var publicUser = User{
	Username:  PublicUserName,
	UserPerms: UserPermissionPublic,
}

func GetPublicUser() *User {
	return &publicUser
}

var unknownUser = User{
	Username:  UnknownUserName,
	UserPerms: UserPermissionPublic,
}

func GetUnknownUser() *User {
	return &unknownUser
}
