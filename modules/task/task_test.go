//go:build test

package task_test

import (
	"testing"

	"github.com/ethanrous/weblens/modules/task"
	"github.com/stretchr/testify/assert"
)

func TestResultToMap(t *testing.T) {
	t.Run("clones result to map", func(t *testing.T) {
		result := task.Result{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		cloned := result.ToMap()

		assert.Equal(t, result["key1"], cloned["key1"])
		assert.Equal(t, result["key2"], cloned["key2"])
		assert.Equal(t, result["key3"], cloned["key3"])
	})

	t.Run("returns independent copy", func(t *testing.T) {
		result := task.Result{
			"key": "original",
		}
		cloned := result.ToMap()

		// Modify the clone
		cloned["key"] = "modified"
		cloned["new"] = "value"

		// Original should be unchanged
		assert.Equal(t, "original", result["key"])
		assert.Nil(t, result["new"])
	})

	t.Run("handles empty result", func(t *testing.T) {
		result := task.Result{}
		cloned := result.ToMap()
		assert.Equal(t, 0, len(cloned))
	})

	t.Run("handles nil values", func(t *testing.T) {
		result := task.Result{
			"nilKey": nil,
		}
		cloned := result.ToMap()
		assert.Nil(t, cloned["nilKey"])
	})

	t.Run("handles complex values", func(t *testing.T) {
		result := task.Result{
			"slice": []int{1, 2, 3},
			"map":   map[string]string{"nested": "value"},
		}
		cloned := result.ToMap()

		assert.Equal(t, result["slice"], cloned["slice"])
		assert.Equal(t, result["map"], cloned["map"])
	})
}

func TestExitStatusConstants(t *testing.T) {
	t.Run("TaskNoStatus is empty string", func(t *testing.T) {
		assert.Equal(t, task.ExitStatus(""), task.TaskNoStatus)
	})

	t.Run("TaskSuccess value", func(t *testing.T) {
		assert.Equal(t, task.ExitStatus("success"), task.TaskSuccess)
	})

	t.Run("TaskCanceled value", func(t *testing.T) {
		assert.Equal(t, task.ExitStatus("cancelled"), task.TaskCanceled)
	})

	t.Run("TaskError value", func(t *testing.T) {
		assert.Equal(t, task.ExitStatus("error"), task.TaskError)
	})
}

func TestPoolStatus(t *testing.T) {
	t.Run("zero value pool status", func(t *testing.T) {
		status := task.PoolStatus{}
		assert.Equal(t, int64(0), status.Complete)
		assert.Equal(t, 0, status.Failed)
		assert.Equal(t, int64(0), status.Total)
		assert.Equal(t, float64(0), status.Progress)
	})

	t.Run("pool status with values", func(t *testing.T) {
		status := task.PoolStatus{
			Complete: 5,
			Failed:   1,
			Total:    10,
			Progress: 0.5,
		}
		assert.Equal(t, int64(5), status.Complete)
		assert.Equal(t, 1, status.Failed)
		assert.Equal(t, int64(10), status.Total)
		assert.Equal(t, 0.5, status.Progress)
	})
}

func TestErrChildTaskFailed(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		assert.Equal(t, "child task failed", task.ErrChildTaskFailed.Error())
	})
}
