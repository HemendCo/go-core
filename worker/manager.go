package worker

import (
	"HemendCo/go-core"
	"HemendCo/go-core/worker/worker_drivers"
	"HemendCo/go-core/worker/worker_interfaces"
	"fmt"
)

type WorkerManager struct {
	drivers map[string]worker_interfaces.WorkerDriver
	app     *core.App
}

func NewWorkerManager(app *core.App, drivers ...worker_interfaces.WorkerDriver) *WorkerManager {
	manager := &WorkerManager{
		app:     app,
		drivers: make(map[string]worker_interfaces.WorkerDriver),
	}

	// register default driver
	drivers = append(drivers, &worker_drivers.RedisWorkerDriver{}, &worker_drivers.FileWorkerDriver{})
	manager.RegisterDrivers(drivers...)

	return manager
}

func (dm *WorkerManager) RegisterDrivers(drivers ...worker_interfaces.WorkerDriver) {
	for _, driver := range drivers {
		dm.drivers[driver.Name()] = driver
	}
}

func (dm *WorkerManager) CreateWorkerFactory(driverName string, config interface{}) (worker_interfaces.WorkerDriver, error) {
	driver, exists := dm.drivers[driverName]
	if !exists {
		return nil, fmt.Errorf("unsupported worker driver %s", driverName)
	}

	if err := driver.Init(dm.app, config); err != nil {
		return nil, err
	}

	return driver, nil
}
