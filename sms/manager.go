package sms

import (
	"fmt"

	"github.com/HemendCo/go-core"
	"github.com/HemendCo/go-core/sms/sms_drivers"
)

type SMSManager struct {
	app     *core.App
	drivers map[string]SMSDriver
}

func NewSMSManager(app *core.App) *SMSManager {
	manager := &SMSManager{
		app:     app,
		drivers: make(map[string]SMSDriver),
	}

	// register default driver
	manager.RegisterDriver(&sms_drivers.HemendSMSDriver{})

	return manager
}

func (sm *SMSManager) RegisterDriver(driver SMSDriver) {
	sm.drivers[driver.Name()] = driver
}

func (sm *SMSManager) CreateSMSFactory(driverName string, config interface{}) (SMSDriver, error) {
	driver, exists := sm.drivers[driverName]
	if !exists {
		return nil, fmt.Errorf("unsupported sms driver %s", driverName)
	}

	if err := driver.Init(sm.app, config); err != nil {
		return nil, err
	}

	return driver, nil
}
