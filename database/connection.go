package database

import (
	"HemendCo/go-core/database/db_config"
	"HemendCo/go-core/database/db_interfaces"
	"database/sql"

	MigrateDB "github.com/golang-migrate/migrate/v4/database"
	"gorm.io/gorm"
)

// DBConnection struct
type DBConnection struct {
	connectionName string
	driver         *db_interfaces.DatabaseDriver
}

func CreateDBConnection(connectionName string, driver *db_interfaces.DatabaseDriver) db_interfaces.DatabaseConnection {
	return &DBConnection{
		connectionName: connectionName,
		driver:         driver,
	}
}

func (dc *DBConnection) Name() string {
	return (*dc).connectionName
}

func (dc *DBConnection) DriverName() string {
	return (*dc.driver).Name()
}

func (dc *DBConnection) DB() *gorm.DB {
	return (*dc.driver).DB()
}

func (dc *DBConnection) Config() *db_config.DBConfig {
	return (*dc.driver).Config()
}

func (dc *DBConnection) SqlDB() (*sql.DB, error) {
	return (*dc.driver).SqlDB()
}

func (dc *DBConnection) MigrateDriver() (MigrateDB.Driver, error) {
	return (*dc.driver).MigrateDriver()
}
