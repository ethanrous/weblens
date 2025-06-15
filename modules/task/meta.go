package task

import "maps"

type TaskResult map[string]any

type TaskMetadata interface {
	JobName() string
	MetaString() string
	FormatToResult() TaskResult
	Verify() error
}

func (tr TaskResult) ToMap() map[string]any {
	return maps.Clone(tr)
}
