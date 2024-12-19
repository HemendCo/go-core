package filemanager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	_  = iota             // Skip the first value (0)
	B  = 1 << (10 * iota) // 1 << (10 * 1) = 1024
	KB = 1 << (10 * iota) // 1 << (10 * 2) = 1048576
	MB = 1 << (10 * iota) // 1 << (10 * 3) = 1073741824
	GB = 1 << (10 * iota) // 1 << (10 * 4) = 1099511627776
	TB = 1 << (10 * iota) // 1 << (10 * 5) = 1125899906842624
)

// FileType defines the type for specifying file content formats
type FileType int

// Enum for file content types
const (
	_    FileType = iota // Skip the first value (0)
	Text                 // 1
	JSON                 // 2
	Byte                 // 3
)

// FileManager structure for managing file and directory content
type FileManager struct {
	locks sync.Map // Holds locks for different paths
}

// تعریف یک متغیر برای نگه‌داری نمونه Singleton
var (
	instance *FileManager
	once     sync.Once // برای اطمینان از این‌که تنها یک نمونه ساخته می‌شود
)

// NewFileManager creates a new instance of FileManager
func NewFileManager() *FileManager {
	once.Do(func() {
		instance = &FileManager{
			locks: sync.Map{},
		}
	})

	return instance
}

// getLock retrieves or creates a lock for a given path
func (fm *FileManager) getLock(path string) *sync.RWMutex {
	lock, _ := fm.locks.LoadOrStore(path, &sync.RWMutex{})
	return lock.(*sync.RWMutex)
}

// ListFilesInDirectory lists files and directories in a specified path
func (fm *FileManager) ListFilesInDirectory(dir string, filterFunc func(entry os.DirEntry) bool) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	var files []os.FileInfo

	for _, entry := range entries {
		// If filterFunc is provided, check if the entry matches the filter condition
		if filterFunc != nil && filterFunc(entry) {
			continue // Skip this entry if filterFunc returns true
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("error retrieving file info: %v", err)
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// ListFilesRecursively lists all files and directories in a specified path recursively
func (fm *FileManager) ListFilesRecursively(dir string) ([]os.FileInfo, error) {
	var files []os.FileInfo
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, info)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through directory: %v", err)
	}
	return files, nil
}

// WriteFile creates or updates a file with specified content in the given format
func (fm *FileManager) WriteFile(filePath string, content interface{}) error {
	lock := fm.getLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	var data []byte
	var err error

	// Determine content type and convert it to []byte
	switch out := content.(type) {
	case string:
		// Handle string content
		data = []byte(out)
	case json.RawMessage:
		// Handle Raw JSON content
		data = out
	case []byte:
		// Handle byte array content
		data = out
	default:
		// Handle any other interface type by attempting to marshal to JSON
		data, err = json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %v", err)
		}
	}

	// Write the data to the file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}
	return nil
}

func (fm *FileManager) WriteFileWithFunc(filePath string, contentFunc func() (interface{}, error)) error {
	lock := fm.getLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	var data []byte
	var err error

	// Call the provided function to get the content
	content, err := contentFunc()
	if err != nil {
		// If there was an error in contentFunc, release the lock and return the error
		return fmt.Errorf("error getting content: %w", err)
	}

	// Determine content type and convert it to []byte
	switch out := content.(type) {
	case string:
		// Handle string content
		data = []byte(out)
	case json.RawMessage:
		// Handle Raw JSON content
		data = out
	case []byte:
		// Handle byte array content
		data = out
	default:
		// Handle any other interface type by attempting to marshal to JSON
		data, err = json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %v", err)
		}
	}

	// Write the data to the file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	return nil
}

// ReadFile retrieves file content in the specified format
func (fm *FileManager) ReadFile(filePath string, output interface{}) error {
	lock := fm.getLock(filePath)
	lock.RLock()
	defer lock.RUnlock()

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.ErrNotExist
		}
		return fmt.Errorf("error reading file: %w", err)
	}

	// Determine output type and assign content
	switch out := output.(type) {
	case *string:
		*out = string(content)
	case *json.RawMessage:
		*out = json.RawMessage(content)
	case *[]byte:
		*out = content
	default:
		// Attempt to unmarshal JSON into a struct if the output is of type struct
		if err := json.Unmarshal(content, out); err != nil {
			return fmt.Errorf("error unmarshaling JSON: %w", err)
		}
	}

	return nil
}

// Has checks if a file or directory exists at the specified path
func (fm *FileManager) Has(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil // file or directory does not exist
	} else if err != nil {
		return false, err // an error occurred while checking the file
	}
	return true, nil // file or directory exists
}

// RemoveFileOrDirectory deletes a file or directory
func (fm *FileManager) RemoveFileOrDirectory(path string) error {
	lock := fm.getLock(path)

	if lock.TryLock() {
		lock.Unlock()
	}

	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("error deleting file/directory: %v", err)
	}
	return nil
}

// Rename changes the name of a file or directory
func (fm *FileManager) Rename(oldPath, newPath string) error {
	lock := fm.getLock(oldPath)
	lock.Lock()
	defer lock.Unlock()

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("error renaming: %v", err)
	}
	return nil
}

// ReadUserInput retrieves user input line by line
func (fm *FileManager) ReadUserInput() (string, error) {
	lock := fm.getLock("stdin")
	lock.Lock()
	defer lock.Unlock()

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		input += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading user input: %v", err)
	}
	return input, nil
}

func (fm *FileManager) DeleteFileIfExists(filePath string) error {
	lock := fm.getLock(filePath)

	if lock.TryLock() {
		lock.Unlock()
	}

	// Check if the file exists
	if exists, err := fm.Has(filePath); err != nil {
		return fmt.Errorf("error checking file existence: %v", err)
	} else if !exists {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Remove the file
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}
	return nil
}
