package instance

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type instanceService struct {
	instanceMap     map[types.InstanceId]types.Instance
	instanceMapLock *sync.RWMutex
	local           types.Instance
	localLoading    map[string]bool

	store types.InstanceStore
}

func NewService() types.InstanceService {
	return &instanceService{
		instanceMap:     make(map[types.InstanceId]types.Instance),
		instanceMapLock: &sync.RWMutex{},
		localLoading:    map[string]bool{"all": true},
	}
}

func (is *instanceService) Init(store types.InstanceStore) error {
	is.store = store

	servers, err := is.store.GetAllServers()
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	for _, server := range servers {
		if server.IsLocal() {
			is.local = server
			continue
		}
		is.instanceMap[server.ServerId()] = server
	}

	if is.local == nil {
		is.local = &WeblensInstance{IsThisServer: true, Role: types.Initialization}
	}

	return nil
}

func (is *instanceService) Add(i types.Instance) error {
	if i.ServerId() == "" && !i.IsLocal() {
		return types.WeblensErrorMsg("Remote server must have specified id")
	} else if i.ServerId() == "" {
		i.(*WeblensInstance).Id = is.GenerateNewId(i.GetName())
	}

	err := types.SERV.StoreService.NewServer(i)
	if err != nil {
		return err
	}

	err = types.SERV.StoreService.SetRemoteUsing(i.GetUsingKey(), i.ServerId())
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	is.instanceMap[i.ServerId()] = i

	if i.IsLocal() {
		is.local = i
	}

	return nil
}

func (is *instanceService) Get(iId types.InstanceId) types.Instance {
	util.ShowErr(types.ErrNotImplemented("instance Get"))
	return nil
}

func (is *instanceService) GetLocal() types.Instance {
	return is.local
	// return nil, types.ErrNotImplemented("instance GetLocal")
}

func (is *instanceService) Del(iId types.InstanceId) error {
	err := is.store.DeleteServer(iId)
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	delete(is.instanceMap, iId)

	return nil
}

func (is *instanceService) Size() int {
	return len(is.instanceMap)
}

func (is *instanceService) IsLocalLoaded() bool {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	return len(is.localLoading) == 0
}

func (is *instanceService) AddLoading(loadingKey string) {
	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	is.localLoading[loadingKey] = true
}

func (is *instanceService) RemoveLoading(loadingKey string) {
	is.instanceMapLock.Lock()
	delete(is.localLoading, loadingKey)
	is.instanceMapLock.Unlock()

	if is.IsLocalLoaded() {
		err := types.SERV.RestartRouter()
		if err != nil {
			util.ErrTrace(err)
		}
		types.SERV.Caster.PushWeblensEvent("weblens_loaded")
	}
}

func (is *instanceService) GenerateNewId(name string) types.InstanceId {
	return types.InstanceId(util.GlobbyHash(12, name, time.Now().String()))
}

func (is *instanceService) GetRemotes() []types.Instance {
	return util.MapToValues(is.instanceMap)
}

func (is *instanceService) InitCore(instance types.Instance) error {
	// u := us.Get(username)
	//
	// // Init with existing u
	// if u != nil {
	// 	if !u.CheckLogin(password) {
	// 		return types.ErrUserNotAuthenticated
	// 	} else if !u.IsAdmin() {
	// 		return types.NewWeblensError("TODO")
	// 		// err := u.SetOwner()
	// 		// if err != nil {
	// 		// 	return err
	// 		// }
	// 	}
	//
	// } else { // create new user, this will be the case 99% of the time
	// 	err := user.New(username, password, true, true, ft)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// srvId := types.InstanceId(util.GlobbyHash(12, name, time.Now().String()))
	// wi.Id = srvId
	// wi.Name = name
	// wi.IsThisServer = true
	// wi.Role = types.Core
	//
	// err := wi.db.NewServer(srvId, name, true, types.Core)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil

	return types.ErrNotImplemented("instance InitCore")
}

func (is *instanceService) InitBackup(name, coreAddr string, key types.WeblensApiKey, store types.ProxyStore) error {
	is.store = store

	srvId := types.InstanceId(util.GlobbyHash(12, name, time.Now().String()))
	i := New(srvId, name, key, types.Backup, true, coreAddr)
	remote, err := is.store.AttachToCore(i)
	if err != nil {
		return err
	}

	err = is.Add(i)
	if err != nil {
		return err
	}

	err = is.Add(remote)
	if err != nil {
		return err
	}

	return nil
}
