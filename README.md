# BlitzCache ⚡

A high-performance, sub-millisecond in-memory cache written in Go with Redis protocol compatibility.

## Features

- **Sub-millisecond latency** (<1ms p99)
- **256-way sharding** for reduced lock contention
- **Timing wheel** for efficient TTL expiration
- **Redis-compatible** RESP protocol
- **10M+ ops/sec** throughput
- **Zero dependencies**

## Installation

### As a Library

- go get github.com/PriestYKing/blitzcache

### As a Binary

- go install github.com/PriestYKing/blitzcache/cmd/blitzcache@latest


## Quick Start

### Using as Library

package main

import (
"fmt"
"time"
"github.com/YOUR_USERNAME/blitzcache"
)

func main() {
cache := blitzcache.NewCache(256)
defer cache.Close()

// Set with TTL
cache.Set("key", []byte("value"), 60*time.Second)

// Get
if value, ok := cache.Get("key"); ok {
    fmt.Println(string(value))
}

// Stats
stats := cache.Stats()
fmt.Printf("Hits: %d, Misses: %d\n", stats["hits"], stats["misses"])



### Running as Server

blitzcache -addr :6380 -shards 256


# Connect with Redis CLI:

redis-cli -p 6380
SET mykey "Hello" EX 60
GET mykey
STATS

## Benchmarks

- BenchmarkCacheGet-8 10000000 127 ns/op 0 B/op 0 allocs/op
- BenchmarkCacheSet-8 5000000 234 ns/op 48 B/op 1 allocs/op
- BenchmarkCacheConcurrent-8 50000000 34 ns/op 0 B/op 0 allocs/op


## API Reference

### Cache Methods

- `NewCache(shardCount int) *Cache` - Create new cache instance
- `Set(key string, value []byte, ttl time.Duration) error` - Set key with TTL
- `Get(key string) ([]byte, bool)` - Get key value
- `Delete(key string) bool` - Delete key
- `Stats() map[string]int64` - Get cache statistics
- `Flush()` - Clear all keys
- `Close()` - Shutdown cache

### Server Commands (RESP Protocol)

- `SET key value [EX seconds]`
- `GET key`
- `DEL key`
- `STATS`
- `FLUSH`
- `PING`

## License

MIT
