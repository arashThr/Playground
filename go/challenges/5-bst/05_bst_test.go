package bst

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestBSTBasicOperations(t *testing.T) {
	bst := NewBST[int]()

	// Insert nodes
	bst.Insert(5).Insert(3).Insert(7).Insert(1).Insert(9).Insert(3)
	bst.TraverseTree("Insertion")

	if !bst.Search(3) {
		t.Error("Expected to find 3 in tree")
	}

	if bst.Search(4) {
		t.Error("Should not find 4 in tree")
	}

	if bst.Size() != 5 {
		t.Errorf("Expected size 5, got %d", bst.Size())
	}

	if bst.Height() != 3 {
		t.Errorf("Expected height 3, got %d", bst.Height())
	}
}

func TestBSTTraversals(t *testing.T) {
	bst := NewBST[int]()
	bst.Insert(5).Insert(3).Insert(7).Insert(1).Insert(9)

	inOrder := bst.InOrder()
	expected := []int{1, 3, 5, 7, 9}
	if !slices.Equal(inOrder, expected) {
		t.Errorf("InOrder expected %v, got %v", expected, inOrder)
	}

	preOrder := bst.PreOrder()
	expectedPre := []int{5, 3, 1, 7, 9}
	if !slices.Equal(preOrder, expectedPre) {
		t.Errorf("PreOrder expected %v, got %v", expectedPre, preOrder)
	}
}

func TestBSTJSON(t *testing.T) {
	bst := NewBST[int]()
	bst.Insert(5).Insert(3).Insert(7)

	// Test marshaling
	data, err := json.Marshal(bst)
	fmt.Printf("Serialized data: %s\n", data)
	if err != nil {
		t.Fatalf("Failed to marshal BST: %v", err)
	}

	// Test unmarshaling
	newBST := NewBST[int]()
	err = json.Unmarshal(data, newBST)
	if err != nil {
		t.Fatalf("Failed to unmarshal BST: %v", err)
	}

	if !slices.Equal(bst.InOrder(), newBST.InOrder()) {
		t.Error("Unmarshaled tree should match original")
	}
}

func TestBSTValidation(t *testing.T) {
	bst := NewBST[int]()
	bst.Insert(5).Insert(3).Insert(7).Insert(1).Insert(9)

	if !bst.IsValid() {
		t.Error("Valid BST should pass validation")
	}

	// Manually break the BST property
	bst.Root.Left.Right = &Node[int]{Value: 6} // Invalid placement
	if bst.IsValid() {
		t.Error("Invalid BST should fail validation")
	}
}

/* ----------- IMPLEMENTATION -------------*/

type Node[T cmp.Ordered] struct {
	Value T
	Left  *Node[T]
	Right *Node[T]
}

type BST[T cmp.Ordered] struct {
	Root     *Node[T]
	TreeSize int
}

func NewBST[T cmp.Ordered]() *BST[T] {
	bst := BST[T]{
		Root:     nil,
		TreeSize: 0,
	}
	return &bst
}

func (bst *BST[T]) Insert(value T) *BST[T] {
	node := bst.insert(bst.Root, value)
	if bst.Root == nil {
		bst.Root = node
	}
	return bst
}

func (bst *BST[T]) insert(node *Node[T], value T) *Node[T] {
	if node == nil {
		bst.TreeSize += 1
		node = &Node[T]{
			Value: value,
		}
	} else if value < node.Value {
		node.Left = bst.insert(node.Left, value)
	} else if value > node.Value {
		node.Right = bst.insert(node.Right, value)
	}
	return node
}

func (bst *BST[T]) Search(value T) bool {
	return bst.search(bst.Root, value)
}

func (bst *BST[T]) search(node *Node[T], value T) bool {
	if node == nil {
		return false
	}
	if node.Value == value {
		return true
	}
	if node.Value > value {
		return bst.search(node.Left, value)
	}
	return bst.search(node.Right, value)
}

func (bst *BST[T]) Size() int {
	return bst.TreeSize
}

func (bst *BST[T]) Height() int {
	return bst.height(bst.Root, 0)
}

func (bst *BST[T]) height(node *Node[T], height int) int {
	if node == nil {
		return height
	}
	maxLeft := bst.height(node.Left, height+1)
	maxRight := bst.height(node.Right, height+1)
	if maxLeft > maxRight {
		return maxLeft
	}
	return maxRight
}

