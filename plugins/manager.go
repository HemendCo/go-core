package plugins

import (
	"fmt"
	"plugin"
)

type PluginManager struct {
	plugins map[string]Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) LoadPlugin(path string, config interface{}) error {
	// Load the plugin from the specified path.
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("could not open plugin: %v", err)
	}

	// Assume the plugin provides a function named "NewPlugin".
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("could not find NewPlugin function: %v", err)
	}

	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("invalid plugin type")
	}

	// Create an instance of the plugin.
	pluginInstance := newPluginFunc()

	// Check if the plugin is already loaded.
	if pm.HasPlugin(pluginInstance.Name()) {
		return fmt.Errorf("plugin %s already exists", pluginInstance.Name())
	}

	// Store the plugin in the map.
	pm.plugins[pluginInstance.Name()] = pluginInstance

	// Call the Init method of the plugin.
	return pluginInstance.Init(config)
}

// Method to check if a plugin is already loaded.
func (pm *PluginManager) HasPlugin(name string) bool {
	_, exists := pm.plugins[name]
	return exists
}

// Method to execute a specific plugin.
func (pm *PluginManager) ExecutePlugin(name string, input interface{}) (interface{}, error) {
	if plugin, exists := pm.plugins[name]; exists {
		return plugin.Execute(input)
	}
	return nil, fmt.Errorf("plugin %s not found", name)
}
