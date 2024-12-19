package cache_drivers

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/HemendCo/go-core/cache/cache_models"
)

// mapCacheItem holds cached data along with its expiration time.
type mapCacheItem struct {
	value      interface{}
	expiration time.Time
}

// MapCacheDriver is a structure for in-memory caching.
type MapCacheDriver struct {
	cache map[string]mapCacheItem
	cfg   *cache_models.MapCacheConfig
	mu    sync.RWMutex
}

// Name returns the name of the cache driver.
func (r *MapCacheDriver) Name() string {
	return "map"
}

// Init initializes the cache with the provided configuration.
func (r *MapCacheDriver) Init(config interface{}) error {
	if r.cache != nil {
		return nil
	}

	cfg, ok := config.(cache_models.MapCacheConfig)

	if !ok {
		return errors.New("invalid map cache configuration: expected a cache_models.MapCacheConfig type")
	}

	r.cfg = &cfg
	r.cache = make(map[string]mapCacheItem)

	return nil
}

// Set stores data in the cache with an expiration time.
func (r *MapCacheDriver) Set(key string, value interface{}, expiration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Serialize the value if serialization is enabled.
	if r.cfg.Serialize {
		serializedValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		value = serializedValue
	}

	// Set expiration time.
	exp := time.Now().Add(expiration)
	r.cache[key] = mapCacheItem{
		value:      value,
		expiration: exp,
	}

	return nil
}

// Get retrieves data from the cache by key.
func (r *MapCacheDriver) Get(key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, found := r.cache[key]
	if !found {
		return nil, nil // Key does not exist.
	}

	// Remove and return error if the key has expired.
	if time.Now().After(item.expiration) {
		delete(r.cache, key)
		return nil, errors.New("key expired")
	}

	// Deserialize the value if serialization is enabled.
	if r.cfg.Serialize {
		var deserializedValue interface{}
		err := json.Unmarshal(item.value.([]byte), &deserializedValue)
		if err != nil {
			return nil, err
		}
		return deserializedValue, nil
	}

	return item.value, nil
}

// Has checks if a key exists in the cache and has not expired.
func (r *MapCacheDriver) Has(key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, found := r.cache[key]
	if !found {
		return false, nil
	}

	// Remove expired key and return false.
	if time.Now().After(item.expiration) {
		delete(r.cache, key)
		return false, nil
	}

	return true, nil
}

// Delete removes data from the cache by key.
func (r *MapCacheDriver) Delete(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, key)
	return nil
}
