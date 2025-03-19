package models

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/task"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServicePack struct {
	Log             *zerolog.Logger
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
	StartupChan chan bool

	Db *mongo.Database

	startupTasks     []StartupTask
	currentTaskStart time.Time

	Cnf env.Config

	waitingOnLock sync.RWMutex
	Loaded        atomic.Bool
	Closing       atomic.Bool

	updateMu sync.RWMutex
}

type StartupTask struct {
	StartedAt   time.Time
	Name        string
	Description string
}

func (pack *ServicePack) AddStartupTask(taskName, description string) {
	t := StartupTask{Name: taskName, Description: description, StartedAt: time.Now()}

	pack.waitingOnLock.Lock()
	pack.startupTasks = append(pack.startupTasks, t)
	pack.currentTaskStart = time.Now()
	pack.waitingOnLock.Unlock()

	pack.GetStartupTasks()
	pack.Caster.PushWeblensEvent(StartupProgressEvent, WsC{"waitingOn": pack.GetStartupTasks()})

	pack.Log.Debug().Func(func(e *zerolog.Event) { e.Str("startup_task", taskName).Msg("Beginning startup task") })
}

func (pack *ServicePack) GetStartupTasks() []StartupTask {
	newTasks := make([]StartupTask, len(pack.startupTasks))
	pack.waitingOnLock.RLock()
	defer pack.waitingOnLock.RUnlock()
	copy(newTasks, pack.startupTasks)
	return newTasks
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

	pack.Log.Debug().Func(func(e *zerolog.Event) {
		pack.waitingOnLock.RLock()
		defer pack.waitingOnLock.RUnlock()
		e.Str("startup_task", taskName).Dur("startup_task_time", time.Since(pack.currentTaskStart)).Msg("Finished startup task")
	})
}

func (pack *ServicePack) SetFileService(fs FileService) {
	pack.updateMu.Lock()
	pack.FileService = fs
	pack.updateMu.Unlock()
}

func (pack *ServicePack) GetFileService() FileService {
	pack.updateMu.RLock()
	defer pack.updateMu.RUnlock()
	return pack.FileService
}

func (pack *ServicePack) SetCaster(c Broadcaster) {
	pack.updateMu.Lock()
	pack.Caster = c
	pack.updateMu.Unlock()
}

func (pack *ServicePack) GetCaster() Broadcaster {
	pack.updateMu.RLock()
	defer pack.updateMu.RUnlock()
	return pack.Caster
}

type Server interface {
	Start()
	UseApi() *chi.Mux
	Restart(wait bool)
	Stop()
}
