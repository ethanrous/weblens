package types

type weblensServicePackage struct {
	InstanceService InstanceService

	Database       DatabaseService
	FileTree       FileTree
	MediaRepo      MediaRepo
	Caster         BroadcasterAgent
	ClientManager  ClientManager
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
	SERV.InstanceService = instSrv
}

func (srv *weblensServicePackage) SetDatabase(db DatabaseService) {
	SERV.Database = db
}

func (srv *weblensServicePackage) SetFileTree(ft FileTree) {
	SERV.FileTree = ft
}

func (srv *weblensServicePackage) SetMediaRepo(mediaRepo MediaRepo) {
	SERV.MediaRepo = mediaRepo
}

func (srv *weblensServicePackage) SetClientService(clientService ClientManager) {
	SERV.ClientManager = clientService
}

func (srv *weblensServicePackage) SetCaster(c BroadcasterAgent) {
	SERV.Caster = c
}

func (srv *weblensServicePackage) SetUserService(us UserService) {
	SERV.UserService = us
}

func (srv *weblensServicePackage) SetAlbumService(albumService AlbumService) {
	SERV.AlbumManager = albumService
}

func (srv *weblensServicePackage) SetShareService(share ShareService) {
	SERV.ShareService = share
}

func (srv *weblensServicePackage) SetTaskDispatcher(tasker TaskPool) {
	SERV.TaskDispatcher = tasker
}

func (srv *weblensServicePackage) SetRequester(rq Requester) {
	SERV.Requester = rq
}
