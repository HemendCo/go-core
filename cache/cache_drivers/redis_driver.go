package cache_drivers

import (
	"errors"
	"fmt"
	"github.com/HemendCo/go-core/cache/cache_models"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisCacheDriver struct {
	client *redis.Client
	cfg    *cache_models.RedisCacheConfig
}

// Name implements cache.CacheDriver.
func (r *RedisCacheDriver) Name() string {
	return "redis"
}

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

// Set ذخیره داده‌ها در Redis
func (r *RedisCacheDriver) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get بازیابی داده‌ها از Redis
func (r *RedisCacheDriver) Get(key string) (interface{}, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // key does not exist
	} else if err != nil {
		return nil, err
	}
	return val, nil
}

// Has بررسی می‌کند آیا کلید در Redis وجود دارد یا خیر
func (r *RedisCacheDriver) Has(key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// Delete حذف داده‌ها از Redis
func (r *RedisCacheDriver) Delete(key string) error {
	return r.client.Del(ctx, key).Err()
}
