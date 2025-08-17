package reshape

import (
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/structs"
)

func TasksToTaskInfos(tasks []*task.Task) []structs.TaskInfo {
	taskInfos := make([]structs.TaskInfo, 0, len(tasks))
	for _, t := range tasks {
		taskInfos = append(taskInfos, TaskToTaskInfo(t))
	}

	return taskInfos
}

func TaskToTaskInfo(t *task.Task) structs.TaskInfo {
	complete, status := t.Status()
	result := t.GetResults()

	return structs.TaskInfo{
		TaskId:    t.Id(),
		JobName:   t.JobName(),
		Progress:  0,
		Status:    status,
		Completed: complete,
		WorkerId:  int(t.WorkerId),
		Result:    result,
		StartTime: t.StartTime,
	}
}
