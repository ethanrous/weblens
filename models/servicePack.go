package models

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/internal/werror"
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
	StartupChan chan bool

	startupTasks  []StartupTask
	waitingOnLock sync.RWMutex
}

type StartupTask struct {
	Name        string
	Description string
	StartedAt   time.Time
}

func (pack *ServicePack) AddStartupTask(taskName, description string) {
	t := StartupTask{Name: taskName, Description: description, StartedAt: time.Now()}

	pack.waitingOnLock.Lock()
	pack.startupTasks = append(pack.startupTasks, t)
	pack.waitingOnLock.Unlock()

	pack.Caster.PushWeblensEvent(StartupProgressEvent, WsC{"waitingOn": pack.GetStartupTasks()})
}

func (pack *ServicePack) GetStartupTasks() []StartupTask {
	pack.waitingOnLock.RLock()
	defer pack.waitingOnLock.RUnlock()
	return pack.startupTasks
}

func (pack *ServicePack) RemoveStartupTask(taskName string) error {
	pack.waitingOnLock.Lock()
	i := slices.IndexFunc(
		pack.startupTasks, func(t StartupTask) bool {
			return t.Name == taskName
		},
	)

	if i == -1 {
		pack.waitingOnLock.Unlock()
		return werror.Errorf("startup task not found")
	}

	pack.startupTasks = append(pack.startupTasks[:i], pack.startupTasks[i+1:]...)
	pack.waitingOnLock.Unlock()

	pack.Caster.PushWeblensEvent(StartupProgressEvent, WsC{"waitingOn": pack.GetStartupTasks()})

	return nil
}

type Server interface {
	Start()
	UseInit()
	UseCore()
	UseApi()
	Restart()
	Stop()
}
