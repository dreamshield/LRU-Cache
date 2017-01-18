package cache

import (
	"errors"
	"time"
)

const (
	// default cache expired time
	CacheExpired = 1 * time.Hour
	// cache que max element size for data to store
	CacheMaxElementSize = 1024
	// evey ten minutes to clear all expired nodes
	CacheGcInterval   = 10 * time.Minute
	CacheGcMaxRemoved = 100
)

var (
	ErrCacheMiss error = errors.New("cache: key not found.")
	ErrNotStored error = errors.New("cache: not stored.")
)

// CacheStore is a interface to store cache
type CacheStore interface {
	Put(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
}

// Cacher is an interface to provide cache
type Cacher interface {
	GetBean(tableName string, sql string) interface{}
	PutBean(tableName string, sql string, obj interface{})
	DelBean(tableName string, sql string)
	ClearBeans(tableName string)
}
