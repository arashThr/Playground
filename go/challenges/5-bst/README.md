## Challenge 4: Binary Tree with JSON Serialization

**Problem**: Implement a binary search tree with JSON serialization/deserialization, tree validation, and multiple traversal methods.

**Requirements**:
1. Generic `BST[T constraints.Ordered]` with `Insert`, `Search`, `Delete` methods
2. JSON marshaling/unmarshaling using custom `MarshalJSON`/`UnmarshalJSON`
3. Tree validation method `IsValid() bool`
4. Three traversal methods: `InOrder()`, `PreOrder()`, `PostOrder()` returning slices
5. `Height()` and `Size()` methods
6. Handle edge cases (empty tree, single node, duplicates)

**Test Cases**:
```go
func TestBSTBasicOperations(t *testing.T) {
    bst := NewBST[int]()
    
    // Insert nodes
    bst.Insert(5).Insert(3).Insert(7).Insert(1).Insert(9)
    
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
}
```

**Suggested Structure**:
```go
type Node[T constraints.Ordered] struct {
    Value T
    Left  *Node[T]
    Right *Node[T]
}

type BST[T constraints.Ordered] struct {
    root *Node[T]
    size int
}

func NewBST[T constraints.Ordered]() *BST[T]
func (bst *BST[T]) Insert(value T) *BST[T]
func (bst *BST[T]) Search(value T) bool
func (bst *BST[T]) Delete(value T) bool
func (bst *BST[T]) InOrder() []T
func (bst *BST[T]) PreOrder() []T
func (bst *BST[T]) PostOrder() []T
func (bst *BST[T]) Height() int
func (bst *BST[T]) Size() int
func (bst *BST[T]) IsValid() bool
func (bst *BST[T]) MarshalJSON() ([]byte, error)
func (bst *BST[T]) UnmarshalJSON(data []byte) error
```

**Relevant Go packages**:
- `golang.org/x/exp/constraints` - For `Ordered` constraint
- `encoding/json` - JSON marshaling interfaces
- `slices` - For comparison in tests

**Key concepts to practice**:
- Tree algorithms (insertion, deletion, traversal)
- Custom JSON marshaling
- Generic constraints beyond `any`
- Recursive algorithms
- Tree validation algorithms

This challenge combines algorithms, Go's type system, and standard library interfaces - core backend interview material!