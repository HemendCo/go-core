package db_drivers

import (
	"database/sql"
	"fmt"
	"github.com/HemendCo/go-core/database/db_config"
	"github.com/HemendCo/go-core/database/db_interfaces"

	"github.com/golang-migrate/migrate/v4/database"
	migrateMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLDriver struct
type MySQLDriver struct {
	db  *gorm.DB
	cfg *db_config.DBConfig
}

func (d *MySQLDriver) Name() string {
	return "mysql"
}

// Connect method to connect to MySQL database
func (d *MySQLDriver) Connect(dbConfig db_config.DBConfig) error {
	d.cfg = &dbConfig

	if d.db == nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL: %w", err)
		}

		d.db = db
	}

	return nil
}

func (d *MySQLDriver) DB() *gorm.DB {
	return d.db
}

func (d *MySQLDriver) Config() *db_config.DBConfig {
	return d.cfg
}

func (d *MySQLDriver) SqlDB() (*sql.DB, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get *sql.DB: %v", err)
	}

	return sqlDB, nil
}

func (d *MySQLDriver) MigrateDriver() (database.Driver, error) {
	sqlDB, err := d.SqlDB()
	if err != nil {
		return nil, err
	}

	return migrateMysql.WithInstance(sqlDB, &migrateMysql.Config{})
}

func (d *MySQLDriver) Clone() db_interfaces.DatabaseDriver {
	return &MySQLDriver{
		db:  nil,
		cfg: d.cfg,
	}
}
