package db_drivers

import (
	"database/sql"
	"fmt"
	"github.com/HemendCo/go-core/database/db_config"
	"github.com/HemendCo/go-core/database/db_interfaces"

	"github.com/golang-migrate/migrate/v4/database"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDriver struct
type PostgresDriver struct {
	db  *gorm.DB
	cfg *db_config.DBConfig
}

func (d *PostgresDriver) Name() string {
	return "postgres"
}

// Connect method to connect to PostgreSQL database
func (d *PostgresDriver) Connect(dbConfig db_config.DBConfig) error {
	d.cfg = &dbConfig

	if d.db == nil {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}

		d.db = db
	}

	return nil
}

func (d *PostgresDriver) DB() *gorm.DB {
	return d.db
}

func (d *PostgresDriver) Config() *db_config.DBConfig {
	return d.cfg
}

func (d *PostgresDriver) SqlDB() (*sql.DB, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get *sql.DB: %v", err)
	}

	return sqlDB, nil
}

func (d *PostgresDriver) MigrateDriver() (database.Driver, error) {
	sqlDB, err := d.SqlDB()
	if err != nil {
		return nil, err
	}

	return migratePostgres.WithInstance(sqlDB, &migratePostgres.Config{})
}

func (d *PostgresDriver) Clone() db_interfaces.DatabaseDriver {
	return &PostgresDriver{
		db:  nil,
		cfg: d.cfg,
	}
}
