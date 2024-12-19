package interfaces

import (
	"context"
	"github.com/HemendCo/go-core"
)

// AppInterface defines the methods for the App structure.
type AppInterface interface {
	// GetContext returns the application context.
	GetContext() *context.Context

	// Config returns the configuration of the application.
	Config() interface{}

	// SetConfig sets the configuration of the application.
	SetConfig(config interface{}) *core.App

	// RootPath returns the root path of the application.
	RootPath() string

	// SetRootPath sets the root path of the application.
	SetRootPath(path string) *core.App

	// RegisterPlugin registers a plugin at the specified path with the given configuration.
	RegisterPlugin(path string, config interface{}) error

	// Use registers a key with a create function for dependency management.
	Use(key string, createFunc func() (interface{}, error)) error

	// Get retrieves the value associated with the key.
	Get(key string) (interface{}, error)

	// Exists checks if a key is registered.
	Exists(key string) bool
}
