package data

import (
	"fmt"

	"github.com/sjzsdu/wn/config"
)

type StorageType string

const (
	JSONStorageType   StorageType = "json"
	SQLiteStorageType StorageType = "sqlite"
)

var defaultCacheManager *CacheManager

func init() {
	// 从配置中获取存储类型，默认使用 JSON
	storageType := StorageType(config.GetConfig("storage_type"))
	if storageType == "" {
		storageType = JSONStorageType
	}

	manager, err := NewCacheManagerWithType(storageType)
	if err != nil {
		fmt.Printf("Failed to create default cache manager: %v\n", err)
		return
	}

	defaultCacheManager = manager
}

// GetDefaultCacheManager 获取默认的缓存管理器
func GetDefaultCacheManager() *CacheManager {
	return defaultCacheManager
}

// SetDefaultCacheManager 设置默认的缓存管理器
func SetDefaultCacheManager(manager *CacheManager) {
	defaultCacheManager = manager
}

// NewCacheManagerWithType 根据存储类型创建缓存管理器
func NewCacheManagerWithType(storageType StorageType) (*CacheManager, error) {
	var storage CacheStorage
	switch storageType {
	case JSONStorageType:
		storage = NewJSONStorage()
	case SQLiteStorageType:
		storage = NewSQLiteStorage()
	default:
		storage = NewJSONStorage() // 默认使用 JSON 存储
	}

	return NewCacheManager(storage)
}

func init() {
	// 从配置中获取存储类型，默认使用 JSON
	storageType := StorageType(config.GetConfig("storage_type"))
	if storageType == "" {
		storageType = JSONStorageType
	}

	manager, err := NewCacheManagerWithType(storageType)
	if err != nil {
		fmt.Printf("Failed to create default cache manager: %v\n", err)
		return
	}

	defaultCacheManager = manager
}
