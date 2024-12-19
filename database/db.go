package database

import (
	"HemendCo/go-core/database/db_interfaces"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	MigrateDB "github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

type DB struct {
	dbm               *DatabaseManager
	connections       map[string]db_interfaces.DatabaseConnection
	defaultConnection *db_interfaces.DatabaseConnection
	currentConnection *db_interfaces.DatabaseConnection
}

func NewDB(dbm *DatabaseManager, connections map[string]db_interfaces.DatabaseConnection) *DB {
	db := &DB{
		dbm:         dbm,
		connections: connections,
	}

	// Set the default connection if available
	if err := db.setConnection(nil); err != nil {
		log.Fatal(err)
	}

	return db
}

func (db *DB) HasConnection(connectionName string) bool {
	_, exists := db.connections[connectionName]
	return exists
}

func (db *DB) setConnection(connectionName *string) error {
	// Try to find the connection (either specific or default)
	var selectedConnection *db_interfaces.DatabaseConnection

	if connectionName == nil {
		// Look for the default connection
		for _, conn := range db.connections {
			if conn.Config().IsDefaultConnection {
				selectedConnection = &conn
				break
			}
		}
	} else {
		// Look for the specific connection
		if conn, exists := db.connections[*connectionName]; exists {
			selectedConnection = &conn
		}
	}

	// Handle missing connection or default connection
	if selectedConnection == nil {
		if connectionName != nil {
			return fmt.Errorf("connection '%s' does not exist", *connectionName)
		}
		return fmt.Errorf("default connection does not exist")
	}

	// Set the connections
	if connectionName == nil {
		db.defaultConnection = selectedConnection
	}
	db.currentConnection = selectedConnection
	return nil
}

func (db *DB) getConnection(connectionName *string) db_interfaces.DatabaseConnection {
	// Attempt to set the connection and return error if any
	if err := db.setConnection(connectionName); err != nil {
		log.Fatal(err)
	}
	return *db.currentConnection
}

func (db *DB) Connection(connectionName string) db_interfaces.DatabaseConnection {
	// Retrieve the connection for a given name
	return db.getConnection(&connectionName)
}

func (db *DB) DefaultConnection() db_interfaces.DatabaseConnection {
	return *db.defaultConnection
}

func (db *DB) DB() *gorm.DB {
	return db.DefaultConnection().DB()
}

func (db *DB) SqlDB() (*sql.DB, error) {
	return db.DefaultConnection().SqlDB()
}

func (db *DB) MigrateDriver() (MigrateDB.Driver, error) {
	return db.DefaultConnection().MigrateDriver()
}

func (db *DB) Migration(connectionNames ...string) error {
	if len(connectionNames) == 0 {
		connectionNamesList := make([]string, 0, len(db.connections))
		for name := range db.connections {
			connectionNamesList = append(connectionNamesList, name)
		}
		connectionNames = connectionNamesList
	}

	for _, connectionName := range connectionNames {
		var conn db_interfaces.DatabaseConnection
		// Look for the default connection
		if _conn, exists := db.connections[connectionName]; exists {
			conn = _conn
		}

		if conn == nil {
			fmt.Printf("Connection with name '%s' not found. Skipping migrations.\n", connectionName)
			continue
		}

		driver, err := conn.MigrateDriver()
		if err != nil {
			return fmt.Errorf("failed to create mysql driver: %v", err)
		}

		if db.hasSQLFiles(conn.Config().SchemaPath) {
			m, err := migrate.NewWithDatabaseInstance(
				fmt.Sprintf("file://%s", conn.Config().SchemaPath),
				conn.Config().Driver,
				driver,
			)

			if err != nil {
				return fmt.Errorf("failed to create migrate instance: %v", err)
			}

			// Run Migration
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to apply migrations: %v", err)
			}

			fmt.Printf("Migrations successfully applied for connection '%s'.\n", connectionName)
		} else {
			fmt.Printf("No migration files found in the specified path '%s' for connection '%s'. Skipping migrations.\n", conn.Config().SchemaPath, connectionName)
		}
	}
	return nil
}

func (db *DB) AutoMigrate(dst ...interface{}) error {
	models := make(map[string][]interface{})

	for _, item := range dst {
		var connName string
		if connector, ok := item.(db_interfaces.DBConnector); ok {
			connName = connector.ConnectionName()
		} else {
			connName = db.DefaultConnection().Name()
		}

		models[connName] = append(models[connName], item)
	}

	for connName, values := range models {
		if err := db.Connection(connName).DB().AutoMigrate(values...); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetConnectionForModel(model interface{}) db_interfaces.DatabaseConnection {
	var newDb db_interfaces.DatabaseConnection

	if model, ok := model.(db_interfaces.DBConnector); ok {
		// Retrieve the connection for the model
		newDb = db.Connection(model.ConnectionName())
	} else {
		// Use default connection
		newDb = db.DefaultConnection()
	}

	return newDb
}

func (db *DB) hasSQLFiles(dir string) bool {
	// Check if the directory exists and can be read
	files, err := os.ReadDir(dir)
	if err != nil {
		// Return false if there's an error reading the directory
		return false
	}

	// Look for files with the .sql extension
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			// Return true if a .sql file is found
			return true
		}
	}

	// Return false if no .sql files are found
	return false
}
