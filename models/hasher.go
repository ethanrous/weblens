package models

import (
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/task"
)

type Hasher struct {
	taskService task.TaskService
	caster      Broadcaster
	pool        *task.TaskPool
}

func NewHasher(service task.TaskService, caster Broadcaster) *Hasher {
	return &Hasher{taskService: service, caster: caster}
}

func (h *Hasher) Hash(file *fileTree.WeblensFileImpl) error {
	if h.taskService == nil {
		return nil
	}

	if h.pool == nil {
		h.pool = h.taskService.NewTaskPool(false, nil)
	}

	hashMeta := HashFileMeta{File: file, Caster: h.caster}
	t, err := h.taskService.DispatchJob(HashFileTask, hashMeta, h.pool)
	if err != nil {
		return err
	}

	t.SetPostAction(
		func(result task.TaskResult) {
			if result["contentId"] != nil {
				file.SetContentId(result["contentId"].(ContentId))
			} else {
				log.Error.Println("Failed to generate contentId for", file.Filename())
			}
		},
	)

	return nil
}

func (h *Hasher) Wait() {
	if h.pool == nil {
		return
	}

	h.pool.SignalAllQueued()
	h.pool.Wait(false)
}
