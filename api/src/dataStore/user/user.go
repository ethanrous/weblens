package user

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        primitive.ObjectID `bson:"_id" json:"-"`
	Username  types.Username     `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Tokens    []string           `bson:"tokens" json:"-"`
	Admin     bool               `bson:"admin" json:"admin"`
	Activated bool               `bson:"activated" json:"activated"`
	Owner     bool               `bson:"owner" json:"owner"`
	HomeId    types.FileId       `bson:"homeId" json:"homeId"`
	TrashId   types.FileId       `bson:"trashId" json:"trashId"`

	// non-database types
	homeFolder   types.WeblensFile
	trashFolder  types.WeblensFile
	tokensLock   *sync.RWMutex
	isSystemUser bool
}

func New(username types.Username, password string, isAdmin, autoActivate bool) (types.User, error) {
	if types.SERV.UserService.Get(username) != nil {
		return nil, types.ErrUserAlreadyExists
	}
	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}
	passHash := string(passHashBytes)

	newUser := &User{
		Id:         primitive.NewObjectID(),
		Username:   username,
		Password:   passHash,
		Tokens:     []string{},
		Admin:      isAdmin,
		Activated:  autoActivate,
		tokensLock: &sync.RWMutex{},
	}

	homeDir, err := newUser.CreateHomeFolder()
	if err != nil {
		return nil, err
	}

	newUser.homeFolder = homeDir
	newUser.trashFolder = homeDir.GetChildren()[0]

	return newUser, nil
}

func (u *User) Activate() (err error) {
	u.Activated = true

	_, err = u.CreateHomeFolder()
	if err != nil {
		return err
	}

	err = types.SERV.StoreService.ActivateUser(u.GetUsername())
	if err != nil {
		return err
	}

	u.Activated = true

	return
}

func (u *User) GetUsername() types.Username {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *User) GetHomeFolder() types.WeblensFile {
	if u.homeFolder == nil {
		u.homeFolder = types.SERV.FileTree.Get(u.HomeId)
	}
	return u.homeFolder
}

func (u *User) SetHomeFolder(f types.WeblensFile) error {
	u.homeFolder = f
	u.HomeId = f.ID()
	return nil
}

func (u *User) CreateHomeFolder() (types.WeblensFile, error) {
	mediaRoot := types.SERV.FileTree.GetRoot()
	event := history.NewFileEvent()
	homeDir, err := types.SERV.FileTree.MkDir(mediaRoot, strings.ToLower(string(u.Username)), event)
	if err != nil && errors.Is(err, types.ErrDirAlreadyExists) {

	} else if err != nil {
		return nil, err
	}

	homeDir.SetOwner(u)

	_, err = types.SERV.FileTree.MkDir(homeDir, ".user_trash", event)
	if err != nil {
		return homeDir, err
	}

	err = types.SERV.FileTree.GetJournal().LogEvent(event)
	if err != nil {
		return homeDir, types.WeblensErrorFromError(err)
	}

	return homeDir, nil
}

func (u *User) GetTrashFolder() types.WeblensFile {
	if u.trashFolder == nil {
		u.trashFolder = types.SERV.FileTree.Get(u.TrashId)
	}
	return u.trashFolder
}

func (u *User) SetTrashFolder(f types.WeblensFile) error {
	u.trashFolder = f
	u.TrashId = f.ID()
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Admin
}

func (u *User) IsOwner() bool {
	return u.Owner
}

func (u *User) IsActive() bool {
	return u.Activated
}

func (u *User) IsSystemUser() bool {
	return u.isSystemUser
}

func (u *User) GetToken() string {
	u.tokensLock.RLock()
	if len(u.Tokens) != 0 {
		ret := u.Tokens[0]
		u.tokensLock.RUnlock()
		return ret
	}
	u.tokensLock.RUnlock()

	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte("key"))
	if err != nil {
		util.ErrTrace(err)
		return ""
	}

	ret := tokenString
	err = types.SERV.StoreService.AddTokenToUser(u.Username, tokenString)
	if err != nil {
		util.ErrTrace(err)
	}

	u.tokensLock.Lock()
	defer u.tokensLock.Unlock()
	u.Tokens = append(u.Tokens, tokenString)

	return ret
}

func (u *User) CheckLogin(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (u *User) UpdatePassword(oldPass, newPass string) (err error) {
	if auth := u.CheckLogin(oldPass); !auth {
		return types.ErrBadPassword
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPass), 14)
	if err != nil {
		return err
	}
	u.Password = string(passHashBytes)

	return types.ErrNotImplemented("UpdatePassword user")
	// return u.service.db.UpdatePsaswordByUsername(u.Username, u.Password)
}

func (u *User) SetAdmin(isAdmin bool) error {
	u.Admin = isAdmin
	return types.ErrNotImplemented("SetAdmin user")

	// return types.SERV.UserService.SetAdminByUsername(u.Username, isAdmin)
}

func MakeOwner(u types.User) error {
	realU := u.(*User)
	realU.Owner = true

	return types.NewWeblensError("not impl - make user owner")
	// return dataStore.dbServer.updateUser(realU)
}

func ShareGrantsFileAccess(share types.Share, file types.WeblensFile) bool {
	if share == nil {
		return false
	}
	if share.GetShareType() != types.FileShare {
		util.Error.Println("Trying to check if non-file share gives access to file")
		return false
	}
	shareFileId := types.FileId(share.GetItemId())

	tmpF := file
	for {
		if tmpF.ID() == shareFileId {
			return true
		}
		if tmpF.GetParent() == nil {
			break
		}
		tmpF = tmpF.GetParent()
	}
	return false
}

func (u *User) FormatArchive() (map[string]any, error) {
	data := map[string]any{
		"username":     u.Username,
		"password":     u.Password,
		"tokens":       u.Tokens,
		"admin":        u.Admin,
		"activated":    u.Activated,
		"owner":        u.Owner,
		"isSystemUser": u.isSystemUser,
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

	u.Username = types.Username(obj["username"].(string))
	u.Password = obj["password"].(string)
	u.Activated = obj["activated"].(bool)
	u.Admin = obj["admin"].(bool)
	u.Owner = obj["owner"].(bool)
	u.HomeId = types.FileId(obj["homeId"].(string))
	u.TrashId = types.FileId(obj["trashId"].(string))
	u.Tokens = util.SliceConvert[string](obj["tokens"].([]any))
	u.isSystemUser = obj["isSystemUser"].(bool)

	u.tokensLock = &sync.RWMutex{}

	return nil
}

// func (u *User) UnmarshalJSONValue(t bsontype.Type, b []byte) (err error) {
// 	util.Debug.Println(t)
// 	u = types.SERV.UserService.Get(types.Username(b)).(*User)
// 	return nil
// }
