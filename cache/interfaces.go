package cache

import "time"

type CacheDriver interface {
	Name() string
	Init(config interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (interface{}, error)
	Has(key string) (bool, error)
	Delete(key string) error
}
