package configs

import (
	"time"
)

// структура конфига для инмемори шардированных кэшэй с TTL
type CacheConfig struct {
	NumOfShards         int
	SearchCacheTTL      time.Duration
	SearchCacheCleanUp  time.Duration
	VacancyCacheTTL     time.Duration
	VacancyCacheCleanUp time.Duration
	MaxMemoryUsageMB    int
}

// функция, которая возвращает указатель на дэфолтный конфиг для кэшэй
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		NumOfShards:         7,
		SearchCacheTTL:      60 * time.Second,
		SearchCacheCleanUp:  30 * time.Second,
		VacancyCacheTTL:     60 * time.Second,
		VacancyCacheCleanUp: 30 * time.Second,
	}
}
