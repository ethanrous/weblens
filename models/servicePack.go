package models

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/task"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
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

	Server      Server
	Loaded      atomic.Bool
	StartupChan chan bool

	Db *mongo.Database

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
	log.Trace.Printf("Added startup task: %s", taskName)
}

func (pack *ServicePack) GetStartupTasks() []StartupTask {
	pack.waitingOnLock.RLock()
	defer pack.waitingOnLock.RUnlock()
	return pack.startupTasks
}

func (pack *ServicePack) RemoveStartupTask(taskName string) {
	pack.waitingOnLock.Lock()
	i := slices.IndexFunc(
		pack.startupTasks, func(t StartupTask) bool {
			return t.Name == taskName
		},
	)

	if i == -1 {
		pack.waitingOnLock.Unlock()
		panic(werror.Errorf("startup task not found"))
	}

	pack.startupTasks = append(pack.startupTasks[:i], pack.startupTasks[i+1:]...)
	pack.waitingOnLock.Unlock()

	pack.Caster.PushWeblensEvent(StartupProgressEvent, WsC{"waitingOn": pack.GetStartupTasks()})

	log.Trace.Printf("Removed startup task: %s", taskName)
}

type Server interface {
	Start()
	UseApi() *chi.Mux
	Restart()
	Stop()
}
