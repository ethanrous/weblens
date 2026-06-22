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

	var metadata any
	if m := t.FormatMetadata(); m != nil {
		metadata = m
	}

	errMsg := ""
	if err := t.ReadError(); err != nil {
		errMsg = err.Error()
	}

	parentTaskID := ""

	tp := t.GetTaskPool()
	if tp != nil && tp.CreatedInTask() != nil {
		parentTaskID = tp.CreatedInTask().ID()
	}

	ctp := t.GetChildTaskPool()
	totalChildTasks := 0
	completedChildTasks := 0

	if ctp != nil {
		totalChildTasks = ctp.GetTotalTaskCount()
		completedChildTasks = ctp.GetCompletedTaskCount()
	}

	return wlstructs.TaskInfo{
		TaskID:              t.ID(),
		ParentTaskID:        parentTaskID,
		JobName:             t.JobName(),
		Progress:            0,
		Status:              string(status),
		State:               t.QueueState().String(),
		Completed:           complete,
		WorkerID:            t.GetWorkerID(),
		Result:              result,
		Metadata:            metadata,
		Error:               errMsg,
		StartTime:           t.GetStartTime(),
		QueueTime:           t.GetQueueTime(),
		FinishTime:          t.GetFinishTime(),
		TotalChildTasks:     totalChildTasks,
		CompletedChildTasks: completedChildTasks,
	}
}
