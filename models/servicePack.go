package models

import (
	"sync/atomic"

	"github.com/ethrousseau/weblens/task"
)

type ServicePack struct {
	FileService     FileService
	MediaService    MediaService
	AccessService   AccessService
	UserService     UserService
	ShareService    ShareService
	InstanceService InstanceService
	AlbumService    AlbumService
	TaskService     task.TaskService
	ClientService   ClientManager
	Caster          Broadcaster

	Server Server
	Loaded atomic.Bool
}

type Server interface {
	Start()
	UseInit()
	UseCore()
	UseApi()
}
