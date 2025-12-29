package task

import "maps"

// Result represents the results of a task execution as key-value pairs.
type Result map[string]any

// Metadata provides information about a task including its name, configuration, and validation.
type Metadata interface {
	JobName() string
	MetaString() string
	FormatToResult() Result
	Verify() error
}

// ToMap returns a cloned map representation of the TaskResult.
func (tr Result) ToMap() map[string]any {
	return maps.Clone(tr)
}
