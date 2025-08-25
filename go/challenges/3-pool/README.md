## Challenge 3: Concurrent URL Fetcher with Worker Pool

**Problem**: Build a concurrent HTTP client that fetches multiple URLs using a worker pool pattern with proper error handling and timeouts.

**Requirements**:
1. `URLFetcher` struct with configurable worker count and timeout
2. `FetchResult` struct containing URL, response body, status code, and error
3. Method `FetchAll(urls []string) <-chan FetchResult` - returns a channel of results
4. Implement graceful shutdown with context cancellation
5. Handle HTTP timeouts and network errors properly
6. Add a `Close()` method for cleanup

**Test Cases**:
```go
func TestURLFetcher(t *testing.T) {
    fetcher := NewURLFetcher(3, 5*time.Second) // 3 workers, 5s timeout
    defer fetcher.Close()
    
    urls := []string{
        "https://httpbin.org/delay/1",
        "https://httpbin.org/status/200",
        "https://httpbin.org/status/404",
        "https://invalid-url-that-does-not-exist.com",
    }
    
    results := fetcher.FetchAll(urls)
    count := 0
    
    for result := range results {
        count++
        t.Logf("URL: %s, Status: %d, Error: %v", 
               result.URL, result.StatusCode, result.Error)
    }
    
    if count != len(urls) {
        t.Errorf("Expected %d results, got %d", len(urls), count)
    }
}

func TestURLFetcherWithContext(t *testing.T) {
    fetcher := NewURLFetcherWithContext(context.Background(), 2, 1*time.Second)
    defer fetcher.Close()
    
    urls := []string{
        "https://httpbin.org/delay/2", // This should timeout
        "https://httpbin.org/status/200",
    }
    
    results := fetcher.FetchAll(urls)
    timeoutCount := 0
    
    for result := range results {
        if result.Error != nil && 
           (strings.Contains(result.Error.Error(), "timeout") || 
            strings.Contains(result.Error.Error(), "context deadline")) {
            timeoutCount++
        }
    }
    
    if timeoutCount == 0 {
        t.Error("Expected at least one timeout error")
    }
}
```

**Suggested Structure**:
```go
type FetchResult struct {
    URL        string
    Body       string
    StatusCode int
    Error      error
}

type URLFetcher struct {
    // Your fields here
}

func NewURLFetcher(workers int, timeout time.Duration) *URLFetcher
func NewURLFetcherWithContext(ctx context.Context, workers int, timeout time.Duration) *URLFetcher
func (f *URLFetcher) FetchAll(urls []string) <-chan FetchResult
func (f *URLFetcher) Close() error
```

**Relevant Go packages**:
- `net/http` - HTTP client
- `context` - for cancellation and timeouts
- `sync` - WaitGroup for goroutine coordination
- `time` - for timeouts

**Key concepts to practice**:
- Worker pool pattern
- Channel communication
- Context handling
- HTTP client configuration
- Graceful shutdown
- Error propagation in concurrent code

This challenge tests your understanding of Go's concurrency primitives and real-world patterns used in backend services!