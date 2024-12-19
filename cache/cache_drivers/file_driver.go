package cache_drivers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HemendCo/go-core/cache/cache_models"
	"github.com/HemendCo/go-core/filemanager"
	"github.com/HemendCo/go-core/helpers"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// fileCacheItem holds the data with expiration time
type fileCacheItem struct {
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

// FileCacheDriver structure for file-based caching
type FileCacheDriver struct {
	cfg         *cache_models.FileCacheConfig
	path        string
	fileManager *filemanager.FileManager
	mu          sync.RWMutex
}

// Name implements cache.CacheDriver.
func (f *FileCacheDriver) Name() string {
	return "file"
}

// Init initializes the FileCacheDriver with the given configuration
func (f *FileCacheDriver) Init(config interface{}) error {
	if f.cfg != nil {
		return nil
	}

	cfg, ok := config.(cache_models.FileCacheConfig)
	if !ok {
		return errors.New("invalid file cache configuration: expected a cache_models.FileCacheConfig type")
	}

	f.cfg = &cfg
	f.fileManager = filemanager.NewFileManager() // Initialize FileManager
	f.path = helpers.JoinWithProjectPath(f.cfg.Path)

	return nil
}

// Set stores data in a file
func (f *FileCacheDriver) Set(key string, value interface{}, expiration time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// If serialization is enabled, the value will be serialized
	if f.cfg.Serialize {
		serializedValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		value = serializedValue
	}

	// Set expiration time
	exp := time.Now().Add(expiration)
	item := fileCacheItem{
		Value:      value,
		Expiration: exp,
	}

	// Create a file name based on the key
	filePath := f.getFilePathForKey(key)

	// Write new content to the file using FileManager
	return f.fileManager.WriteFile(filePath, item)
}

// Get retrieves data from a file
func (f *FileCacheDriver) Get(key string) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Create a file name based on the key
	filePath := f.getFilePathForKey(key)

	var item fileCacheItem

	// Read cache file content using FileManager
	err := f.fileManager.ReadFile(filePath, &item)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // key does not exist
		}
		return nil, err
	}

	// If the expiration time has passed, remove the value
	if time.Now().After(item.Expiration) {
		f.Delete(key) // Remove the expired file
		return nil, errors.New("key expired")
	}

	// If serialization is enabled, the value will be deserialized
	if f.cfg.Serialize {
		var deserializedValue interface{}
		err := json.Unmarshal(item.Value.([]byte), &deserializedValue)
		if err != nil {
			return nil, err
		}
		return deserializedValue, nil
	}

	return item.Value, nil
}

// Has checks if a key exists in the filesystem
func (f *FileCacheDriver) Has(key string) (bool, error) {
	if f.mu.TryLock() {
		f.mu.Unlock()
	}

	// Create a file name based on the key
	filePath := f.getFilePathForKey(key)

	// Check for file existence
	return f.fileManager.Has(filePath)
}

// Delete removes data from the file
func (f *FileCacheDriver) Delete(key string) error {
	if f.mu.TryLock() {
		f.mu.Unlock()
	}

	// Create a file name based on the key
	filePath := f.getFilePathForKey(key)

	// Remove the file using FileManager
	return f.fileManager.RemoveFileOrDirectory(filePath)
}

// getFilePathForKey constructs the file path based on the key
func (f *FileCacheDriver) getFilePathForKey(key string) string {
	// Use the directory path and file name (constructed from the key)
	return filepath.Join(f.path, fmt.Sprintf("%s.cache", key))
}
