package blitzcache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/PriestYKing/blitzcache"
)

func BenchmarkCacheSet(b *testing.B) {
	cache := blitzcache.NewCache(256)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, []byte("value"), 10*time.Second)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := blitzcache.NewCache(256)
	defer cache.Close()

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, []byte("value"), 10*time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%10000)
		cache.Get(key)
	}
}

func BenchmarkCacheConcurrent(b *testing.B) {
	cache := blitzcache.NewCache(256)
	defer cache.Close()

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, []byte("value"), 10*time.Second)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%10000)
			cache.Get(key)
			i++
		}
	})
}

func TestSubMillisecondLatency(t *testing.T) {
	cache := blitzcache.NewCache(256)
	defer cache.Close()

	const iterations = 100000
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(key, []byte("value"), 10*time.Second)
	}

	for i := 0; i < iterations; i++ {
		key := fmt.Sprintf("key_%d", i)
		start := time.Now()
		cache.Get(key)
		totalDuration += time.Since(start)
	}

	avgLatency := totalDuration / iterations
	t.Logf("Average latency: %v", avgLatency)

	if avgLatency > 1*time.Millisecond {
		t.Errorf("Average latency %v exceeds 1ms threshold", avgLatency)
	}
}
