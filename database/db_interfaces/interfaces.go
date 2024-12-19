package db_interfaces

import (
	"HemendCo/go-core/database/db_config"

	"database/sql"

	"github.com/golang-migrate/migrate/v4/database"
	"gorm.io/gorm"
)

// DatabaseDriver interface
type DatabaseDriver interface {
	Name() string
	Config() *db_config.DBConfig
	Connect(dbConfig db_config.DBConfig) error
	DB() *gorm.DB
	SqlDB() (*sql.DB, error)
	MigrateDriver() (database.Driver, error)
	Clone() DatabaseDriver
}

type DatabaseConnection interface {
	Name() string
	DriverName() string
	Config() *db_config.DBConfig
	DB() *gorm.DB
	SqlDB() (*sql.DB, error)
	MigrateDriver() (database.Driver, error)
}

type DBConnector interface {
	ConnectionName() string
}

type DBTableName interface {
	TableName() string
}
