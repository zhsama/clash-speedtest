package unlock

import (
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	cache     sync.Map      // key: string, value: *CacheEntry
	ttl       time.Duration
	stats     *CacheStats
	statsMutex sync.RWMutex
}

// NewMemoryCache 创建新的内存缓存实例
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		ttl: ttl,
		stats: &CacheStats{
			CreatedAt: time.Now(),
		},
	}

	// 启动清理goroutine
	go cache.startCleanup()

	return cache
}

// NewUnlockCache 创建新的缓存实例（向后兼容）
func NewUnlockCache() UnlockCache {
	return NewMemoryCache(30 * time.Minute)
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) *UnlockResult {
	value, exists := c.cache.Load(key)
	if !exists {
		c.updateStats(false)
		return nil
	}

	entry := value.(*CacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		c.cache.Delete(key)
		c.updateStats(false)
		return nil
	}

	c.updateStats(true)
	return entry.Result
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, result *UnlockResult, duration time.Duration) {
	if duration <= 0 {
		duration = c.ttl
	}

	entry := &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(duration),
	}
	c.cache.Store(key, entry)
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.cache.Delete(key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.cache.Range(func(key, value interface{}) bool {
		c.cache.Delete(key)
		return true
	})
	
	c.statsMutex.Lock()
	c.stats.Hits = 0
	c.stats.Misses = 0
	c.statsMutex.Unlock()
}

// Stats 获取缓存统计信息
func (c *MemoryCache) Stats() CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	
	stats := *c.stats
	stats.Entries = c.countEntries()
	
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(total)
	}
	
	return stats
}

// countEntries 计算缓存条目数量
func (c *MemoryCache) countEntries() int {
	count := 0
	c.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// updateStats 更新统计信息
func (c *MemoryCache) updateStats(hit bool) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	
	if hit {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}
}

// startCleanup 启动定期清理过期缓存
func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired 清理过期的缓存项
func (c *MemoryCache) cleanupExpired() {
	now := time.Now()
	var keysToDelete []interface{}

	c.cache.Range(func(key, value interface{}) bool {
		entry := value.(*CacheEntry)
		if now.After(entry.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
		}
		return true
	})

	for _, key := range keysToDelete {
		c.cache.Delete(key)
	}
}
