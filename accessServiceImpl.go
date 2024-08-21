package weblens

import (
	"maps"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccessServiceImpl struct {
	keyMap map[WeblensApiKey]ApiKeyInfo
	keyMapMu *sync.RWMutex

	collection *mongo.Collection
}

func NewAccessService(col *mongo.Collection) *AccessServiceImpl {
	return &AccessServiceImpl{
		keyMap:     map[WeblensApiKey]ApiKeyInfo{},
		keyMapMu: &sync.RWMutex{},
		collection: col,
	}
}

func (accSrv *AccessServiceImpl) Init() error {
	keys, err := accSrv.collection.GetApiKeys()
	if err != nil {
		return err
	}

	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()

	for _, key := range keys {
		accSrv.keyMap[key.Key] = key
	}

	return nil
}

func (accSrv *AccessServiceImpl) Add(keyInfo ApiKeyInfo) error {
	return werror.NotImplemented("accessService add")
}

func (accSrv *AccessServiceImpl) Get(key WeblensApiKey) ApiKeyInfo {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return accSrv.keyMap[key]
}

func (accSrv *AccessServiceImpl) Del(key WeblensApiKey) error {
	accSrv.keyMapMu.RLock()
	_, ok := accSrv.keyMap[key]
	accSrv.keyMapMu.RUnlock()
	if !ok {
		return werror.WErrMsg("could not find api key to delete")
	}
	err := accSrv.db.DeleteApiKey(key)
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

func (accSrv *AccessServiceImpl) GetApiKeyInfo(key WeblensApiKey) ApiKeyInfo {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return accSrv.keyMap[key]
}

func (accSrv *AccessServiceImpl) GenerateApiKey(acc types.AccessMeta) (ApiKeyInfo, error) {
	// return ApiKeyInfo{}, types.ErrNotImplemented("CreateApiKey")

	if !acc.User().IsAdmin() {
		return ApiKeyInfo{}, ErrUserNotAuthorized
	} else if acc.RequestMode() != dataStore.ApiKeyCreate {
		return ApiKeyInfo{}, dataStore.ErrBadRequestMode
	}

	createTime := time.Now()
	hash := WeblensApiKey(internal.GlobbyHash(0, acc.User().GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := ApiKeyInfo{
		Id:          primitive.NewObjectID(),
		Key:         hash,
		Owner:       acc.User().GetUsername(),
		CreatedTime: createTime,
	}

	err := types.SERV.StoreService.CreateApiKey(newKey)
	if err != nil {
		return ApiKeyInfo{}, werror.Wrap(err)
	}
	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()
	accSrv.keyMap[newKey.Key] = newKey

	return newKey, nil
}

func (accSrv *AccessServiceImpl) GetAllKeys(acc types.AccessMeta) ([]ApiKeyInfo, error) {
	if !acc.User().IsAdmin() {
		return nil, werror.WErrMsg("non-admin attempting to get api keys")
	}

	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return slices.Collect(maps.Values(accSrv.keyMap)), nil

	// if acc.RequestMode() != ApiKeyGet {
	// 	return nil, ErrBadRequestMode
	// }
	// keys := dbServer.getApiKeysByUser(acc.User().GetUsername())
	// if keys == nil {
	// 	keys = []ApiKeyInfo{}
	// }
	// return keys, nil
}

func DeleteApiKey(key WeblensApiKey) error {
	return werror.NotImplemented("Delete api key")
	// keyMapMu.Lock()
	// delete(apiKeyMap, key)
	// keyMapMu.Unlock()
	// dbServer.removeApiKey(key)
}
