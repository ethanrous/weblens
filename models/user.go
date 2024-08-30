package models

import (
	"encoding/json"
	"iter"
	"sync"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id            primitive.ObjectID `bson:"_id" json:"-"`
	Username      Username           `bson:"username" json:"username"`
	Password      string             `bson:"password" json:"-"`
	Tokens        []string           `bson:"tokens" json:"-"`
	Admin         bool               `bson:"admin" json:"admin"`
	Activated     bool               `bson:"activated" json:"activated"`
	IsServerOwner bool               `bson:"owner" json:"owner"`
	HomeId        fileTree.FileId    `bson:"homeId" json:"homeId"`
	TrashId       fileTree.FileId    `bson:"trashId" json:"trashId"`

	// non-database types
	homeFolder  *fileTree.WeblensFileImpl
	trashFolder *fileTree.WeblensFileImpl
	tokensLock sync.RWMutex
	SystemUser bool
}

func NewUser(username Username, password string, isAdmin, autoActivate bool) (*User, error) {
	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return nil, err
	}
	passHash := string(passHashBytes)

	newUser := &User{
		Id:        primitive.NewObjectID(),
		Username:  username,
		Password:  passHash,
		Tokens:    []string{},
		Admin:     isAdmin,
		Activated: autoActivate,
	}

	return newUser, nil
}

func (u *User) GetUsername() Username {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *User) SetHomeFolder(f *fileTree.WeblensFileImpl) {
	u.homeFolder = f
	u.HomeId = f.ID()
}

func (u *User) SetTrashFolder(f *fileTree.WeblensFileImpl) {
	u.trashFolder = f
	u.TrashId = f.ID()
}

func (u *User) IsAdmin() bool {
	return u.Admin || u.IsServerOwner
}

func (u *User) IsOwner() bool {
	return u.IsServerOwner
}

func (u *User) IsActive() bool {
	return u.Activated
}

func (u *User) IsSystemUser() bool {
	return u.SystemUser
}

func (u *User) GetToken() string {
	u.tokensLock.RLock()
	defer u.tokensLock.RUnlock()

	if len(u.Tokens) != 0 {
		ret := u.Tokens[0]
		return ret
	}

	return ""
}

func (u *User) CheckLogin(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (u *User) AddToken(tokenString string) {
	u.tokensLock.Lock()
	defer u.tokensLock.Unlock()
	u.Tokens = append(u.Tokens, tokenString)
}

func (u *User) SocketType() string {
	return "webClient"
}

func MakeOwner(u *User) {
	u.IsServerOwner = true
}

func (u *User) FormatArchive() (map[string]any, error) {
	data := map[string]any{
		"username":     u.Username,
		"password":     u.Password,
		"tokens":       u.Tokens,
		"admin":        u.Admin,
		"activated":    u.Activated,
		"owner":        u.IsServerOwner,
		"isSystemUser": u.SystemUser,
		"homeId":       "",
		"trashId":      "",
	}

	if u.homeFolder != nil && u.trashFolder != nil {
		data["homeId"] = u.homeFolder.ID()
		data["trashId"] = u.trashFolder.ID()
	}

	return data, nil
}

func (u *User) UnmarshalJSON(data []byte) error {
	obj := map[string]any{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	u.Username = Username(obj["username"].(string))
	u.Password = obj["password"].(string)
	u.Activated = obj["activated"].(bool)
	u.Admin = obj["admin"].(bool)
	u.IsServerOwner = obj["owner"].(bool)
	u.HomeId = fileTree.FileId(obj["homeId"].(string))
	u.TrashId = fileTree.FileId(obj["trashId"].(string))
	u.Tokens = internal.SliceConvert[string](obj["tokens"].([]any))
	u.SystemUser = obj["isSystemUser"].(bool)

	return nil
}

type Username string

type UserService interface {
	Init() error
	Size() int
	Get(id Username) *User
	Add(user *User) error
	Del(id Username) error
	GetAll() (iter.Seq[*User], error)
	GetPublicUser() *User
	SearchByUsername(searchString string) (iter.Seq[*User], error)
	SetUserAdmin(*User, bool) error
	ActivateUser(*User) error
	GetRootUser() *User

	GenerateToken(user *User) (string, error)
	UpdateUserPassword(username Username, oldPassword, newPassword string, allowEmptyOld bool) error
}
