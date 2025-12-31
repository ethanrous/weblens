package structs

import (
	"time"

	"github.com/ethanrous/weblens/modules/task"
)

// TaskInfo represents task status information for API responses.
type TaskInfo struct {
	TaskID string `json:"taskID" validate:"required"`

	Completed bool            `json:"Completed" validate:"required"`
	JobName   string          `json:"jobName" validate:"required"`
	Progress  int             `json:"progress" validate:"required"`
	Status    task.ExitStatus `json:"status" validate:"required"`
	WorkerID  int             `json:"workerID" validate:"required"`
	StartTime time.Time       `json:"startTime"`
	Result    any             `json:"result,omitempty"`
} // @name TaskInfo
