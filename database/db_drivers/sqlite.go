package db_drivers

import (
	"HemendCo/go-core/database/db_config"
	"HemendCo/go-core/database/db_interfaces"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	migrateSqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDriver struct
type SQLiteDriver struct {
	db  *gorm.DB
	cfg *db_config.DBConfig
}

func (d *SQLiteDriver) Name() string {
	return "sqlite"
}

// Connect method to connect to SQLite database
func (d *SQLiteDriver) Connect(dbConfig db_config.DBConfig) error {
	d.cfg = &dbConfig

	if d.db == nil {
		db, err := gorm.Open(sqlite.Open(dbConfig.Database), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite: %w", err)
		}

		d.db = db
	}

	return nil
}

func (d *SQLiteDriver) DB() *gorm.DB {
	return d.db
}

func (d *SQLiteDriver) Config() *db_config.DBConfig {
	return d.cfg
}

func (d *SQLiteDriver) SqlDB() (*sql.DB, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get *sql.DB: %v", err)
	}

	return sqlDB, nil
}

func (d *SQLiteDriver) MigrateDriver() (database.Driver, error) {
	sqlDB, err := d.SqlDB()
	if err != nil {
		return nil, err
	}

	return migrateSqlite.WithInstance(sqlDB, &migrateSqlite.Config{})
}

func (d *SQLiteDriver) Clone() db_interfaces.DatabaseDriver {
	return &SQLiteDriver{
		db:  nil,
		cfg: d.cfg,
	}
}
