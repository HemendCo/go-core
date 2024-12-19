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
	// بارگذاری پلاگین از مسیر مشخص شده
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("could not open plugin: %v", err)
	}

	// فرض می‌کنیم که پلاگین دارای یک تابع به نام "NewPlugin" است
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("could not find NewPlugin function: %v", err)
	}

	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("invalid plugin type")
	}

	// ایجاد یک نمونه از پلاگین
	pluginInstance := newPluginFunc()

	// بررسی اینکه آیا پلاگین قبلاً بارگذاری شده است
	if pm.HasPlugin(pluginInstance.Name()) {
		return fmt.Errorf("plugin %s already exists", pluginInstance.Name())
	}

	// ذخیره پلاگین در نقشه
	pm.plugins[pluginInstance.Name()] = pluginInstance

	// فراخوانی متد Init پلاگین
	return pluginInstance.Init(config)
}

// متد برای بررسی اینکه آیا پلاگین بارگذاری شده است
func (pm *PluginManager) HasPlugin(name string) bool {
	_, exists := pm.plugins[name]
	return exists
}

// متد برای اجرای یک پلاگین خاص
func (pm *PluginManager) ExecutePlugin(name string, input interface{}) (interface{}, error) {
	if plugin, exists := pm.plugins[name]; exists {
		return plugin.Execute(input)
	}
	return nil, fmt.Errorf("plugin %s not found", name)
}
