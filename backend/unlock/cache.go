package unlock

import (
	"sync"
	"time"
)

// CachedResult 缓存的检测结果
type CachedResult struct {
	Result    *UnlockResult
	ExpiresAt time.Time
}

// UnlockCache 解锁检测结果缓存
type UnlockCache struct {
	cache sync.Map // key: string, value: *CachedResult
	ttl   time.Duration
}

// NewUnlockCache 创建新的缓存实例
func NewUnlockCache() *UnlockCache {
	cache := &UnlockCache{
		ttl: 30 * time.Minute, // 缓存有效期30分钟
	}

	// 启动清理goroutine
	go cache.startCleanup()

	return cache
}

// Set 设置缓存
func (c *UnlockCache) Set(key string, result *UnlockResult) {
	cached := &CachedResult{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.cache.Store(key, cached)
}

// Get 获取缓存
func (c *UnlockCache) Get(key string) *UnlockResult {
	value, exists := c.cache.Load(key)
	if !exists {
		return nil
	}

	cached := value.(*CachedResult)
	if time.Now().After(cached.ExpiresAt) {
		c.cache.Delete(key)
		return nil
	}

	return cached.Result
}

// Clear 清空缓存
func (c *UnlockCache) Clear() {
	c.cache.Range(func(key, value interface{}) bool {
		c.cache.Delete(key)
		return true
	})
}

// startCleanup 启动定期清理过期缓存
func (c *UnlockCache) startCleanup() {
	ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired 清理过期的缓存项
func (c *UnlockCache) cleanupExpired() {
	now := time.Now()
	var keysToDelete []interface{}

	c.cache.Range(func(key, value interface{}) bool {
		cached := value.(*CachedResult)
		if now.After(cached.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
		}
		return true
	})

	for _, key := range keysToDelete {
		c.cache.Delete(key)
	}
}
