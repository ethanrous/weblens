package structs

import (
	"time"
)

// TaskInfo represents task status information for API responses.
type TaskInfo struct {
	TaskID string `json:"taskID" validate:"required"`

	Completed bool      `json:"Completed" validate:"required"`
	JobName   string    `json:"jobName" validate:"required"`
	Progress  int       `json:"progress" validate:"required"`
	Status    string    `json:"status" validate:"required"`
	WorkerID  int       `json:"workerID" validate:"required"`
	StartTime time.Time `json:"startTime"`
	Result    any       `json:"result,omitempty"`
} // @name TaskInfo
