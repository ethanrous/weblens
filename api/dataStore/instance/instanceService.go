package instance

import (
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type instanceService struct {
	repo  map[types.InstanceId]types.Instance
	local types.Instance
	db    types.InstanceDB
}

func NewService() types.InstanceService {
	return &instanceService{
		repo: make(map[types.InstanceId]types.Instance),
	}
}

func (is *instanceService) Init(db types.DatabaseService) error {
	is.db = db

	servers, err := db.GetAllServers()
	if err != nil {
		return err
	}

	for _, server := range servers {
		if server.IsLocal() {
			is.local = server
		}
		is.repo[server.ServerId()] = server
	}

	if is.local == nil {
		is.local = &WeblensInstance{IsThisServer: true, Role: types.Initialization}
	}

	return nil
}

func (is *instanceService) Add(i types.Instance) error {
	if i.ServerId() == "" && !i.IsLocal() {
		return types.NewWeblensError("Remote server must have specified id")
	} else if i.ServerId() == "" {
		i.(*WeblensInstance).Id = is.GenerateNewId(i.GetName())
	}
	return types.ErrNotImplemented("instance add")
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
	return types.ErrNotImplemented("instance Del")
}

func (is *instanceService) Size() int {
	return len(is.repo)
}

func (is *instanceService) GenerateNewId(name string) types.InstanceId {
	return types.InstanceId(util.GlobbyHash(12, name, time.Now().String()))
}

func (is *instanceService) GetRemotes() []types.Instance {
	return util.MapToSlicePure(is.repo)
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

func (is *instanceService) InitBackup(i types.Instance) error {
	// srvId := types.InstanceId(util.GlobbyHash(12, name, time.Now().String()))
	// err := rq.AttachToCore(srvId, coreAddress, name, key)
	// if err != nil {
	// 	return err
	// }
	//
	// wi.Id = srvId
	// wi.Name = name
	// wi.IsThisServer = true
	// wi.Role = types.Backup
	//
	// wi.CoreAddress = coreAddress
	// wi.UsingKey = key
	//
	// err = wi.db.NewServer(srvId, name, true, types.Backup)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil

	return types.ErrNotImplemented("instance InitBackup")
}
