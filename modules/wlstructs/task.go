package wlstructs

import (
	"time"
)

// TaskInfo represents task status information for API responses.
type TaskInfo struct {
	TaskID       string `json:"taskID" validate:"required"`
	ParentTaskID string `json:"parentTaskID,omitempty"`

	Completed bool      `json:"Completed" validate:"required"`
	JobName   string    `json:"jobName" validate:"required"`
	Progress  int       `json:"progress" validate:"required"`
	Status    string    `json:"status" validate:"required"`
	State     string    `json:"State" validate:"required"`
	WorkerID  int       `json:"workerID" validate:"required"`
	StartTime  time.Time `json:"startTime"`
	QueueTime  time.Time `json:"queueTime"`
	FinishTime time.Time `json:"finishTime"`
	Result    any       `json:"result,omitempty"`
	Metadata  any       `json:"metadata,omitempty"`
	Error     string    `json:"error,omitempty"`

	TotalChildTasks     int `json:"totalChildTasks,omitempty"`
	CompletedChildTasks int `json:"completedChildTasks,omitempty"`
} //	@name	TaskInfo
