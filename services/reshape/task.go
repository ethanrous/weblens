package reshape

import (
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/wlstructs"
)

// TasksToTaskInfos converts a slice of Task models to TaskInfo transfer objects.
func TasksToTaskInfos(tasks []*task.Task) []wlstructs.TaskInfo {
	taskInfos := make([]wlstructs.TaskInfo, 0, len(tasks))
	for _, t := range tasks {
		taskInfos = append(taskInfos, TaskToTaskInfo(t))
	}

	return taskInfos
}

// TaskToTaskInfo converts a Task model to a TaskInfo transfer object.
func TaskToTaskInfo(t *task.Task) wlstructs.TaskInfo {
	complete, status := t.Status()
	result := t.GetResults()

	return wlstructs.TaskInfo{
		TaskID:    t.ID(),
		JobName:   t.JobName(),
		Progress:  0,
		Status:    string(status),
		Completed: complete,
		WorkerID:  t.GetWorkerID(),
		Result:    result,
		StartTime: t.GetStartTime(),
	}
}
