package service

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.AccessService = (*AccessServiceImpl)(nil)

type AccessServiceImpl struct {
	userService models.UserService
	apiKeyMap   map[models.WeblensApiKey]models.ApiKey
	collection  *mongo.Collection
	keyMapMu    sync.RWMutex
}

type WlClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAccessService(userService models.UserService, col *mongo.Collection) (*AccessServiceImpl, error) {
	accSrv := &AccessServiceImpl{
		apiKeyMap: map[models.WeblensApiKey]models.ApiKey{},

		collection: col,
	}

	ret, err := accSrv.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var target []models.ApiKey
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()

	for _, key := range target {
		accSrv.apiKeyMap[key.Key] = key
	}

	accSrv.userService = userService

	return accSrv, nil
}

func (accSrv *AccessServiceImpl) CanUserAccessFile(
	user *models.User, file *fileTree.WeblensFileImpl, share *models.FileShare,
) bool {
	if user == nil || user.IsPublic() {
		return share != nil && share.IsPublic()
	}

	if getFileOwnerName(file) == user.GetUsername() {
		return true
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return true
	}

	if share == nil || !share.Enabled || (!share.Public && !slices.Contains(share.Accessors, user.GetUsername())) {
		return false
	}

	tmpFile := file
	for tmpFile.GetParent() != nil {
		if tmpFile.ID() == share.FileId {
			return true
		}
		tmpFile = tmpFile.GetParent()
	}
	return false
}

func (accSrv *AccessServiceImpl) CanUserModifyShare(user *models.User, share models.Share) bool {
	return user.GetUsername() == share.GetOwner()
}

func (accSrv *AccessServiceImpl) CanUserAccessAlbum(
	user *models.User, album *models.Album,
	share *models.AlbumShare,
) bool {
	if album.Owner == user.GetUsername() {
		return true
	}

	if share == nil || !share.Enabled || (!share.Public && !slices.Contains(share.Accessors, user.GetUsername())) {
		return false
	}

	return false
}

func (accSrv *AccessServiceImpl) GetApiKey(key models.WeblensApiKey) (models.ApiKey, error) {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	if keyInfo, ok := accSrv.apiKeyMap[key]; !ok {
		return models.ApiKey{}, werror.ErrKeyNotFound
	} else {
		return keyInfo, nil
	}
}

func (accSrv *AccessServiceImpl) GenerateJwtToken(user *models.User) (string, time.Time, error) {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	claims := WlClaims{
		user.GetUsername(),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("key"))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expires, nil
}

func (accSrv *AccessServiceImpl) GetUserFromToken(tokenStr string) (*models.User, error) {
	if tokenStr == "" {
		return nil, werror.ErrNoAuth
	}

	jwtToken, err := jwt.ParseWithClaims(
		tokenStr,
		&WlClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("key"), nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, werror.WithStack(werror.ErrTokenExpired)
		}
		return nil, werror.WithStack(err)
	}

	usr := accSrv.userService.Get(jwtToken.Claims.(*WlClaims).Username)
	if usr == nil {
		return nil, werror.ErrNoUser
	}

	return usr, nil
}

func (accSrv *AccessServiceImpl) DeleteApiKey(key models.WeblensApiKey) error {
	accSrv.keyMapMu.RLock()
	_, ok := accSrv.apiKeyMap[key]
	accSrv.keyMapMu.RUnlock()
	if !ok {
		return werror.Errorf("could not find api key to delete")
	}

	_, err := accSrv.collection.DeleteOne(context.Background(), bson.M{"key": key})
	if err != nil {
		return err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()
	delete(accSrv.apiKeyMap, key)

	return nil
}

func (accSrv *AccessServiceImpl) Size() int {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return len(accSrv.apiKeyMap)
}

func (accSrv *AccessServiceImpl) GenerateApiKey(creator *models.User, local *models.Instance, keyName string) (
	models.ApiKey, error,
) {
	createTime := time.Now()
	hash := models.WeblensApiKey(internal.GlobbyHash(0, creator.GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := models.ApiKey{
		Name:        keyName,
		Id:          primitive.NewObjectID(),
		Key:         hash,
		Owner:       creator.GetUsername(),
		CreatedTime: createTime,
		CreatedBy:   local.ServerId(),
		LastUsed:    time.UnixMilli(0),
	}

	_, err := accSrv.collection.InsertOne(context.Background(), newKey)
	if err != nil {
		return models.ApiKey{}, err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()
	accSrv.apiKeyMap[newKey.Key] = newKey

	return newKey, nil
}

func (accSrv *AccessServiceImpl) SetKeyUsedBy(key models.WeblensApiKey, remote *models.Instance) error {
	accSrv.keyMapMu.RLock()
	keyInfo, ok := accSrv.apiKeyMap[key]
	accSrv.keyMapMu.RUnlock()

	if !ok {
		return werror.WithStack(werror.ErrKeyNotFound)
	}

	if keyInfo.RemoteUsing != "" && remote != nil {
		return werror.WithStack(werror.ErrKeyInUse)
	}

	newUsingId := ""
	if remote != nil {
		newUsingId = remote.ServerId()
	}

	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"remoteUsing": newUsingId}}
	_, err := accSrv.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return werror.WithStack(err)
	}

	keyInfo.RemoteUsing = newUsingId
	accSrv.keyMapMu.Lock()
	accSrv.apiKeyMap[key] = keyInfo
	accSrv.keyMapMu.Unlock()

	return nil
}

func (accSrv *AccessServiceImpl) GetKeysByUser(accessor *models.User) ([]models.ApiKey, error) {
	accSrv.keyMapMu.RLock()
	keys := slices.Collect(maps.Values(accSrv.apiKeyMap))
	accSrv.keyMapMu.RUnlock()

	var usersKeys []models.ApiKey
	for _, key := range keys {
		if key.Owner == accessor.Username {
			usersKeys = append(usersKeys, key)
		}
	}

	return usersKeys, nil
}

func (accSrv *AccessServiceImpl) GetAllKeysByServer(
	accessor *models.User, serverId models.InstanceId,
) ([]models.ApiKey, error) {
	if !accessor.IsAdmin() {
		return nil, werror.Errorf("non-admin attempting to get api keys")
	}

	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	// Filter by serverId
	ret := []models.ApiKey{}
	for _, key := range accSrv.apiKeyMap {
		if key.CreatedBy == serverId {
			ret = append(ret, key)
		}
	}

	return ret, nil
}

func (accSrv *AccessServiceImpl) AddApiKey(key models.ApiKey) error {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	if _, ok := accSrv.apiKeyMap[key.Key]; ok {
		return werror.WithStack(werror.ErrKeyAlreadyExists)
	}

	if key.CreatedBy == "" {
		return werror.WithStack(werror.ErrKeyNoServer)
	}

	_, err := accSrv.collection.InsertOne(context.Background(), key)
	if err != nil {
		return werror.WithStack(err)
	}

	accSrv.apiKeyMap[key.Key] = key
	return nil
}

func getFileOwnerName(file *fileTree.WeblensFileImpl) models.Username {
	portable := file.GetPortablePath()
	if portable.RootName() != "USERS" {
		return "WEBLENS"
	}
	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username models.Username
	if slashIndex == -1 {
		username = models.Username(portable.RelativePath())
	} else {
		username = models.Username(portable.RelativePath()[:slashIndex])
	}

	return username
}
