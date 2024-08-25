package service

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.AccessService = (*AccessServiceImpl)(nil)

type AccessServiceImpl struct {
	keyMap   map[models.WeblensApiKey]models.ApiKeyInfo
	keyMapMu *sync.RWMutex

	collection *mongo.Collection
}

func (accSrv *AccessServiceImpl) CanUserAccessFile(
	user *models.User, file *fileTree.WeblensFile, share *models.FileShare,
) bool {
	// wlog.Error.Println("IMPLEMENT CAN USER ACCESS FILE")
	return true
}

func (accSrv *AccessServiceImpl) CanUserAccessShare(user *models.User, share models.Share) bool {
	// TODO implement me
	panic("implement me")
}

func (accSrv *AccessServiceImpl) CanUserAccessAlbum(user *models.User, album *models.Album) bool {
	// TODO implement me
	panic("implement me")
}

func (accSrv *AccessServiceImpl) GetApiKeyById(key models.WeblensApiKey) (models.ApiKeyInfo, error) {
	// TODO implement me
	panic("implement me")
}

func NewAccessService(col *mongo.Collection) *AccessServiceImpl {
	return &AccessServiceImpl{
		keyMap:     map[models.WeblensApiKey]models.ApiKeyInfo{},
		keyMapMu:   &sync.RWMutex{},
		collection: col,
	}
}

func (accSrv *AccessServiceImpl) Init() error {
	ret, err := accSrv.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	var target []models.ApiKeyInfo
	err = ret.All(context.Background(), &target)
	if err != nil {
		return err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()

	for _, key := range target {
		accSrv.keyMap[key.Key] = key
	}

	return nil
}

func (accSrv *AccessServiceImpl) Add(keyInfo models.ApiKeyInfo) error {
	return werror.NotImplemented("accessService add")
}

func (accSrv *AccessServiceImpl) Get(key models.WeblensApiKey) (models.ApiKeyInfo, error) {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	if keyInfo, ok := accSrv.keyMap[key]; !ok {
		return models.ApiKeyInfo{}, errors.New("Could not find api key")
	} else {
		return keyInfo, nil
	}
}

func (accSrv *AccessServiceImpl) Del(key models.WeblensApiKey) error {
	accSrv.keyMapMu.RLock()
	_, ok := accSrv.keyMap[key]
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
	delete(accSrv.keyMap, key)

	return nil
}

func (accSrv *AccessServiceImpl) Size() int {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return len(accSrv.keyMap)
}

func (accSrv *AccessServiceImpl) GetApiKeyInfo(key models.WeblensApiKey) models.ApiKeyInfo {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return accSrv.keyMap[key]
}

func (accSrv *AccessServiceImpl) GenerateApiKey(creator *models.User) (models.ApiKeyInfo, error) {
	// return ApiKeyInfo{}, types.ErrNotImplemented("CreateApiKey")

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
	accSrv.keyMap[newKey.Key] = newKey

	return newKey, nil
}

func (accSrv *AccessServiceImpl) SetKeyUsedBy(key models.WeblensApiKey, server *models.Instance) error {
	return werror.NotImplemented("accessService setKeyUsedBy")
}

func (accSrv *AccessServiceImpl) GetAllKeys(accessor *models.User) ([]models.ApiKeyInfo, error) {
	if !accessor.IsAdmin() {
		return nil, werror.New("non-admin attempting to get api keys")
	}

	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return slices.Collect(maps.Values(accSrv.keyMap)), nil
}

// setRemoteUsing
// filter := bson.M{"key": key}
// update := bson.M{"$set": bson.M{"remoteUsing": remoteId}}
// _, err := db.apiKeys.UpdateOne(db.ctx, filter, update)
// if err != nil {
// return error2.Wrap(err)
// }
