package models

import (
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/task"
)

type Hasher struct {
	taskService task.TaskService
	pool        *task.TaskPool
}

func NewHasher(service task.TaskService) *Hasher {
	return &Hasher{taskService: service}
}

func NewHollowHasher() *Hasher {
	return &Hasher{}
}

func (h *Hasher) Hash(file *fileTree.WeblensFile, event *fileTree.FileEvent) error {
	if h.taskService == nil {
		return nil
	}

	if h.pool == nil {
		h.pool = h.taskService.NewTaskPool(false, nil)
	}

	hashMeta := HashFileMeta{File: file}
	t, err := h.taskService.DispatchJob(HashFile, hashMeta, h.pool)
	if err != nil {
		return err
	}

	t.SetPostAction(
		func(result task.TaskResult) {
			if result["contentId"] != nil {
				file.SetContentId(string(result["contentId"].(ContentId)))
				event.NewCreateAction(file)
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