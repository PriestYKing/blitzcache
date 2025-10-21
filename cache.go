package blitzcache

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultShardCount   = 256
	DefaultTickDuration = 100 * time.Millisecond
)

type CacheItem struct {
	Value     []byte
	ExpiresAt int64
	Version   uint64
}

type Shard struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
	count atomic.Int64
}

type Cache struct {
	shards      []*Shard
	shardCount  uint32
	timingWheel *TimingWheel
	stats       *Stats
}

type Stats struct {
	hits      atomic.Int64
	misses    atomic.Int64
	sets      atomic.Int64
	deletes   atomic.Int64
	evictions atomic.Int64
}

func NewCache(shardCount int) *Cache {
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}

	c := &Cache{
		shards:     make([]*Shard, shardCount),
		shardCount: uint32(shardCount),
		stats:      &Stats{},
	}

	for i := 0; i < shardCount; i++ {
		c.shards[i] = &Shard{
			items: make(map[string]*CacheItem, 1024),
		}
	}

	c.timingWheel = NewTimingWheel(DefaultTickDuration, 1*time.Hour, c.onExpire)
	c.timingWheel.Start()

	return c
}

func (c *Cache) getShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.shards[h.Sum32()%c.shardCount]
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {
	shard := c.getShard(key)
	now := time.Now().UnixNano()

	item := &CacheItem{
		Value:     value,
		ExpiresAt: now + ttl.Nanoseconds(),
		Version:   1,
	}

	shard.mu.Lock()
	if existing, ok := shard.items[key]; ok {
		item.Version = existing.Version + 1
	} else {
		shard.count.Add(1)
	}
	shard.items[key] = item
	shard.mu.Unlock()

	if ttl > 0 {
		c.timingWheel.Add(key, ttl)
	}

	c.stats.sets.Add(1)
	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) {
	shard := c.getShard(key)

	shard.mu.RLock()
	item, ok := shard.items[key]
	shard.mu.RUnlock()

	if !ok {
		c.stats.misses.Add(1)
		return nil, false
	}

	now := time.Now().UnixNano()
	if item.ExpiresAt > 0 && now >= item.ExpiresAt {
		c.Delete(key)
		c.stats.misses.Add(1)
		return nil, false
	}

	c.stats.hits.Add(1)
	return item.Value, true
}

func (c *Cache) Delete(key string) bool {
	shard := c.getShard(key)

	shard.mu.Lock()
	_, ok := shard.items[key]
	if ok {
		delete(shard.items, key)
		shard.count.Add(-1)
		c.stats.deletes.Add(1)
	}
	shard.mu.Unlock()

	return ok
}

func (c *Cache) onExpire(key string) {
	c.Delete(key)
	c.stats.evictions.Add(1)
}

func (c *Cache) Stats() map[string]int64 {
	return map[string]int64{
		"hits":      c.stats.hits.Load(),
		"misses":    c.stats.misses.Load(),
		"sets":      c.stats.sets.Load(),
		"deletes":   c.stats.deletes.Load(),
		"evictions": c.stats.evictions.Load(),
		"keys":      c.Count(),
	}
}

func (c *Cache) Count() int64 {
	var total int64
	for _, shard := range c.shards {
		total += shard.count.Load()
	}
	return total
}

func (c *Cache) Flush() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.items = make(map[string]*CacheItem, 1024)
		shard.count.Store(0)
		shard.mu.Unlock()
	}
}

func (c *Cache) Close() {
	c.timingWheel.Stop()
}
