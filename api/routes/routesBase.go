package routes

import "github.com/ethrousseau/weblens/api/types"

type routesControllers struct {
	FileTree       types.FileTree
	MediaRepo      types.MediaRepo
	Caster         types.BroadcasterAgent
	ClientManager  types.ClientManager
	TaskDispatcher types.TaskPool
	Requester      types.Requester
	AlbumManager   types.AlbumService
}

var rc routesControllers

func SetControllers(fileTree types.FileTree, mediaRepo types.MediaRepo, caster types.BroadcasterAgent,
	clientManager types.ClientManager, taskDispatcher types.TaskPool, requester types.Requester,
	albumService types.AlbumService) {

	rc = routesControllers{
		FileTree:       fileTree,
		MediaRepo:      mediaRepo,
		Caster:         caster,
		ClientManager:  clientManager,
		TaskDispatcher: taskDispatcher,
		Requester:      requester,
		AlbumManager:   albumService,
	}
}
