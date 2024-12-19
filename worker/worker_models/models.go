package worker_models

import "time"

type TaskMessage struct {
	ID        string
	TypeName  string
	FileName  string
	Queue     string
	Payload   []byte
	Retried   int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func NewQueue(typeName string) TaskMessage {
	return TaskMessage{
		TypeName: typeName,
		Payload:  []byte{},
		Retried:  0,
	}
}

func (q *TaskMessage) SetQueue(queue string) {
	q.Queue = queue
}

func (q *TaskMessage) SetPayload(payload []byte) {
	q.Payload = payload
}

type RedisWorkerConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Database    int
	MaxRetry    int
	Concurrency int
	Priorities  map[string]int
}

type FileWorkerConfig struct {
	Path              string
	MaxRetry          int
	CheckInterval     int
	TaskSleepDuration float64
	Concurrency       int
	Priorities        map[string]int
	Timezone          string
}
