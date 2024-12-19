package cache

import (
	"fmt"
	"github.com/HemendCo/go-core/cache/cache_drivers"
)

type CacheManager struct {
	drivers map[string]CacheDriver
}

func NewCacheManager(drivers ...CacheDriver) *CacheManager {
	manager := &CacheManager{
		drivers: make(map[string]CacheDriver),
	}

	// register default driver
	drivers = append(drivers, &cache_drivers.FileCacheDriver{}, &cache_drivers.MapCacheDriver{}, &cache_drivers.RedisCacheDriver{})
	manager.RegisterDrivers(drivers...)

	return manager
}

func (dm *CacheManager) RegisterDrivers(drivers ...CacheDriver) {
	for _, driver := range drivers {
		dm.drivers[driver.Name()] = driver
	}
}

func (dm *CacheManager) CreateCacheFactory(driverName string, config interface{}) (CacheDriver, error) {
	driver, exists := dm.drivers[driverName]
	if !exists {
		return nil, fmt.Errorf("unsupported cache driver %s", driverName)
	}

	if err := driver.Init(config); err != nil {
		return nil, err
	}

	return driver, nil
}
