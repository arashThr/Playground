## Challenge 5: Rate Limiter with Multiple Algorithms

**Problem**: Implement a configurable rate limiter supporting multiple algorithms (token bucket and sliding window) with Redis-like interface and metrics.

**Requirements**:
1. `RateLimiter` interface with `Allow(key string) (bool, *RateInfo)` method
2. Two implementations: `TokenBucketLimiter` and `SlidingWindowLimiter`
3. `RateInfo` struct with remaining tokens, reset time, and retry after
4. Thread-safe operations with per-key limiting
5. Background cleanup of expired keys
6. Configurable limits and time windows
7. `GetStats(key string) *RateInfo` method

**Test Cases**:
```go
func TestTokenBucketLimiter(t *testing.T) {
    limiter := NewTokenBucketLimiter(5, time.Second) // 5 requests per second
    defer limiter.Close()
    
    key := "user123"
    
    // Should allow first 5 requests
    for i := 0; i < 5; i++ {
        allowed, info := limiter.Allow(key)
        if !allowed {
            t.Errorf("Request %d should be allowed", i+1)
        }
        if info.Remaining != 4-i {
            t.Errorf("Expected %d remaining, got %d", 4-i, info.Remaining)
        }
    }
    
    // 6th request should be denied
    allowed, info := limiter.Allow(key)
    if allowed {
        t.Error("6th request should be denied")
    }
    if info.Remaining != 0 {
        t.Errorf("Expected 0 remaining, got %d", info.Remaining)
    }
    
    // Wait for token refill and try again
    time.Sleep(time.Second + 100*time.Millisecond)
    allowed, _ = limiter.Allow(key)
    if !allowed {
        t.Error("Request after refill should be allowed")
    }
}

func TestSlidingWindowLimiter(t *testing.T) {
    limiter := NewSlidingWindowLimiter(3, time.Second) // 3 requests per second
    defer limiter.Close()
    
    key := "user456"
    
    // Make 3 requests quickly
    for i := 0; i < 3; i++ {
        allowed, _ := limiter.Allow(key)
        if !allowed {
            t.Errorf("Request %d should be allowed", i+1)
        }
    }
    
    // 4th request should be denied
    allowed, info := limiter.Allow(key)
    if allowed {
        t.Error("4th request should be denied")
    }
    
    // Wait for window to slide
    time.Sleep(time.Second + 100*time.Millisecond)
    allowed, _ = limiter.Allow(key)
    if !allowed {
        t.Error("Request after window slide should be allowed")
    }
}

func TestRateLimiterConcurrency(t *testing.T) {
    limiter := NewTokenBucketLimiter(10, time.Second)
    defer limiter.Close()
    
    var wg sync.WaitGroup
    var allowed, denied int32
    
    // 20 goroutines making requests
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if success, _ := limiter.Allow("concurrent_test"); success {
                atomic.AddInt32(&allowed, 1)
            } else {
                atomic.AddInt32(&denied, 1)
            }
        }()
    }
    
    wg.Wait()
    
    if allowed != 10 {
        t.Errorf("Expected 10 allowed requests, got %d", allowed)
    }
    if denied != 10 {
        t.Errorf("Expected 10 denied requests, got %d", denied)
    }
}
```

**Suggested Structure**:
```go
type RateInfo struct {
    Remaining   int           // Tokens/requests remaining
    ResetTime   time.Time     // When the limit resets
    RetryAfter  time.Duration // How long to wait before retry
}

type RateLimiter interface {
    Allow(key string) (bool, *RateInfo)
    GetStats(key string) *RateInfo
    Close()
}

type TokenBucketLimiter struct {
    // Your fields here
}

type SlidingWindowLimiter struct {
    // Your fields here
}

func NewTokenBucketLimiter(limit int, window time.Duration) *TokenBucketLimiter
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter
```

**Relevant Go packages**:
- `sync` - RWMutex for thread safety
- `time` - For rate limiting logic and cleanup
- `sync/atomic` - For atomic counters
- `context` - For cleanup coordination

**Key concepts to practice**:
- Rate limiting algorithms
- Interface design and multiple implementations
- Per-key state management
- Background cleanup goroutines
- Thread-safe map operations
- Time-based calculations

This challenge tests system design knowledge, algorithm implementation, and Go concurrency patterns - very common in backend infrastructure interviews!