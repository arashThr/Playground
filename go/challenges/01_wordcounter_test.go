package challenges

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
)

type WordCount struct {
	Word  string
	Count int
}

func CountWords(text string) map[string]int {
	parts := strings.Fields(text)
	result := map[string]int{}
	for _, p := range parts {
		trimmed := strings.Trim(p, "!.,")
		result[trimmed] += 1
	}
	return result
}

func TopWords(frequencies map[string]int, n int) []WordCount {
	if n <= 0 {
		return []WordCount{}
	}

	words := []WordCount{}
	for k, v := range frequencies {
		words = append(words, WordCount{
			Word:  k,
			Count: v,
		})
	}

	slices.SortFunc(words, func(a, b WordCount) int {
		return b.Count - a.Count
	})

	if len(words) <= n {
		return words
	}
	return words[:n]
}

func TestCountWords(t *testing.T) {
	text := "Hello world! Hello Go programming. Go is great, Go is powerful."
	expected := map[string]int{
		"hello": 2, "world": 1, "go": 3, "programming": 1,
		"is": 2, "great": 1, "powerful": 1,
	}
	result := CountWords(text)
	if !maps.Equal(result, expected) {
		fmt.Printf("Got: %v\n", result)
	}
}

func TestTopWords(t *testing.T) {
	frequencies := map[string]int{"go": 3, "hello": 2, "is": 2, "world": 1}
	result := TopWords(frequencies, 2)
	// Should return [{"go", 3}, {"hello", 2}] or [{"go", 3}, {"is", 2}]
	// (either "hello" or "is" for second place is acceptable)

	if len(result) != 2 {
		t.Fail()
		return
	}

	if result[0].Word != "go" {
		t.Fail()
		return
	}

	if result[1].Word != "hello" && result[1].Word != "is" {
		t.Fail()
		return
	}
}
