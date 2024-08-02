package dataStore

import (
	"strconv"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccessService struct {
	keyMap   map[types.WeblensApiKey]types.ApiKeyInfo
	keyMapMu *sync.RWMutex

	db types.AccessStore
}

func NewAccessService() types.AccessService {
	return &AccessService{
		keyMap:   map[types.WeblensApiKey]types.ApiKeyInfo{},
		keyMapMu: &sync.RWMutex{},
	}
}

func (accSrv *AccessService) Init(db types.AccessStore) error {
	accSrv.db = db

	keys, err := accSrv.db.GetApiKeys()
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

func (accSrv *AccessService) Add(keyInfo types.ApiKeyInfo) error {
	return types.ErrNotImplemented("accessService add")
}

func (accSrv *AccessService) Get(key types.WeblensApiKey) types.ApiKeyInfo {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return accSrv.keyMap[key]
}

func (accSrv *AccessService) Del(key types.WeblensApiKey) error {
	accSrv.keyMapMu.RLock()
	_, ok := accSrv.keyMap[key]
	accSrv.keyMapMu.RUnlock()
	if !ok {
		return types.WeblensErrorMsg("could not find api key to delete")
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

func (accSrv *AccessService) Size() int {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return len(accSrv.keyMap)
}

func (accSrv *AccessService) GetApiKeyInfo(key types.WeblensApiKey) types.ApiKeyInfo {
	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()

	return accSrv.keyMap[key]
}

func (accSrv *AccessService) GenerateApiKey(acc types.AccessMeta) (types.ApiKeyInfo, error) {
	// return types.ApiKeyInfo{}, types.ErrNotImplemented("CreateApiKey")

	if !acc.User().IsAdmin() {
		return types.ApiKeyInfo{}, types.ErrUserNotAuthorized
	} else if acc.RequestMode() != ApiKeyCreate {
		return types.ApiKeyInfo{}, ErrBadRequestMode
	}

	createTime := time.Now()
	hash := types.WeblensApiKey(util.GlobbyHash(0, acc.User().GetUsername(), strconv.Itoa(int(createTime.Unix()))))

	newKey := types.ApiKeyInfo{
		Id:          primitive.NewObjectID(),
		Key:         hash,
		Owner:       acc.User().GetUsername(),
		CreatedTime: createTime,
	}

	err := types.SERV.StoreService.CreateApiKey(newKey)
	if err != nil {
		return types.ApiKeyInfo{}, types.WeblensErrorFromError(err)
	}
	accSrv.keyMapMu.Lock()
	defer accSrv.keyMapMu.Unlock()
	accSrv.keyMap[newKey.Key] = newKey

	return newKey, nil
}

func (accSrv *AccessService) GetAllKeys(acc types.AccessMeta) ([]types.ApiKeyInfo, error) {
	if !acc.User().IsAdmin() {
		return nil, types.WeblensErrorMsg("non-admin attempting to get api keys")
	}

	accSrv.keyMapMu.RLock()
	defer accSrv.keyMapMu.RUnlock()
	return util.MapToValues(accSrv.keyMap), nil

	// if acc.RequestMode() != ApiKeyGet {
	// 	return nil, ErrBadRequestMode
	// }
	// keys := dbServer.getApiKeysByUser(acc.User().GetUsername())
	// if keys == nil {
	// 	keys = []ApiKeyInfo{}
	// }
	// return keys, nil
}

func DeleteApiKey(key types.WeblensApiKey) error {
	return types.ErrNotImplemented("Delete api key")
	// keyMapMu.Lock()
	// delete(apiKeyMap, key)
	// keyMapMu.Unlock()
	// dbServer.removeApiKey(key)
}
