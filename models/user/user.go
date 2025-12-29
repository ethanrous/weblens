package user

import (
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserCollectionKey is the MongoDB collection name for users.
const UserCollectionKey = "users"

// Permissions defines the permission level for a user in the system.
type Permissions int

// User permission levels in ascending order of privilege.
const (
	UserPermissionPublic Permissions = iota
	UserPermissionBasic
	UserPermissionAdmin
	UserPermissionOwner
	UserPermissionSystem
)

// ErrUserNotFound is returned when a user cannot be found.
var ErrUserNotFound = errors.New("user not found")

// User represents a user account in the Weblens system with authentication and permission information.
type User struct {
	// Database id of the user
	ID primitive.ObjectID `bson:"_id"`

	// Username is the unique identifier for the user. can only contain alphanumeric characters, underscores, and hyphens
	Username string `bson:"username"`

	// DisplayName is the name shown in the gui for the user, typically the full name of the user
	DisplayName string `bson:"fullName"`

	// Password is the bcrypt hash of the user's password
	Password string `bson:"password"`

	// The id of the user's home folder
	HomeID string `bson:"homeID"`

	// The id of the user's trash folder
	TrashID string `bson:"trashID"`

	// The id of the server instance that created this user
	CreatedBy string `bson:"createdBy"`

	// Level of user permissions: basic, admin, or owner
	UserPerms Permissions `bson:"userPerms"`

	// Is the user activated
	Activated bool `bson:"activated"`
}

// GetUsername returns the user's unique username.
func (u *User) GetUsername() string {
	return u.Username
}

// GetDisplayName returns the user's display name (full name).
func (u *User) GetDisplayName() string {
	return u.DisplayName
}

// SetDisplayName updates the user's display name.
func (u *User) SetDisplayName(fullName string) {
	u.DisplayName = fullName
}

// IsPublic returns true if the user has public (unauthenticated) permissions.
func (u *User) IsPublic() bool {
	return u.UserPerms == UserPermissionPublic
}

// IsAdmin returns true if the user has administrator permissions or higher.
func (u *User) IsAdmin() bool {
	return u.UserPerms >= UserPermissionAdmin
}

// IsOwner returns true if the user has owner permissions or higher.
func (u *User) IsOwner() bool {
	return u.UserPerms >= UserPermissionOwner
}

// IsSystemUser returns true if the user is a system-level user with highest privileges.
func (u *User) IsSystemUser() bool {
	return u.UserPerms >= UserPermissionSystem
}

// IsActive returns true if the user account is activated.
func (u *User) IsActive() bool {
	return u.Activated
}

// CheckLogin verifies a login attempt by comparing the provided password against the stored hash.
// Returns false if the user is not activated.
func (u *User) CheckLogin(attempt string) bool {
	if !u.Activated {
		return false
	}

	return crypto.VerifyUserPassword(attempt, u.Password) == nil
}

// SocketType returns the websocket client type for this user.
func (u *User) SocketType() websocket.ClientType {
	return websocket.WebClient
}
