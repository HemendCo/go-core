package cache_drivers

import (
	"HemendCo/go-core/cache/cache_models"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// mapCacheItem نگهدارنده داده‌های کش با زمان انقضاء
type mapCacheItem struct {
	value      interface{}
	expiration time.Time
}

// MapCacheDriver ساختار برای نگهداری کش در حافظه
type MapCacheDriver struct {
	cache map[string]mapCacheItem
	cfg   *cache_models.MapCacheConfig
	mu    sync.RWMutex
}

// Name implements cache.CacheDriver.
func (r *MapCacheDriver) Name() string {
	return "map"
}

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

// Set ذخیره داده‌ها در Redis
func (r *MapCacheDriver) Set(key string, value interface{}, expiration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// اگر گزینه serialize فعال باشد، مقدار سریالایز می‌شود
	if r.cfg.Serialize {
		serializedValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		value = serializedValue
	}

	// تنظیم زمان انقضاء
	exp := time.Now().Add(expiration)
	r.cache[key] = mapCacheItem{
		value:      value,
		expiration: exp,
	}

	return nil
}

// Get بازیابی داده‌ها از Redis
func (r *MapCacheDriver) Get(key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, found := r.cache[key]
	if !found {
		return nil, nil // key does not exist
	}

	// اگر زمان انقضا گذشته باشد، مقدار حذف می‌شود
	if time.Now().After(item.expiration) {
		delete(r.cache, key)
		return nil, errors.New("key expired")
	}

	// اگر گزینه serialize فعال باشد، مقدار Deserialize می‌شود
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

func (r *MapCacheDriver) Has(key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, found := r.cache[key]
	if !found {
		return false, nil
	}

	// اگر زمان انقضاء گذشته باشد، کلید حذف می‌شود و false برمی‌گرداند
	if time.Now().After(item.expiration) {
		delete(r.cache, key)
		return false, nil
	}

	return true, nil
}

// Delete حذف داده‌ها از Redis
func (r *MapCacheDriver) Delete(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, key)
	return nil
}
