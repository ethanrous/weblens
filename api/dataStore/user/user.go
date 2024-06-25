package user

import (
	"encoding/json"
	"errors"
	"strings"

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
	// HomeId types.FileId `bson:"homeId" json:"-"`
	// TrashId types.FileId `bson:"trashId" json:"-"`

	// non-database types
	HomeFolder  types.WeblensFile `bson:"-" json:"homeId"`
	TrashFolder types.WeblensFile `bson:"-" json:"trashId"`

	service *userService
}

func New(username types.Username, password string, isAdmin, autoActivate bool) (types.User, error) {
	if types.SERV.UserService.Get(username) != nil {
		return nil, types.ErrUserAlreadyExists
	}
	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	passHash := string(passHashBytes)

	newUser := &User{
		Username:  username,
		Password:  passHash,
		Tokens:    []string{},
		Admin:     isAdmin,
		Activated: autoActivate,
	}

	homeDir, err := newUser.CreateHomeFolder()
	if err != nil {
		return nil, err
	}

	newUser.HomeFolder = homeDir
	newUser.TrashFolder = homeDir.GetChildren()[0]

	return newUser, nil
}

func (u *User) Activate() (err error) {
	u.Activated = true

	_, err = u.CreateHomeFolder()
	if err != nil {
		return err
	}

	err = types.SERV.Database.ActivateUser(u.GetUsername())
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
	return u.HomeFolder
}

func (u *User) SetHomeFolder(f types.WeblensFile) error {
	u.HomeFolder = f
	return nil
}

func (u *User) CreateHomeFolder() (types.WeblensFile, error) {
	mediaRoot := types.SERV.FileTree.Get("MEDIA")
	homeDir, err := types.SERV.FileTree.MkDir(mediaRoot, strings.ToLower(string(u.Username)))
	if err != nil && errors.Is(err, types.ErrDirAlreadyExists) {
		return homeDir, nil
	}

	if err != nil {
		return nil, err
	}

	_, err = types.SERV.FileTree.MkDir(homeDir, ".user_trash")
	if err != nil {
		return homeDir, err
	}

	return homeDir, nil
}

func (u *User) GetTrashFolder() types.WeblensFile {
	return u.TrashFolder
}

func (u *User) SetTrashFolder(f types.WeblensFile) error {
	u.TrashFolder = f
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

func (u *User) GetToken() string {
	if len(u.Tokens) != 0 {
		ret := u.Tokens[0]
		return ret
	}

	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte("key"))
	if err != nil {
		util.ErrTrace(err)
		return ""
	}

	ret := tokenString
	u.Tokens = append(u.Tokens, tokenString)
	util.ErrTrace(types.NewWeblensError("not impl - add token to user"))
	// dataStore.dbServer.AddTokenToUser(u.Username, tokenString)
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

	return u.service.db.UpdatePsaswordByUsername(u.Username, u.Password)
}

func (u *User) SetAdmin(isAdmin bool) error {
	u.Admin = isAdmin
	return u.service.db.SetAdminByUsername(u.Username, isAdmin)
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

func (u User) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"username":  u.Username,
		"admin":     u.Admin,
		"activated": u.Activated,
		"owner":     u.Owner,
		"homeId":    u.HomeFolder.ID(),
		"trashId":   u.TrashFolder.ID(),
	}

	return json.Marshal(m)
}

func (u *User) UnmarshalJSON(data []byte) error {
	obj := map[string]any{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	u.Activated = obj["activated"].(bool)
	u.Admin = obj["admin"].(bool)
	u.Owner = obj["owner"].(bool)
	// u.HomeFolder = FsTreeGet(types.FileId(obj["homeId"].(string)))
	// u.TrashFolder = FsTreeGet(types.FileId(obj["trashId"].(string)))
	u.Username = types.Username(obj["username"].(string))

	return nil
}
