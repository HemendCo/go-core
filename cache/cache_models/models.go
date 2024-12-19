package cache_models

type FileCacheConfig struct {
	Path      string
	Serialize bool
}

type MapCacheConfig struct {
	Path      string
	Serialize bool
}

type RedisCacheConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database int
}
