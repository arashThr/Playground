package ratelimitter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
		allowed, info := limiter.Allow(key)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}

		if info.Remaining != 2-i {
			t.Errorf("Expected %d remaining, got %d", 2-i, info.Remaining)
		}
	}

	// 4th request should be denied
	allowed, info := limiter.Allow(key)
	if allowed {
		t.Error("4th request should be denied")
	}

	if info.Remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", info.Remaining)
	}

	// Wait for window to slide
	time.Sleep(time.Second + 100*time.Millisecond)
	allowed, info = limiter.Allow(key)
	if !allowed {
		t.Error("Request after window slide should be allowed")
	}

	if info.Remaining != 2 {
		t.Errorf("Expected 2 remaining, got %d", info.Remaining)
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

func TestTokenBucketConcurrency(t *testing.T) {
	limiter := NewTokenBucketLimiter(100, time.Second)
	defer limiter.Close()

	var wg sync.WaitGroup
	var allowed int32
	numGoroutines := 200
	key := "stress_test"

	// All goroutines start at the same time
	start := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // Wait for signal

			if success, _ := limiter.Allow(key); success {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}

	close(start) // Start all goroutines simultaneously
	wg.Wait()

	// Should allow exactly 100, not more due to race conditions
	if allowed != 100 {
		t.Errorf("Expected exactly 100 allowed, got %d", allowed)
	}
}

/*------------------ IMPLEMENTATION ------------------*/

type RateInfo struct {
	Remaining  int           // Tokens/requests remaining
	ResetTime  time.Time     // When the limit resets
	RetryAfter time.Duration // How long to wait before retry
}

type RateLimiter interface {
	Allow(key string) (bool, *RateInfo)
	GetStats(key string) *RateInfo
	Close()
}

type TokenBucketLimiter struct {
	capacity   int
	refillRate float64 // Tokens per second
	mu         sync.RWMutex
	buckets    map[string]*Bucket
	cleaner    *time.Ticker
	done       chan bool
}

type Bucket struct {
	key        string
	tokens     int
	lastRefill time.Time
	lastUsed   time.Time
}

func NewTokenBucketLimiter(limit int, window time.Duration) *TokenBucketLimiter {
	ticker := time.NewTicker(window)
	bucketLimiter := TokenBucketLimiter{
		capacity:   limit,
		buckets:    make(map[string]*Bucket),
		cleaner:    ticker,
		done:       make(chan bool),
		refillRate: float64(limit) / window.Seconds(),
	}
	go bucketLimiter.cleanup()
	return &bucketLimiter
}

func (l *TokenBucketLimiter) cleanup() {
	for {
		select {
		case <-l.cleaner.C:
			l.mu.Lock()
			cutoff := time.Now().Add(-5 * time.Minute)
			for key, bucket := range l.buckets {
				if bucket.lastUsed.Before(cutoff) {
					delete(l.buckets, key)
				}
			}
			l.mu.Unlock()
		case <-l.done:
			return
		}
	}
}

func (l *TokenBucketLimiter) Allow(key string) (bool, *RateInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	bucket, found := l.buckets[key]
	if !found {
		bucket = &Bucket{
			key:        key,
			tokens:     l.capacity - 1,
			lastRefill: now,
			lastUsed:   now,
		}
		l.buckets[key] = bucket
		return true, l.getStatsUnlocked(bucket, now)
	}

	l.refillTokens(bucket)

	if bucket.tokens > 0 {
		bucket.tokens -= 1
		return true, l.getStatsUnlocked(bucket, now)
	}
	return false, l.getStatsUnlocked(bucket, now)
}

func (l *TokenBucketLimiter) refillTokens(bucket *Bucket) {
	elapsed := time.Since(bucket.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	floor := elapsed * l.refillRate
	refill := min(l.capacity, int(floor)+bucket.tokens)
	bucket.tokens = refill
}

func (l *TokenBucketLimiter) GetStats(key string) *RateInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	now := time.Now()
	r, found := l.buckets[key]
	if !found {
		return &RateInfo{
			Remaining:  l.capacity,
			ResetTime:  now.Add(time.Duration(float64(time.Second) / l.refillRate)),
			RetryAfter: 0,
		}
	}
	l.refillTokens(r)
	return l.getStatsUnlocked(r, now)
}

func (l *TokenBucketLimiter) getStatsUnlocked(bucket *Bucket, now time.Time) *RateInfo {
	var retryAfter time.Duration
	if bucket.tokens == 0 {
		// Time until next token is available
		timeForOneToken := time.Duration(float64(time.Second) / l.refillRate)
		retryAfter = timeForOneToken
	}

	resetTime := now.Add(time.Duration(float64(l.capacity-bucket.tokens)/l.refillRate) * time.Second)

	return &RateInfo{
		Remaining:  bucket.tokens,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}

func (l *TokenBucketLimiter) Close() {
	l.cleaner.Stop()
	l.done <- true
}

// Sliding Window Limiter Implementation

type SlidingWindowLimiter struct {
	limit      int
	windowSize time.Duration
	mu         sync.RWMutex
	requests   map[string]*Window
	cleaner    *time.Ticker
	done       chan bool
}

type Window struct {
	key      string
	access   []time.Time
	lastUsed time.Time
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	cleaner := time.NewTicker(time.Second)
	slider := SlidingWindowLimiter{
		limit:      limit,
		windowSize: window,
		requests:   map[string]*Window{},
		cleaner:    cleaner,
		done:       make(chan bool),
	}
	go slider.clean()
	return &slider
}

func (l *SlidingWindowLimiter) Allow(key string) (bool, *RateInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	window, found := l.requests[key]
	if !found {
		window = &Window{
			key:    key,
			access: []time.Time{now},
		}
		l.requests[key] = window
		return true, l.getStatsUnlocked(window, now)
	}

	window.access = l.cleanupWindow(window, now)
	window.lastUsed = now
	if len(window.access) >= l.limit {
		return false, l.getStatsUnlocked(window, now)
	}
	window.access = append(window.access, now)
	return true, l.getStatsUnlocked(window, now)
}

func (l *SlidingWindowLimiter) GetStats(key string) *RateInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	now := time.Now()
	r, found := l.requests[key]
	if !found {
		return &RateInfo{
			Remaining:  l.limit,
			ResetTime:  now.Add(l.windowSize),
			RetryAfter: 0,
		}
	}
	return l.getStatsUnlocked(r, now)
}

func (l *SlidingWindowLimiter) cleanupWindow(r *Window, now time.Time) []time.Time {
	windowAccesses := []time.Time{}
	windowStart := now.Add(-l.windowSize)
	// Access list should be sorted
	for i, access := range r.access {
		if !access.Before(windowStart) {
			windowAccesses = r.access[i:]
			break
		}
	}
	return windowAccesses
}

func (l *SlidingWindowLimiter) getStatsUnlocked(r *Window, now time.Time) *RateInfo {
	windowAccesses := l.cleanupWindow(r, now)

	// Update the accesses
	r.access = windowAccesses
	remaining := max(l.limit-len(windowAccesses), 0)

	var resetTime time.Time
	var retryAfter time.Duration
	if len(windowAccesses) > 0 {
		resetTime = windowAccesses[0].Add(l.windowSize)
		retryAfter = time.Until(resetTime)
	} else {
		resetTime = now.Add(l.windowSize)
		retryAfter = 0
	}

	return &RateInfo{
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}

func (l *SlidingWindowLimiter) clean() {
	for {
		select {
		case <-l.cleaner.C:
			l.mu.Lock()
			cutoff := time.Now().Add(-5 * time.Minute)
			for key, window := range l.requests {
				if len(window.access) > 0 && window.access[len(window.access)-1].Before(cutoff) {
					delete(l.requests, key)
				}
			}
			l.mu.Unlock()
		case <-l.done:
			return
		}
	}
}

func (l *SlidingWindowLimiter) Close() {
	l.cleaner.Stop()
	l.done <- true
}
