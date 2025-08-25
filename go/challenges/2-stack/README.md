## Challenge 2: Custom Stack with Generic Types

**Problem**: Implement a generic stack data structure with proper error handling and interface compliance.

**Requirements**:
1. Implement a generic `Stack[T]` type
2. Methods: `Push(item T)`, `Pop() (T, error)`, `Peek() (T, error)`, `Size() int`, `IsEmpty() bool`
3. Implement the `fmt.Stringer` interface
4. Create a custom error type `StackError` 
5. Support method chaining where appropriate
6. Add a `Clear()` method that returns the stack for chaining

**Test Cases**:
```go
func TestStackBasicOperations(t *testing.T) {
    stack := NewStack[int]()
    
    // Test empty stack
    if !stack.IsEmpty() {
        t.Error("New stack should be empty")
    }
    
    _, err := stack.Pop()
    if err == nil {
        t.Error("Pop from empty stack should return error")
    }
    
    // Test push and peek
    stack.Push(1).Push(2).Push(3)
    
    if stack.Size() != 3 {
        t.Errorf("Expected size 3, got %d", stack.Size())
    }
    
    top, err := stack.Peek()
    if err != nil || top != 3 {
        t.Errorf("Expected top to be 3, got %v with error %v", top, err)
    }
    
    // Size should remain same after peek
    if stack.Size() != 3 {
        t.Error("Peek should not change stack size")
    }
}

func TestStackString(t *testing.T) {
    stack := NewStack[string]()
    stack.Push("hello").Push("world")
    
    str := stack.String()
    // Should show stack representation (format is up to you)
    if str == "" {
        t.Error("String() should not return empty string")
    }
}

func TestStackError(t *testing.T) {
    stack := NewStack[int]()
    _, err := stack.Pop()
    
    var stackErr *StackError
    if !errors.As(err, &stackErr) {
        t.Error("Error should be of type StackError")
    }
}
```

**Relevant Go packages to explore**:
- `fmt` - for Stringer interface and error formatting
- `errors` - for error handling and error wrapping
- `strings` - for string building (if needed for String() method)

**Key concepts to practice**:
- Generic types (`[T any]`)
- Custom error types
- Interface implementation
- Method chaining (fluent interface)
- Error handling patterns
- Zero values and type constraints

This challenge will test your understanding of modern Go generics, proper error handling, and interface design - all critical for backend interviews!