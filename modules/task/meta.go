package task

type TaskResult map[string]any

type TaskMetadata interface {
	JobName() string
	MetaString() string
	FormatToResult() TaskResult
	Verify() error
}

func (tr TaskResult) ToMap() map[string]any {
	m := map[string]any{}
	for k, v := range tr {
		m[string(k)] = v
	}
	return m
}
