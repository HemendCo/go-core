package worker_drivers

import (
	"HemendCo/go-core"
	"HemendCo/go-core/worker/worker_interfaces"
	"HemendCo/go-core/worker/worker_models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type RedisWorkerDriver struct {
	app    *core.App
	cfg    worker_models.RedisWorkerConfig
	client *asynq.Client
	server *asynq.Server
	jobs   []worker_interfaces.Job
}

func (r *RedisWorkerDriver) Name() string {
	return "redis"
}

func (r *RedisWorkerDriver) Init(app *core.App, config interface{}) error {
	cfg, ok := config.(worker_models.RedisWorkerConfig)
	if !ok {
		return fmt.Errorf("unsupported worker config: worker_models.RedisWorkerConfig")
	}

	r.app = app
	r.cfg = cfg

	return nil
}

func (r *RedisWorkerDriver) Enqueue(job worker_interfaces.Job, params interface{}) error {
	preJson, err := job.NewTask(r.app, params)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(preJson)
	if err != nil {
		return err
	}

	// Create a new Asynq task
	asynqTask := asynq.NewTask(job.TypeName(), payload)

	queue := "default"

	// Enqueue the task to the Redis queue
	_, err = r.getClient().Enqueue(asynqTask, asynq.Queue(queue), asynq.MaxRetry(r.cfg.MaxRetry))
	return err
}

// RegisterJobHandlers adds multiple job handlers to the worker driver.
func (r *RedisWorkerDriver) RegisterJobHandlers(handlers ...worker_interfaces.Job) {
	if len(handlers) == 0 {
		fmt.Println("No job handlers provided to register.")
		return
	}

	for _, handler := range handlers {
		if !r.JobHandlerExists(handler) {
			// If the job handler does not exist, add it to the jobs list.
			r.jobs = append(r.jobs, handler)
		} else {
			// If it already exists, log a message or take other actions if needed.
			fmt.Printf("Job handler %s is already registered and will not be added again.\n", handler.TypeName())
		}
	}
}

// JobHandlerExists checks if a specific job handler is already registered.
func (r *RedisWorkerDriver) JobHandlerExists(handler worker_interfaces.Job) bool {
	for _, h := range r.jobs {
		if h.TypeName() == handler.TypeName() { // Assuming TypeName() is a method of JobTask
			return true
		}
	}
	return false
}

func (r *RedisWorkerDriver) Run(handlers ...worker_interfaces.Job) error {
	r.RegisterJobHandlers(handlers...)

	if err := r.getServer().Run(asynq.HandlerFunc(r.jobHandler())); err != nil {
		return err
	}

	return nil
}

func (r *RedisWorkerDriver) Close() error {
	return r.client.Close()
}

func (r *RedisWorkerDriver) getClient() *asynq.Client {
	if r.client != nil {
		return r.client
	}

	r.client = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", r.cfg.Host, r.cfg.Port),
		Username: r.cfg.Username,
		Password: r.cfg.Password,
		DB:       r.cfg.Database,
	})

	return r.client
}

func (r *RedisWorkerDriver) getServer() *asynq.Server {
	if r.server != nil {
		return r.server
	}

	r.server = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%s", r.cfg.Host, r.cfg.Port),
			Username: r.cfg.Username,
			Password: r.cfg.Password,
			DB:       r.cfg.Database,
		},
		asynq.Config{
			// Specify the number of concurrent workers
			Concurrency: r.cfg.Concurrency,
			// Specify multiple queues with different priorities
			Queues: r.cfg.Priorities,
		},
	)

	return r.server
}

func (r *RedisWorkerDriver) jobHandler() func(context.Context, *asynq.Task) error {
	// jobHandler returns a function that processes tasks from the queue.
	// It matches the task type with registered job handlers and processes the task accordingly.
	return func(ctx context.Context, t *asynq.Task) error {
		// Iterate through the list of jobs to find a matching handler
		for _, job := range r.jobs {
			// Check if the job type matches the task type
			if job.TypeName() == t.Type() {
				// Process the task with the corresponding handler
				err := job.Handler(r.app, t.Payload())
				if err != nil {
					// Log the error and return it
					fmt.Println("Error processing task:", err)
					return err
				}

				// Successfully processed the task, return nil
				return nil
			}
		}

		// Return an error if no matching job handler was found
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}
}
