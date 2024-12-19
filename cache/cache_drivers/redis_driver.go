package cache_drivers

import (
	"errors"
	"fmt"
	"time"

	"github.com/HemendCo/go-core/cache/cache_models"

	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// RedisCacheDriver is a structure for managing caching using Redis.
type RedisCacheDriver struct {
	client *redis.Client
	cfg    *cache_models.RedisCacheConfig
}

// Name returns the name of the cache driver.
func (r *RedisCacheDriver) Name() string {
	return "redis"
}

// Init initializes the Redis cache driver with the provided configuration.
func (r *RedisCacheDriver) Init(config interface{}) error {
	if r.client != nil {
		return nil
	}

	cfg, ok := config.(cache_models.RedisCacheConfig)

	if !ok {
		return errors.New("invalid redis cache configuration: expected a cache_models.RedisCacheConfig type")
	}

	r.cfg = &cfg
	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", r.cfg.Host, r.cfg.Port),
		Username: r.cfg.Username,
		Password: r.cfg.Password,
		DB:       r.cfg.Database,
	})

	return nil
}

// Set stores data in Redis with an expiration time.
func (r *RedisCacheDriver) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves data from Redis by key.
func (r *RedisCacheDriver) Get(key string) (interface{}, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Key does not exist.
	} else if err != nil {
		return nil, err
	}
	return val, nil
}

// Has checks if a key exists in Redis.
func (r *RedisCacheDriver) Has(key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// Delete removes data from Redis by key.
func (r *RedisCacheDriver) Delete(key string) error {
	return r.client.Del(ctx, key).Err()
}
