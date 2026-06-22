package tower_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/routers/api/v1/tower"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterTasks(t *testing.T) {
	wp := task.NewTestWorkerPool(2)
	wp.Run(task.NewTestContext())

	wp.RegisterJob("filter-tasks-job", func(tsk *task.Task) {
		tsk.Success()
	})

	tsk, err := wp.DispatchJob(context.Background(), "filter-tasks-job", task.NewTestMetadata("filter-tasks-job"), nil)
	require.NoError(t, err)
	tsk.Wait()

	finishMs := tsk.GetFinishTime().UnixMilli()
	all := wp.GetTasks()

	contains := func(ts []*task.Task, id string) bool {
		for _, t := range ts {
			if t.ID() == id {
				return true
			}
		}

		return false
	}

	assert.False(t, contains(tower.FilterTasks(all, false, 0), tsk.ID()), "exited task must be hidden when includeExited is false")
	assert.True(t, contains(tower.FilterTasks(all, true, 0), tsk.ID()), "exited task is returned when includeExited and no cursor")
	assert.True(t, contains(tower.FilterTasks(all, true, finishMs-1000), tsk.ID()), "exited task is returned when it finished after the cursor")
	assert.True(t, contains(tower.FilterTasks(all, true, finishMs), tsk.ID()), "exited task finishing exactly at the cursor is still returned (inclusive bound)")
	assert.False(t, contains(tower.FilterTasks(all, true, finishMs+1000), tsk.ID()), "exited task is dropped when it finished before the cursor")
}
