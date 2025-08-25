Great approach to interview preparation! Let's start with a foundational challenge that will help you get comfortable with Go's core constructs and standard library.

## Challenge 1: Word Frequency Counter

**Problem**: Implement a function that takes a text string and returns a map of word frequencies, along with the top N most frequent words.

**Requirements**:
1. Function `CountWords(text string) map[string]int` - returns word frequencies
2. Function `TopWords(frequencies map[string]int, n int) []WordCount` - returns top N words
3. Handle case-insensitive counting
4. Strip punctuation and handle multiple whitespace
5. Define a `WordCount` struct with `Word string` and `Count int` fields

**Test Cases**:
```go
func TestCountWords(t *testing.T) {
    text := "Hello world! Hello Go programming. Go is great, Go is powerful."
    expected := map[string]int{
        "hello": 2, "world": 1, "go": 3, "programming": 1,
        "is": 2, "great": 1, "powerful": 1,
    }
    result := CountWords(text)
    // Assert maps are equal
}

func TestTopWords(t *testing.T) {
    frequencies := map[string]int{"go": 3, "hello": 2, "is": 2, "world": 1}
    result := TopWords(frequencies, 2)
    // Should return [{"go", 3}, {"hello", 2}] or [{"go", 3}, {"is", 2}]
    // (either "hello" or "is" for second place is acceptable)
}
```

**Relevant Go packages to explore**:
- `strings` - for text manipulation
- `sort` - for sorting the results
- `regexp` - for advanced text cleaning (optional)
- `unicode` - for character classification

This challenge will help you practice:
- Maps and slices
- String manipulation
- Sorting custom types
- Struct definitions
- Method receivers (if you want to add methods to WordCount)

Go ahead and implement this! Focus on clean, idiomatic Go code. I'll review your solution and point out any improvements or Go-specific best practices.