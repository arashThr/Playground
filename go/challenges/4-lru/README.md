## Challenge 4: LRU Cache with TTL and Metrics

**Problem**: Implement a thread-safe LRU (Least Recently Used) cache with time-to-live (TTL) expiration and basic metrics collection.

**Requirements**:
1. Generic `LRUCache[K comparable, V any]` with configurable capacity
2. Methods: `Set(key K, value V, ttl time.Duration)`, `Get(key K) (V, bool)`, `Delete(key K) bool`
3. Automatic TTL-based expiration with background cleanup
4. Thread-safe operations using appropriate synchronization
5. Metrics: hit rate, miss rate, eviction count, expired count
6. `GetStats() CacheStats` method
7. `Close()` method for cleanup

**Test Cases**:
```go
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
    cache.Get("a")      // hit
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
```

**Suggested Structure**:
```go
type CacheStats struct {
    Hits       int64
    Misses     int64
    Evictions  int64
    Expirations int64
}

type cacheItem[V any] struct {
    value     V
    expiresAt time.Time
}

type LRUCache[K comparable, V any] struct {
    // Your fields here - think about what you need for:
    // - LRU ordering (hint: doubly linked list + map)
    // - Thread safety
    // - TTL cleanup
    // - Metrics
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V]
func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration)
func (c *LRUCache[K, V]) Get(key K) (V, bool)
func (c *LRUCache[K, V]) Delete(key K) bool
func (c *LRUCache[K, V]) GetStats() CacheStats
func (c *LRUCache[K, V]) Close()
```

**Relevant Go packages**:
- `sync` - RWMutex for thread safety
- `time` - For TTL and cleanup timers
- `container/list` - For efficient LRU ordering
- `sync/atomic` - For atomic counter operations

**Key concepts to practice**:
- Data structure design (LRU with hash map + doubly linked list)
- Thread safety patterns
- Background goroutines and cleanup
- Atomic operations for metrics
- Generic constraints (`comparable`)
- Resource cleanup patterns

This is a classic systems programming problem that tests algorithm knowledge, concurrency, and Go-specific patterns!