package stack

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

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
	if str != "{hello world}" {
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

func TestStackClear(t *testing.T) {
	stack := NewStack[int]()
	stack.Push(1).Push(2).Push(3)

	result := stack.Clear()
	if result != stack { // Compare pointers
		t.Error("Clear should return the same stack instance")
	}
	if !stack.IsEmpty() {
		t.Error("Stack should be empty after Clear")
	}
}

/* ------------- IMPLEMENTATIONS ------------ */

type StackError struct {
	message string
}

func (se *StackError) Error() string {
	return "stack error: " + se.message
}

func NewStackError(message string) error {
	err := &StackError{message: message}
	return err
}

type Stack[T any] struct {
	elements []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) IsEmpty() bool {
	return s.Size() == 0
}

func (s *Stack[T]) Pop() (T, error) {
	var e T
	if s.Size() == 0 {
		return e, NewStackError("no elements in stack")
	}
	e, s.elements = s.elements[s.Size()-1], s.elements[:s.Size()-1]
	return e, nil
}

func (s *Stack[T]) Push(p T) *Stack[T] {
	s.elements = append(s.elements, p)
	return s
}

func (s *Stack[T]) Size() int {
	return len(s.elements)
}

func (s *Stack[T]) Peek() (T, error) {
	if s.Size() == 0 {
		return *new(T), fmt.Errorf("stack is empty")
	}
	return s.elements[len(s.elements)-1], nil
}

func (s *Stack[T]) String() string {
	str := []string{}
	for _, e := range s.elements {
		str = append(str, fmt.Sprintf("%v", e))
	}
	return "{" + strings.Join(str, " ") + "}"
}

func (s *Stack[T]) Clear() *Stack[T] {
	s.elements = s.elements[:0] // Reuse underlying array
	return s
}
