package logger

type LoggerDriver interface {
	Name() string
	Init(config interface{}) error
	Log(message string) error
	Close() error
}
