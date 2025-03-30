package task

type TaskExitStatus string

const (
	TaskNoStatus TaskExitStatus = ""
	TaskSuccess  TaskExitStatus = "success"
	TaskCanceled TaskExitStatus = "cancelled"
	TaskError    TaskExitStatus = "error"
)
