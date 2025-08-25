package pool

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestURLFetcher(t *testing.T) {
	fetcher := NewURLFetcher(3, 5*time.Second) // 3 workers, 5s timeout
	defer fetcher.Close()

	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://invalid-url-123.com",
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

/* ------------------ IMPLEMENTATION ---------------- */

type FetchResult struct {
	URL        string
	Body       string
	StatusCode int
	Error      error
}

type URLFetcher struct {
	workers int
	client  http.Client
	context context.Context
	cancel  context.CancelFunc
}

func NewURLFetcher(workers int, timeout time.Duration) *URLFetcher {
	return NewURLFetcherWithContext(context.Background(), workers, timeout)
}

func NewURLFetcherWithContext(ctx context.Context, workers int, timeout time.Duration) *URLFetcher {
	fetcherContext, cancel := context.WithCancel(ctx)
	client := http.Client{
		Timeout: timeout,
	}
	return &URLFetcher{
		workers: workers,
		client:  client,
		context: fetcherContext,
		cancel:  cancel,
	}
}

func (f *URLFetcher) FetchAll(urls []string) <-chan FetchResult {
	jobs := make(chan string, len(urls))
	results := make(chan FetchResult, len(urls)) // Buffered to prevent blocking

	var wg sync.WaitGroup
	for i := range f.workers {
		wg.Add(1)
		go f.fetcher(i, jobs, results, &wg)
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func (f *URLFetcher) fetcher(id int, jobs <-chan string, result chan<- FetchResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case url, ok := <-jobs:
			if !ok {
				return // channel is closed
			}
			p := f.getPage(url)
			result <- *p
		case <-f.context.Done():
			return
		}
	}
}

func (f *URLFetcher) getPage(url string) *FetchResult {
	result := FetchResult{URL: url}
	req, err := http.NewRequestWithContext(f.context, http.MethodGet, url, nil)
	if err != nil {
		result.Error = err
		return &result
	}
	resp, err := f.client.Do(req)
	if err != nil {
		result.Error = err
		return &result
	}
	defer resp.Body.Close() //If not closed can cause memory leak
	result.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return &result
	}
	result.Body = string(body)
	return &result
}

func (f *URLFetcher) Close() error {
	f.cancel()
	return nil
}
