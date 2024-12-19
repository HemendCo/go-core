package logger

import (
	"HemendCo/go-core/logger/logger_drivers"
	"fmt"
)

type LoggerManager struct {
	drivers map[string]LoggerDriver
}

func NewLoggerManager(drivers ...LoggerDriver) *LoggerManager {
	manager := &LoggerManager{
		drivers: make(map[string]LoggerDriver),
	}

	// register default driver
	drivers = append(drivers, &logger_drivers.FileLoggerDriver{})
	manager.RegisterDrivers(drivers...)

	return manager
}

func (lm *LoggerManager) RegisterDrivers(drivers ...LoggerDriver) {
	for _, driver := range drivers {
		lm.drivers[driver.Name()] = driver
	}
}

func (lm *LoggerManager) CreateLoggerFactory(driverName string, config interface{}) (LoggerDriver, error) {
	driver, exists := lm.drivers[driverName]
	if !exists {
		return nil, fmt.Errorf("unsupported logger driver %s", driverName)
	}

	if err := driver.Init(config); err != nil {
		return nil, err
	}

	return driver, nil
}
