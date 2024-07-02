package types

type weblensServicePackage struct {
	InstanceService InstanceService

	Database       DatabaseService
	FileTree       FileTree
	MediaRepo      MediaRepo
	Caster         BroadcasterAgent
	ClientManager  ClientManager
	WorkerPool     WorkerPool
	TaskDispatcher TaskPool
	Requester      Requester
	AlbumManager   AlbumService
	ShareService   ShareService
	UserService    UserService
}

var SERV *weblensServicePackage

func init() {
	SERV = &weblensServicePackage{}
}

func (srv *weblensServicePackage) SetInstance(instSrv InstanceService) {
	if SERV.InstanceService == nil {
		SERV.InstanceService = instSrv
	}
}

func (srv *weblensServicePackage) SetDatabase(db DatabaseService) {
	if SERV.Database == nil {
		SERV.Database = db
	}
}

func (srv *weblensServicePackage) SetFileTree(ft FileTree) {
	if SERV.FileTree == nil {
		SERV.FileTree = ft
	}
}

func (srv *weblensServicePackage) SetMediaRepo(mediaRepo MediaRepo) {
	if SERV.MediaRepo == nil {
		SERV.MediaRepo = mediaRepo
	}
}

func (srv *weblensServicePackage) SetClientService(clientService ClientManager) {
	if SERV.ClientManager == nil {
		SERV.ClientManager = clientService
	}
}

func (srv *weblensServicePackage) SetCaster(c BroadcasterAgent) {
	if SERV.Caster == nil {
		SERV.Caster = c
	}
}

func (srv *weblensServicePackage) SetUserService(us UserService) {
	if SERV.UserService == nil {
		SERV.UserService = us
	}
}

func (srv *weblensServicePackage) SetAlbumService(albumService AlbumService) {
	if SERV.AlbumManager == nil {
		SERV.AlbumManager = albumService
	}
}

func (srv *weblensServicePackage) SetShareService(share ShareService) {
	if SERV.ShareService == nil {
		SERV.ShareService = share
	}
}

func (srv *weblensServicePackage) SetTaskDispatcher(tasker TaskPool) {
	if SERV.TaskDispatcher == nil {
		SERV.TaskDispatcher = tasker
	}
}

func (srv *weblensServicePackage) SetRequester(rq Requester) {
	if SERV.Requester == nil {
		SERV.Requester = rq
	}
}

func (srv *weblensServicePackage) SetWorkerPool(wp WorkerPool) {
	if SERV.WorkerPool == nil {
		SERV.WorkerPool = wp
	}
}
