package models

import (
	"encoding/json"
	"iter"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	// non-database types
	homeFolder  *fileTree.WeblensFileImpl
	trashFolder *fileTree.WeblensFileImpl

	Username     Username        `bson:"username"` // Username is the unique identifier for the user
	PasswordHash string          `bson:"password"` // PasswordHash is the bcrypt hash of the user's password
	FullName     string          `bson:"fullName"`
	HomeId       fileTree.FileId `bson:"homeId"`
	TrashId      fileTree.FileId `bson:"trashId"`

	// The id of the server instance that created this user
	CreatedBy InstanceId `bson:"createdBy"`

	Id            primitive.ObjectID `bson:"_id"`
	Admin         bool               `bson:"admin"`
	Activated     bool               `bson:"activated"`
	IsServerOwner bool               `bson:"owner"`
	SystemUser    bool
}

func NewUser(username Username, password, fullName string, isAdmin, autoActivate bool) (*User, error) {
	if username == "" {
		return nil, werror.Errorf("username is empty")
	}
	if password == "" {
		return nil, werror.Errorf("password is empty")
	}
	if username == "PUBLIC" || username == "WEBLENS" {
		return nil, werror.Errorf("username not allowed")
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return nil, err
	}
	passHash := string(passHashBytes)

	newUser := &User{
		Id:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: passHash,
		FullName:     fullName,
		Admin:        isAdmin,
		Activated:    autoActivate,
	}

	return newUser, nil
}

func (u *User) GetUsername() Username {
	return u.Username
}

func (u *User) GetFullName() string {
	return u.FullName
}

func (u *User) SetFullName(fullName string) {
	u.FullName = fullName
}

func (u *User) SetHomeFolder(f *fileTree.WeblensFileImpl) {
	if !f.IsDir() {
		panic("home folder is not a directory")
	}
	if f.Filename() != u.Username {
		panic("home folder filename does not match user")
	}

	u.homeFolder = f
	u.HomeId = f.ID()
}

func (u *User) SetTrashFolder(f *fileTree.WeblensFileImpl) {
	if !f.IsDir() {
		panic("trash folder is not a directory")
	}
	if f.Filename() != ".user_trash" {
		panic("trash folder filename is not correct")
	}

	u.trashFolder = f
	u.TrashId = f.ID()
}

func (u *User) IsAdmin() bool {
	return u.Admin || u.IsServerOwner
}

func (u *User) IsOwner() bool {
	return u.IsServerOwner
}

func (u *User) IsPublic() bool {
	return u == nil || u.Username == "PUBLIC"
}

func (u *User) IsActive() bool {
	return u.Activated
}

func (u *User) IsSystemUser() bool {
	return u.SystemUser
}

func (u *User) CheckLogin(password string) bool {
	if !u.Activated {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (u *User) SocketType() string {
	return "webClient"
}

func MakeOwner(u *User) {
	u.IsServerOwner = true
}

func (u *User) UnmarshalJSON(data []byte) error {
	obj := map[string]any{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	u.Username = obj["username"].(string)
	u.PasswordHash = obj["password"].(string)
	u.Activated = obj["activated"].(bool)
	u.Admin = obj["admin"].(bool)
	u.IsServerOwner = obj["owner"].(bool)
	u.HomeId = obj["homeId"].(string)
	u.TrashId = obj["trashId"].(string)
	u.SystemUser = obj["isSystemUser"].(bool)

	return nil
}

type Username = string

type UserService interface {
	Size() int
	Get(id Username) *User
	Add(user *User) error
	Del(id Username) error
	GetAll() (iter.Seq[*User], error)
	CreateOwner(username, password, fullName string) (*User, error)
	GetPublicUser() *User
	SearchByUsername(searchString string) (iter.Seq[*User], error)
	SetUserAdmin(*User, bool) error
	ActivateUser(*User, bool) error
	GetRootUser() *User
	UpdateUserHome(u *User) error
	UpdateFullName(u *User, newFullName string) error

	UpdateUserPassword(username Username, oldPassword, newPassword string, allowEmptyOld bool) error
}
