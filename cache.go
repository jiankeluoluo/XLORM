package xlorm

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/spaolacci/murmur3"
)

const (
	defaultLRUCacheSize = 1024 // 默认每个分片的 LRU 缓存大小
)

// Cache 缓存接口定义
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

// shardedCache 分片锁缓存
type shardedCache struct {
	shards     []*shard                       // 动态分片切片
	shardCount uint32                         // 分片数量（CPU核心数）
	lruCaches  []*lru.Cache[string, []string] // LRU 缓存
	lruLocks   []sync.RWMutex                 // 专门为 LRU 缓存的读写锁
}

type shard struct {
	sync.RWMutex
	m      map[string][]string
	hits   atomic.Uint64
	misses atomic.Uint64
}

func newShardedCache() *shardedCache {
	// 动态计算分片数（基于 CPU 核心数，上限 64）
	numShards := runtime.NumCPU()
	if numShards < 1 {
		numShards = 1
	} else if numShards > 64 {
		numShards = 64
	}

	c := &shardedCache{
		shards:     make([]*shard, numShards),
		shardCount: uint32(numShards),
		lruCaches:  make([]*lru.Cache[string, []string], numShards),
		lruLocks:   make([]sync.RWMutex, numShards),
	}

	for i := 0; i < numShards; i++ {
		c.shards[i] = &shard{
			m: make(map[string][]string),
		}

		// 为每个分片创建 LRU 缓存
		lruCache, err := lru.New[string, []string](defaultLRUCacheSize)
		if err != nil {
			panic(fmt.Sprintf("创建 LRU 缓存失败: %v", err))
		}
		c.lruCaches[i] = lruCache
	}
	return c
}

func (c *shardedCache) Get(key string) ([]string, bool) {
	shardIndex := c.hash(key)
	shard := c.shards[shardIndex]
	lruCache := c.lruCaches[shardIndex]
	lruLock := &c.lruLocks[shardIndex]

	shard.RLock()
	defer shard.RUnlock()

	// 先从 LRU 缓存获取（使用专门的 LRU 锁）
	lruLock.RLock()
	if value, ok := lruCache.Get(key); ok {
		lruLock.RUnlock()
		shard.hits.Add(1)
		return value, true
	}
	lruLock.RUnlock()

	// 如果 LRU 缓存未命中，则从普通缓存获取
	if value, exists := shard.m[key]; exists {
		// 使用写锁将值加入 LRU 缓存
		lruLock.Lock()
		lruCache.Add(key, value)
		lruLock.Unlock()

		shard.hits.Add(1)
		return value, true
	}

	shard.misses.Add(1)
	return nil, false
}

func (c *shardedCache) Set(key string, value []string) {
	shardIndex := c.hash(key)
	shard := c.shards[shardIndex]
	lruCache := c.lruCaches[shardIndex]
	lruLock := &c.lruLocks[shardIndex]

	shard.Lock()
	defer shard.Unlock()

	// 更新普通缓存
	shard.m[key] = value

	// 使用专门的 LRU 锁更新 LRU 缓存
	lruLock.Lock()
	lruCache.Add(key, value)
	lruLock.Unlock()
}

func (c *shardedCache) Delete(key string) error {
	shardIndex := c.hash(key)
	shard := c.shards[shardIndex]
	lruCache := c.lruCaches[shardIndex]
	lruLock := &c.lruLocks[shardIndex]

	shard.Lock()
	defer shard.Unlock()

	// 删除普通缓存
	delete(shard.m, key)

	// 使用专门的 LRU 锁删除 LRU 缓存
	lruLock.Lock()
	lruCache.Remove(key)
	lruLock.Unlock()

	return nil
}

// 获取缓存统计信息
func (c *shardedCache) Stats() map[string]uint64 {
	stats := make(map[string]uint64)
	shardCount := int(atomic.LoadUint32(&c.shardCount)) // 原子读取当前分片数

	// 遍历所有分片收集统计信息
	for i := 0; i < shardCount; i++ {
		shard := c.shards[i]
		stats[fmt.Sprintf("shard_%d_hits", i)] = shard.hits.Load()
		stats[fmt.Sprintf("shard_%d_misses", i)] = shard.misses.Load()
	}
	return stats
}

// Clear 清理所有缓存并重置统计信息
func (c *shardedCache) Clear() {
	shardCount := int(atomic.LoadUint32(&c.shardCount)) // 原子读取当前分片数

	for i := 0; i < shardCount; i++ {
		shard := c.shards[i]
		lruCache := c.lruCaches[i]
		lruLock := &c.lruLocks[i]

		shard.Lock()
		shard.m = make(map[string][]string)
		shard.hits.Store(0)
		shard.misses.Store(0)
		lruLock.Lock()
		lruCache.Purge() // 清空 LRU 缓存
		lruLock.Unlock()
		shard.Unlock()
	}
}

// hash 计算键的哈希值，用于确定分片索引
func (c *shardedCache) hash(key string) uint32 {
	// 使用 MurmurHash3 获得更均匀的哈希分布
	hash := murmur3.Sum32([]byte(key))
	return hash % c.shardCount // 动态分片取模
}
