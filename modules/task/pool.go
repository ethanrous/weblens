package task

type Pool interface {
	QueueTask(task Task) (err error)
}
