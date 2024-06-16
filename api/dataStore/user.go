package dataStore

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
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
}

type UserArray []types.User

var userMap = map[types.Username]*user{}

func CreateUser(username types.Username, password string, isAdmin, autoActivate bool, ft types.FileTree) error {
	if GetUser(username) != nil {
		return ErrUserAlreadyExists
	}
	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	passHash := string(passHashBytes)

	newUser := &user{
		Username:  username,
		Password:  passHash,
		Tokens:    []string{},
		Admin:     isAdmin,
		Activated: autoActivate,

		// HomeFolder:  homeDir,
		// TrashFolder: homeDir.GetChildren()[0],
	}

	homeDir, err := newUser.CreateHomeFolder(ft)
	if err != nil {
		return err
	}

	newUser.HomeFolder = homeDir
	newUser.TrashFolder = homeDir.GetChildren()[0]

	if len(userMap) == 0 {
		newUser.Owner = true
	}

	if err := dbServer.CreateUser(newUser); err != nil {
		return err
	}

	userMap[username] = newUser

	return nil
}

// GetUser returns pointer to user instance given their unique username
func GetUser(username types.Username) types.User {
	u, ok := userMap[username]
	if !ok {
		return nil
	}
	return u
}

func (store coreStore) LoadUsers(ft types.FileTree) (err error) {
	var users []user
	users, err = dbServer.getUsers(ft)
	if err != nil {
		util.ErrTrace(err)
		return err
	}
	for _, u := range users {
		userMap[u.Username] = &u
	}
	return
}

// func (store backupStore) LoadUsers(ft types.FileTree) (err error) {
// 	var users []types.User
// 	users, err = store.req.GetCoreUsers()
// 	if err != nil {
// 		return
// 	}
// 	for _, u := range users {
// 		realU := u.(*user)
// 		userMap[realU.Username] = realU
// 	}
// 	return
// }

func loadUsersStaticFolders(ft types.FileTree) {
	mediaRoot := ft.Get("MEDIA")
	for _, u := range userMap {
		homeId := ft.GenerateFileId(filepath.Join(mediaRoot.GetAbsPath(), string(u.Username)) + "/")
		u.HomeFolder = ft.Get(homeId)
		if u.HomeFolder == nil {
			panic(errors.Join(ErrNoFile, errors.New(string(homeId))))
		}

		trash, err := getChildByName(u.HomeFolder, ".user_trash")
		if err != nil || trash == nil {
			panic(err)
		}
		u.TrashFolder = trash
	}
}

func UserCount() int {
	return len(userMap)
}

func ActivateUser(username types.Username, ft types.FileTree) (err error) {
	u := GetUser(username)
	if u == nil {
		return ErrNoUser
	}

	u.(*user).Activated = true

	_, err = u.CreateHomeFolder(ft)
	if err != nil {
		return err
	}

	dbServer.activateUser(username)
	return
}

func getUsers() []types.User {
	return util.MapToSliceMutate(userMap, func(un types.Username, u *user) types.User { return u })
	// users := dbServer.GetUsers()
	// return util.Map(users, func(u user) types.User { return types.User(&u) })
}

func (store coreStore) GetUsers() []types.User {
	return getUsers()
}

func (store backupStore) GetUsers() []types.User {
	return getUsers()
}

func (u *user) GetUsername() types.Username {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *user) GetHomeFolder() types.WeblensFile {
	return u.HomeFolder
}

func (u *user) CreateHomeFolder(ft types.FileTree) (types.WeblensFile, error) {
	mediaRoot := ft.Get("MEDIA")
	homeDir, err := ft.MkDir(mediaRoot, strings.ToLower(string(u.Username)))
	// if err != nil && !errors.Is(err, ErrDirAlreadyExists) {
	if err != nil {
		return homeDir, err
	}

	_, err = ft.MkDir(homeDir, ".user_trash")
	if err != nil {
		return homeDir, err
	}

	return homeDir, nil
}

func (u *user) GetTrashFolder() types.WeblensFile {
	return u.TrashFolder
}

func (u *user) IsAdmin() bool {
	return u.Admin
}

func (u *user) IsOwner() bool {
	return u.Owner
}

func (u *user) IsActive() bool {
	return u.Activated
}

func (u *user) GetToken() string {
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
	dbServer.AddTokenToUser(u.Username, tokenString)
	return ret
}

func CheckLogin(u types.User, pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.(*user).Password), []byte(pass)) == nil
}

func CheckUserToken(u types.User, token string) bool {
	return u.GetToken() == token
}

func UpdatePassword(username types.Username, oldPass, newPass string) (err error) {
	u := GetUser(username)
	if u == nil {
		return ErrNoUser
	}

	if auth := CheckLogin(u, oldPass); !auth {
		return ErrBadPassword
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPass), 14)
	if err != nil {
		return err
	}
	realU := u.(*user)
	realU.Password = string(passHashBytes)
	err = dbServer.updateUser(realU)

	return
}

func UpdateAdmin(username types.Username, isAdmin bool) error {
	u := GetUser(username)
	if u == nil {
		return ErrNoUser
	}

	realU := u.(*user)
	realU.Admin = isAdmin
	err := dbServer.updateUser(realU)
	if err != nil {
		return err
	}

	return nil
}

func MakeOwner(u types.User) error {
	realU := u.(*user)
	realU.Owner = true
	return dbServer.updateUser(realU)
}

func ShareGrantsFileAccess(share types.Share, file types.WeblensFile) bool {
	if share == nil {
		return false
	}
	if share.GetShareType() != FileShare {
		util.Error.Println("Trying to check if non-file share gives access to file")
		return false
	}
	shareFileId := types.FileId(share.GetContentId())

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

func DeleteUser(user types.User) {
	delete(userMap, user.GetUsername())
	dbServer.deleteUser(user.GetUsername())
}

func (u user) MarshalJSON() ([]byte, error) {
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

func (u *user) UnmarshalJSON(data []byte) error {
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

func (ua *UserArray) UnmarshalJSON(data []byte) error {
	var realUsers []*user
	err := json.Unmarshal(data, &realUsers)
	if err != nil {
		return err
	}

	*ua = util.Map(realUsers, func(u *user) types.User { return u })

	return nil
}

func (ua UserArray) MarshalArchive() []map[string]any {
	var m []map[string]any
	for _, ui := range ua {
		u := ui.(*user)
		m = append(m, map[string]any{
			"id":        u.Id,
			"username":  u.Username,
			"password":  u.Password,
			"tokens":    u.Tokens,
			"admin":     u.Admin,
			"activated": u.Activated,
			"owner":     u.Owner,

			"homeId":  u.HomeFolder.ID(),
			"trashId": u.TrashFolder.ID(),
		})
	}

	return m
}
