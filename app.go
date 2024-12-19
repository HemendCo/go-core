package core

import (
	"HemendCo/go-core/helpers"
	"HemendCo/go-core/plugins"
	"context"
	"fmt"
	"sync"
)

type Keywords string

const (
	CacheKeyword         Keywords = "cache"
	LoggerKeyword        Keywords = "logger"
	CLIKeyword           Keywords = "cli"
	DatabaseKeyword      Keywords = "database"
	WorkerKeyword        Keywords = "worker"
	SMSKeyword           Keywords = "sms"
	PluginManagerKeyword Keywords = "pluginManager"
)

func (k Keywords) IsValid() bool {
	switch k {
	case CacheKeyword, LoggerKeyword, CLIKeyword, DatabaseKeyword, WorkerKeyword, SMSKeyword, PluginManagerKeyword:
		return true
	}
	return false
}

type Option interface {
}

// Option specifies the task processing behavior.
type option struct {
	rootPath string
	config   interface{}
}

// Internal option representations.
type (
	rootPathOption string
	configOption   interface{}
)

// Queue returns an option to specify the queue to enqueue the task into.
func RootPath(name string) Option {
	return rootPathOption(name)
}

func Config(config interface{}) Option {
	return configOption(config)
}

type App struct {
	context   *context.Context
	opt       option
	singleton Singleton
}

// Global variable to hold the singleton instance
var appInstance *App
var onceApp sync.Once

func defaultOptions() []Option {
	var defaultOpts []Option

	rootPath, err := helpers.ProjectPath()
	if err == nil {
		defaultOpts = append(defaultOpts, RootPath(rootPath))
	}

	return defaultOpts
}

// Function to get the app instance (similar to app() in Laravel)
func CreateApp(opts ...Option) *App {
	onceApp.Do(func() {
		opts = append(defaultOptions(), opts...)
		opt, err := composeOptions(opts...)
		if err != nil {
			// options is not valid
			return
		}

		appInstance = &App{
			opt:       opt,
			singleton: *NewSingleton(),
		}
	})
	return appInstance
}

func composeOptions(opts ...Option) (option, error) {
	res := option{
		rootPath: "/",
		config:   nil,
	}
	for _, opt := range opts {
		switch opt := opt.(type) {
		case rootPathOption:
			rootPath := string(opt)
			res.rootPath = rootPath
		case configOption:
			config := interface{}(opt)
			res.config = config
		default:
			// ignore unexpected option
		}
	}
	return res, nil
}

func (a *App) GetContext() *context.Context {
	return a.context
}

func (a *App) Config() interface{} {
	return a.opt.config
}

func (a *App) SetConfig(config interface{}) *App {
	a.opt.config = config
	return a
}

func (a *App) RootPath() string {
	return a.opt.rootPath
}

func (a *App) SetRootPath(path string) *App {
	a.opt.rootPath = path
	return a
}

func (a *App) RegisterPlugin(path string, config interface{}) error {
	keyword := PluginManagerKeyword

	if !a.Exists(keyword) {
		err := a.Use(keyword, func() (interface{}, error) {
			pm := plugins.NewPluginManager()
			return pm, nil
		})

		if err != nil {
			return err
		}
	}

	pluginManager, err := a.Get(keyword)
	if err != nil {
		return err
	}

	if pm, ok := pluginManager.(plugins.PluginManager); ok {
		return pm.LoadPlugin(path, config)
	}

	return fmt.Errorf("unsupported plugin: %s", path)
}

func (a *App) Use(key Keywords, createFunc func() (interface{}, error)) error {
	return a.singleton.Register(string(key), createFunc)
}

func (a *App) Get(key Keywords) (interface{}, error) {
	return a.singleton.Get(string(key))
}

func (a *App) Exists(key Keywords) bool {
	return a.singleton.Exists(string(key))
}
