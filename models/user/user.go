package user

import (
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const UserCollectionKey = "users"

type UserPermissions int

const (
	UserPermissionPublic UserPermissions = iota
	UserPermissionBasic
	UserPermissionAdmin
	UserPermissionOwner
	UserPermissionSystem
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	// Database id of the user
	Id primitive.ObjectID `bson:"_id"`

	// Username is the unique identifier for the user. can only contain alphanumeric characters, underscores, and hyphens
	Username string `bson:"username"`

	// DisplayName is the name shown in the gui for the user, typically the full name of the user
	DisplayName string `bson:"fullName"`

	// Password is the bcrypt hash of the user's password
	Password string `bson:"password"`

	// The id of the user's home folder
	HomeId string `bson:"homeId"`

	// The id of the user's trash folder
	TrashId string `bson:"trashId"`

	// The id of the server instance that created this user
	CreatedBy string `bson:"createdBy"`

	// Level of user permissions: basic, admin, or owner
	UserPerms UserPermissions `bson:"userPerms"`

	// Is the user activated
	Activated bool `bson:"activated"`
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetDisplayName() string {
	return u.DisplayName
}

func (u *User) SetDisplayName(fullName string) {
	u.DisplayName = fullName
}

func (u *User) IsPublic() bool {
	return u.UserPerms == UserPermissionPublic
}

func (u *User) IsAdmin() bool {
	return u.UserPerms >= UserPermissionAdmin
}

func (u *User) IsOwner() bool {
	return u.UserPerms >= UserPermissionOwner
}

func (u *User) IsSystemUser() bool {
	return u.UserPerms >= UserPermissionSystem
}

func (u *User) IsActive() bool {
	return u.Activated
}

func (u *User) CheckLogin(attempt string) bool {
	if !u.Activated {
		return false
	}

	return crypto.VerifyUserPassword(attempt, u.Password) == nil
}

func (u *User) SocketType() websocket.ClientType {
	return websocket.WebClient
}
