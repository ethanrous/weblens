// Package user provides user account models and operations for the Weblens system.
package user

// PublicUserName is the username used for public access.
const PublicUserName = "PUBLIC"

// UnknownUserName is the username used for unknown users.
const UnknownUserName = "UNKNOWN"

var publicUser = User{
	Username:  PublicUserName,
	UserPerms: UserPermissionPublic,
}

// GetPublicUser returns the shared public user instance.
func GetPublicUser() *User {
	return &publicUser
}

var unknownUser = User{
	Username:  UnknownUserName,
	UserPerms: UserPermissionPublic,
}

// GetUnknownUser returns the shared unknown user instance.
func GetUnknownUser() *User {
	return &unknownUser
}
