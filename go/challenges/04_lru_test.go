package challenges

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLRUBasicOperations(t *testing.T) {
	cache := NewLRUCache[string, int](3) // capacity of 3
	defer cache.Close()

	// Test basic set/get
	cache.Set("a", 1, time.Hour)
	cache.Set("b", 2, time.Hour)
	cache.Set("c", 3, time.Hour)

	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected a=1, got %v, %v", val, ok)
	}

	// Test LRU eviction
	cache.Set("d", 4, time.Hour) // Should evict "b" (least recently used)

	if _, ok := cache.Get("b"); ok {
		t.Error("Expected 'b' to be evicted")
	}

	// Test that "a" is still there (was accessed recently)
	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Error("Expected 'a' to still be in cache")
	}
}

func TestLRUTTL(t *testing.T) {
	cache := NewLRUCache[string, int](5)
	defer cache.Close()

	// Set with short TTL
	cache.Set("short", 100, 50*time.Millisecond)
	cache.Set("long", 200, time.Hour)

	// Should be available immediately
	if val, ok := cache.Get("short"); !ok || val != 100 {
		t.Error("Expected short-lived item to be available immediately")
	}

	// Wait for TTL expiration
	time.Sleep(100 * time.Millisecond)

	if _, ok := cache.Get("short"); ok {
		t.Error("Expected short-lived item to be expired")
	}

	if val, ok := cache.Get("long"); !ok || val != 200 {
		t.Error("Expected long-lived item to still be available")
	}
}

func TestLRUStats(t *testing.T) {
	cache := NewLRUCache[string, int](2)
	defer cache.Close()

	cache.Set("a", 1, time.Hour)
	cache.Get("a")       // hit
	cache.Get("missing") // miss
	cache.Set("b", 2, time.Hour)
	cache.Set("c", 3, time.Hour) // should evict "a"

	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
}

/* --------------- IMPLEMENTATION ------------------*/

type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
}

type atomicStats struct {
	Hits        atomic.Int64
	Misses      atomic.Int64
	Evictions   atomic.Int64
	Expirations atomic.Int64
}

type LRUCache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	lru      *list.List
	items    map[K]*list.Element
	stats    atomicStats
	ticker   *time.Ticker
	done     chan bool
}

type Entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	t := time.NewTicker(10 * time.Second)
	cache := LRUCache[K, V]{
		capacity: capacity,
		lru:      list.New(),
		items:    make(map[K]*list.Element),
		ticker:   t,
		done:     make(chan bool),
	}
	go cache.cleanUp()
	return &cache
}

func (c *LRUCache[K, V]) cleanUp() {
	for {
		select {
		case <-c.done:
			return
		case <-c.ticker.C:
			// Once a ticker is stopped it won’t receive any more values on its channel
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.items {
				el := c.getEntry(v)
				if el.expiresAt.Before(now) {
					c.stats.Expirations.Add(1)
					c.deleteUnlocked(k)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, found := c.items[key]
	if found {
		e := c.getEntry(el)
		e.expiresAt = time.Now().Add(ttl)
		e.value = value
		c.lru.MoveToFront(el)
		return
	}
	if len(c.items) >= c.capacity {
		el := c.lru.Back()
		key := el.Value.(*Entry[K, V]).key
		c.stats.Evictions.Add(1)
		c.deleteUnlocked(key)
	}
	entry := &Entry[K, V]{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	el = c.lru.PushFront(entry)
	c.items[key] = el
}

func (c *LRUCache[K, V]) Iterate() {
	for e := c.lru.Front(); e != nil; e = e.Next() {
		fmt.Printf("%v\n", c.getEntry(e).key)
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var v V
	el, found := c.items[key]
	if found {
		e := c.getEntry(el)
		if e.expiresAt.Before(time.Now()) {
			c.deleteUnlocked(key)
			c.stats.Expirations.Add(1)
			c.stats.Misses.Add(1)
			return v, false
		}
		v = e.value
		c.lru.MoveToFront(el)
		c.stats.Hits.Add(1)
		return v, true
	}
	c.stats.Misses.Add(1)
	return v, false
}

func (c *LRUCache[K, V]) getEntry(el *list.Element) *Entry[K, V] {
	e, ok := el.Value.(*Entry[K, V])
	if !ok {
		panic("unexpected type in LRU list")
	}
	return e
}

func (c *LRUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deleteUnlocked(key)
}

func (c *LRUCache[K, V]) deleteUnlocked(key K) bool {
	el, found := c.items[key]
	if !found {
		return false
	}
	c.lru.Remove(el)
	delete(c.items, key)
	return true
}

func (c *LRUCache[K, V]) GetStats() CacheStats {
	return CacheStats{
		Hits:        c.stats.Hits.Load(),
		Misses:      c.stats.Misses.Load(),
		Evictions:   c.stats.Evictions.Load(),
		Expirations: c.stats.Expirations.Load(),
	}
}

func (c *LRUCache[K, V]) Close() {
	c.ticker.Stop()
	c.done <- true
}
