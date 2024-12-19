package cli

import (
	"fmt"
	"github.com/HemendCo/go-core/cli/cli_drivers"
	"github.com/HemendCo/go-core/cli/cli_models"
)

type CLIManager struct {
	drivers map[string]CLIDriver
}

func NewCLIManager(drivers ...CLIDriver) *CLIManager {
	manager := &CLIManager{
		drivers: make(map[string]CLIDriver),
	}

	// register default driver
	drivers = append(drivers, &cli_drivers.CobraCommand{})
	manager.RegisterDrivers(drivers...)

	return manager
}

func (c *CLIManager) RegisterDrivers(drivers ...CLIDriver) {
	for _, driver := range drivers {
		c.drivers[driver.Name()] = driver
	}
}

func (c *CLIManager) CreateCLIFactory(driverName string, info cli_models.Info) (CLIDriver, error) {
	driver, exists := c.drivers[driverName]
	if !exists {
		return nil, fmt.Errorf("unsupported cli driver %s", driverName)
	}

	if err := driver.Init(info); err != nil {
		return nil, err
	}

	return driver, nil
}

// ListDrivers returns the names of all registered CLI drivers.
func (c *CLIManager) ListDrivers() []string {
	drivers := make([]string, 0, len(c.drivers))
	for name := range c.drivers {
		drivers = append(drivers, name)
	}
	return drivers
}
