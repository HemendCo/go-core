package core

import (
	"fmt"
	"sync"
)

// instanceData struct to hold instance-related information
type instanceData struct {
	instance interface{} // Store the creator function as an interface{}
	executed sync.Once   // Flag to ensure the function is executed only onceSingleton
}

// Singleton struct for managing Singletons
type Singleton struct {
	instances sync.Map // Using sync.Map for concurrent access
}

// NewSingleton creates a new Singleton factory
func NewSingleton() *Singleton {
	return &Singleton{}
}

// Register a creator function for a given key
func (s *Singleton) Register(key string, createFunc func() (interface{}, error)) error {
	if _, ok := s.instances.Load(key); ok {
		return fmt.Errorf("[Singleton] instance for key %s already exists", key)
	}

	// Store the creator function as instanceData with executed flag set to sync.Once
	s.instances.Store(key, &instanceData{instance: createFunc})
	return nil
}

// Get returns an instance for the given key, creating it if it does not exist
func (s *Singleton) Get(key string) (interface{}, error) {
	if data, ok := s.instances.Load(key); ok {
		info := data.(*instanceData) // Type assert to *instanceData

		var newInstance interface{}
		var err error

		// Execute the creator function if it has not been executed yet
		info.executed.Do(func() {
			createFunc := info.instance.(func() (interface{}, error))
			newInstance, err = createFunc() // Execute createFunc to create a new instance
			if err == nil && newInstance != nil {
				info.instance = newInstance // Update instance with the newly created instance
			}
		})

		if err != nil {
			return nil, fmt.Errorf("[Singleton] failed to create instance for key %s: %v", key, err)
		}

		// Return the instance (either created or previously stored)
		return info.instance, nil
	}

	return nil, fmt.Errorf("[Singleton] no instance found for key %s", key)
}

// Exists checks if an instance exists for the given key
func (s *Singleton) Exists(key string) bool {
	_, ok := s.instances.Load(key)
	return ok
}
