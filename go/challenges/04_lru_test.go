package challenges

import (
	"fmt"
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

	cache.Iterate()
}

// func TestLRUTTL(t *testing.T) {
// 	cache := NewLRUCache[string, int](5)
// 	defer cache.Close()

// 	// Set with short TTL
// 	cache.Set("short", 100, 50*time.Millisecond)
// 	cache.Set("long", 200, time.Hour)

// 	// Should be available immediately
// 	if val, ok := cache.Get("short"); !ok || val != 100 {
// 		t.Error("Expected short-lived item to be available immediately")
// 	}

// 	// Wait for TTL expiration
// 	time.Sleep(100 * time.Millisecond)

// 	if _, ok := cache.Get("short"); ok {
// 		t.Error("Expected short-lived item to be expired")
// 	}

// 	if val, ok := cache.Get("long"); !ok || val != 200 {
// 		t.Error("Expected long-lived item to still be available")
// 	}
// }

// func TestLRUStats(t *testing.T) {
// 	cache := NewLRUCache[string, int](2)
// 	defer cache.Close()

// 	cache.Set("a", 1, time.Hour)
// 	cache.Get("a")       // hit
// 	cache.Get("missing") // miss
// 	cache.Set("b", 2, time.Hour)
// 	cache.Set("c", 3, time.Hour) // should evict "a"

// 	stats := cache.GetStats()
// 	if stats.Hits != 1 {
// 		t.Errorf("Expected 1 hit, got %d", stats.Hits)
// 	}
// 	if stats.Misses != 1 {
// 		t.Errorf("Expected 1 miss, got %d", stats.Misses)
// 	}
// 	if stats.Evictions != 1 {
// 		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
// 	}
// }

/* --------------- IMPLEMENTATION ------------------*/

type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
}

type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

type LRUCache[K comparable, V any] struct {
	head     *Node[K, V]
	capacity int
	list     map[K]*Node[K, V]
	// Your fields here - think about what you need for:
	// - LRU ordering (hint: doubly linked list + map)
	// - Thread safety
	// - TTL cleanup
	// - Metrics
}

type Node[K comparable, V any] struct {
	next *Node[K, V]
	prev *Node[K, V]
	key  K
	ttl  time.Duration
	item cacheItem[V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity < 2 {
		fmt.Printf("Capacity must be at least 2, got %d\n", capacity)
		return nil
	}
	cache := LRUCache[K, V]{
		head:     nil,
		capacity: capacity,
		list:     make(map[K]*Node[K, V]),
	}
	return &cache
}

func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) {
	var node Node[K, V]
	if oldItem, exists := c.list[key]; exists {
		node = *oldItem
	} else {
		fmt.Printf("Adding new key: %v - Size: %d - Cap: %d\n", key, len(c.list), c.capacity)
		if len(c.list) >= c.capacity {
			// Clean up the list
			var t *Node[K, V]
			// t will be the tail
			for t = c.head; t.next != nil; t = t.next {
			}
			// Question: How to make sure memory is freed
			t.prev.next = nil
			delete(c.list, t.key)
		}
		node = Node[K, V]{key: key}
	}
	node.ttl = ttl
	node.item = cacheItem[V]{value: value, expiresAt: time.Now().Add(ttl)}
	c.moveToHead(&node)
}

func (c *LRUCache[K, V]) Iterate() {
	for i := c.head; i != nil; i = i.next {
		fmt.Printf("%v -> %v - %v\n", i.key, i.item.value, i.item.expiresAt.Format(time.RFC3339))
	}
}

func (c *LRUCache[K, V]) moveToHead(node *Node[K, V]) {
	var head *Node[K, V]
	prev := node.prev
	next := node.next
	fmt.Printf("Moving to head: %v\n", node.key)
	if c.head != nil {
		head = c.head
		head.prev = node
	}
	if next != nil {
		next.prev = prev
	}
	if prev != nil {
		prev.next = next
	}
	node.next = head
	node.prev = nil
	c.head = node
	c.list[node.key] = node
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	fmt.Printf("Getting key: %v\n", key)
	var value V
	var node *Node[K, V]
	found := false
	if node, found = c.list[key]; found {
		node.item.expiresAt = time.Now().Add(node.ttl)
		value = node.item.value
		c.moveToHead(node)
	}
	return value, found
}

// func (c *LRUCache[K, V]) Delete(key K) bool {

// }

// func (c *LRUCache[K, V]) GetStats() CacheStats {

// }

func (c *LRUCache[K, V]) Close() {

}