func (bst *BST[T]) InOrder() []T {
	return bst.inOrder(bst.Root, []T{})
}

func (bst *BST[T]) inOrder(node *Node[T], list []T) []T {
	if node == nil {
		return list
	}
	list = bst.inOrder(node.Left, list)
	list = append(list, node.Value)
	list = bst.inOrder(node.Right, list)
	return list
}

func (bst *BST[T]) PreOrder() []T {
	return bst.preOrder(bst.Root, []T{})
}

func (bst *BST[T]) preOrder(node *Node[T], list []T) []T {
	if node == nil {
		return list
	}
	list = append(list, node.Value)
	list = bst.preOrder(node.Left, list)
	list = bst.preOrder(node.Right, list)
	return list
}

func (bst *BST[T]) TraverseTree(message string) {
	fmt.Println(message)
	bst.print(bst.Root)
	fmt.Println()
}

func (bst *BST[T]) print(node *Node[T]) {
	if node == nil {
		return
	}
	bst.print(node.Left)
	fmt.Printf(" %v ", node.Value)
	bst.print(node.Right)
}

type SerializedTree struct {
	Tree string
	Size int
}

func (bst *BST[T]) MarshalJSON() ([]byte, error) {
	serialized := SerializedTree{
		Tree: bst.marshalJSON(bst.Root),
		Size: bst.Size(),
	}
	result, err := json.Marshal(serialized)
	if err != nil {
		return nil, fmt.Errorf("marshal BST: %v", err)
	}
	return result, nil
}

func (bst *BST[T]) marshalJSON(node *Node[T]) string {
	if node == nil {
		return "#"
	}
	v, _ := json.Marshal(node.Value)
	return strings.Join([]string{string(v), bst.marshalJSON(node.Left), bst.marshalJSON(node.Right)}, " ")
}

func (bst *BST[T]) UnmarshalJSON(input []byte) error {
	var serialized SerializedTree
	err := json.Unmarshal(input, &serialized)
	if err != nil {
		return fmt.Errorf("unmarshal BST: %v", err)
	}
	fields := strings.Fields(serialized.Tree)
	bst.Root, _, err = bst.unmarshalJSON(fields)
	if err != nil {
		return fmt.Errorf("unmarshal tree: %v", err)
	}
	bst.TraverseTree("Unmarshal")
	return nil
}

func (bst *BST[T]) unmarshalJSON(values []string) (*Node[T], []string, error) {
	if len(values) == 0 {
		return nil, []string{}, nil
	}
	val := values[0]
	if val == "#" {
		return nil, values[1:], nil
	}
	var t T
	err := json.Unmarshal([]byte(values[0]), &t)
	if err != nil {
		return nil, values, fmt.Errorf("unmarshal value %s: %v", val, err)
	}
	node := &Node[T]{
		Value: t,
	}
	left, rest, err := bst.unmarshalJSON(values[1:])
	if err != nil {
		return nil, rest, fmt.Errorf("unmarshal left child of %v: %v", t, err)
	}
	node.Left = left
	right, rest, err := bst.unmarshalJSON(rest)
	if err != nil {
		return nil, rest, fmt.Errorf("unmarshal right child of %v: %v", t, err)
	}
	node.Right = right
	return node, rest, nil
}

func (bst *BST[T]) IsValid() bool {
	if bst.Root == nil {
		return true
	}
	// or I could do pre-order traverse and check if the result is sorted
	return bst.isBigger(bst.Root, bst.Root.Left) && bst.isSmaller(bst.Root, bst.Root.Right)
}

func (bst *BST[T]) isBigger(node, left *Node[T]) bool {
	if left == nil {
		return true
	}
	return node.Value > left.Value && bst.isBigger(node, left.Left) && bst.isBigger(node, left.Right) && bst.isBigger(left, left.Left) && bst.isSmaller(left, left.Right)
}

func (bst *BST[T]) isSmaller(node, right *Node[T]) bool {
	if right == nil {
		return true
	}
	return node.Value < right.Value && bst.isSmaller(node, right.Left) && bst.isSmaller(node, right.Right) && bst.isBigger(right, right.Left) && bst.isSmaller(right, right.Right)
}
