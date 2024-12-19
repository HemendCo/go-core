package database

import (
	"fmt"
	"github.com/HemendCo/go-core/database/db_config"
	"github.com/HemendCo/go-core/database/db_drivers"
	"github.com/HemendCo/go-core/database/db_interfaces"
)

type DatabaseManager struct {
	drivers map[string]db_interfaces.DatabaseDriver
}

func NewDatabaseManager(drivers ...db_interfaces.DatabaseDriver) *DatabaseManager {
	manager := &DatabaseManager{
		drivers: make(map[string]db_interfaces.DatabaseDriver),
	}

	// register default driver
	drivers = append(drivers, &db_drivers.MySQLDriver{}, &db_drivers.SQLiteDriver{}, &db_drivers.PostgresDriver{})
	manager.RegisterDrivers(drivers...)

	return manager
}

func (dm *DatabaseManager) RegisterDrivers(drivers ...db_interfaces.DatabaseDriver) {
	for _, driver := range drivers {
		dm.drivers[driver.Name()] = driver
	}
}

func (dm *DatabaseManager) CreateDatabaseFactory(connectionName string, config db_config.DBConfig) (db_interfaces.DatabaseConnection, error) {
	originalDriver, exists := dm.drivers[config.Driver]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver %s", config.Driver)
	}

	driver := originalDriver.Clone()

	if err := driver.Connect(config); err != nil {
		return nil, err
	}

	return CreateDBConnection(connectionName, &driver), nil
}
