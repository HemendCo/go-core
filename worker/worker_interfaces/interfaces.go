package worker_interfaces

import (
	"HemendCo/go-core"
)

// Task defines the methods that all tasks should implement.
type Job interface {
	TypeName() string
	NewTask(app *core.App, params interface{}) (interface{}, error)
	Handler(app *core.App, payload []byte) error
}

type WorkerDriver interface {
	Name() string
	Init(app *core.App, config interface{}) error
	Enqueue(job Job, params interface{}) error
	RegisterJobHandlers(handlers ...Job)
	JobHandlerExists(handler Job) bool
	Close() error
	Run(handlers ...Job) error
}
