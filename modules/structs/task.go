package structs

import (
	"time"

	"github.com/ethanrous/weblens/modules/task"
)

type TaskInfo struct {
	TaskId string `json:"taskId" validate:"required"`

	Completed bool                `json:"Completed" validate:"required"`
	JobName   string              `json:"jobName" validate:"required"`
	Progress  int                 `json:"progress" validate:"required"`
	Status    task.TaskExitStatus `json:"status" validate:"required"`
	WorkerId  int                 `json:"workerId" validate:"required"`
	StartTime time.Time           `json:"startTime"`
	Result    any                 `json:"result,omitempty"`
} // @name TaskInfo
