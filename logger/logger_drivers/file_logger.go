package logger_drivers

import (
	"HemendCo/go-core/filemanager"
	"HemendCo/go-core/logger/logger_models"
	"errors"
	"fmt"
	"time"
)

// FileLoggerDriver for logging to a file
type FileLoggerDriver struct {
	fileManager *filemanager.FileManager
	cfg         *logger_models.FileLoggerConfig
}

// Name implements the name for the driver
func (fl *FileLoggerDriver) Name() string {
	return "file"
}

// Init creates a new FileLoggerDriver and configures it with settings
func (fl *FileLoggerDriver) Init(config interface{}) error {
	if fl.cfg != nil {
		return nil
	}

	cfg, ok := config.(logger_models.FileLoggerConfig)
	if !ok {
		return errors.New("invalid file logger configuration: expected a logger_models.FileLoggerConfig type")
	}

	fl.cfg = &cfg
	fl.fileManager = filemanager.NewFileManager()

	_, err := fl.fileManager.ListFilesInDirectory(fl.cfg.Filepath, nil)
	if err != nil {
		return fmt.Errorf("error accessing log directory at '%s': %v", fl.cfg.Filepath, err)
	}

	return nil
}

// Log records a log message with a timestamp
func (fl *FileLoggerDriver) Log(message string) error {
	timestamp := time.Now().Format(time.RFC3339)
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// Use FileManager to write the log message to the file
	if err := fl.fileManager.WriteFile(fl.cfg.Filepath, logMessage); err != nil {
		return fmt.Errorf("error writing log message to file '%s': %v", fl.cfg.Filepath, err)
	}

	return nil
}

// Close releases resources and cleans up
func (fl *FileLoggerDriver) Close() error {
	return nil
}
