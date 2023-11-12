package dataProcess

type taskTracker map[string]Task

type Task struct {
	TaskType string
	Metadata string
}

var ttInstance taskTracker

func verifyTaskTracker() taskTracker {
	if ttInstance == nil {
		ttInstance = taskTracker{}
	}

	return ttInstance
}

func RequestTask(task Task) {
	verifyTaskTracker()

	_, ok := ttInstance[task.TaskType]
	if ok {
		return
	}
	ttInstance[task.TaskType] = task
	ttInstance.dispatchTask(task)
}

func (tt taskTracker) RemoveTask(taskKey string) {
	delete(tt, taskKey)
}

func (tt taskTracker) dispatchTask(task Task) {
	defer tt.RemoveTask(task.TaskType)

	switch task.TaskType {
		case "scan_directory": Scan(task.Metadata)
	}

}