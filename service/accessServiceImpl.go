package service

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.AccessService = (*AccessServiceImpl)(nil)

type AccessServiceImpl struct {
	apiKeyMap map[models.WeblensApiKey]models.ApiKeyInfo
	keyMapMu  sync.RWMutex

	tokenMap   map[string]*models.User
	tokenMapMu sync.RWMutex

	userService models.UserService
	collection  *mongo.Collection
}

func NewAccessService(userService models.UserService, col *mongo.Collection) (*AccessServiceImpl, error) {
	accSrv := &AccessServiceImpl{
		apiKeyMap: map[models.WeblensApiKey]models.ApiKeyInfo{},
		tokenMap:  map[string]*models.User{},

		collection: col,
	}

	ret, err := accSrv.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var target []models.ApiKeyInfo
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
	if getFileOwnerName(file) == user.GetUsername() {
		return true
	}

	if user.GetUsername() == "WEBLENS" {
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

func (accSrv *AccessServiceImpl) GetApiKey(key models.WeblensApiKey) (models.ApiKeyInfo, error) {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	if keyInfo, ok := accSrv.apiKeyMap[key]; !ok {
		return models.ApiKeyInfo{}, werror.Errorf("Could not find api key")
	} else {
		return keyInfo, nil
	}
}

func (accSrv *AccessServiceImpl) GenerateJwtToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte("key"))
	if err != nil {
		return "", err
	}

	dbToken := models.Token{
		Token:    tokenString,
		Username: user.GetUsername(),
	}

	_, err = accSrv.collection.InsertOne(context.Background(), dbToken)
	if err != nil {
		return "", err
	}

	user.AddToken(tokenString)

	return tokenString, nil
}

func (accSrv *AccessServiceImpl) GetUserFromToken(token string) (*models.User, error) {
	if token == "" {
		return nil, nil
	}

	accSrv.tokenMapMu.RLock()
	usr, ok := accSrv.tokenMap[token]
	accSrv.tokenMapMu.RUnlock()
	if !ok {
		var target models.Token
		err := accSrv.collection.FindOne(context.Background(), bson.M{"token": token}).Decode(&target)
		if err != nil {
			return nil, werror.WithStack(err)
		}

		usr = accSrv.userService.Get(target.Username)

		// This is ok even if the user is nil, because then the next lookup
		// won't need to go to mongo to find no user, the map will remember
		accSrv.tokenMapMu.Lock()
		accSrv.tokenMap[token] = usr
		accSrv.tokenMapMu.Unlock()
	}

	if usr == nil {
		return nil, werror.Errorf("Could not find token")
	}

	return usr, nil
	// if keyInfo, ok := accSrv.apiKeyMap[key]; !ok {
	// 	return models.ApiKeyInfo{}, errors.New("Could not find api key")
	// } else {
	// 	return keyInfo, nil
	// }
}

func (accSrv *AccessServiceImpl) DeleteApiKey(key models.WeblensApiKey) error {
	accSrv.keyMapMu.RLock()
	_, ok := accSrv.apiKeyMap[key]
	accSrv.keyMapMu.RUnlock()
	if !ok {
		return errors.New("could not find api key to delete")
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

func (accSrv *AccessServiceImpl) GenerateApiKey(creator *models.User) (models.ApiKeyInfo, error) {
	if !creator.IsAdmin() {
		return models.ApiKeyInfo{}, werror.ErrUserNotAuthorized
	}

	createTime := time.Now()
	hash := models.WeblensApiKey(internal.GlobbyHash(0, creator.GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := models.ApiKeyInfo{
		Id:          primitive.NewObjectID(),
		Key:         hash,
		Owner:       creator.GetUsername(),
		CreatedTime: createTime,
	}

	_, err := accSrv.collection.InsertOne(context.Background(), newKey)
	if err != nil {
		return models.ApiKeyInfo{}, err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()
	accSrv.apiKeyMap[newKey.Key] = newKey

	return newKey, nil
}

func (accSrv *AccessServiceImpl) SetKeyUsedBy(key models.WeblensApiKey, server *models.Instance) error {
	return werror.NotImplemented("accessService setKeyUsedBy")
	return werror.ErrKeyInUse
}

func (accSrv *AccessServiceImpl) GetAllKeys(accessor *models.User) ([]models.ApiKeyInfo, error) {
	if !accessor.IsAdmin() {
		return nil, errors.New("non-admin attempting to get api keys")
	}

	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return slices.Collect(maps.Values(accSrv.apiKeyMap)), nil
}

func getFileOwnerName(file *fileTree.WeblensFileImpl) models.Username {
	portable := file.GetPortablePath()
	if portable.RootName() != "MEDIA" {
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
