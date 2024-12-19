package plugins

type ActionCallback func(data interface{})

type Plugin interface {
	Name() string
	Init(config interface{}) error
	Execute(input interface{}) (interface{}, error)
}
