package interpreter

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Mock expression types for testing
type mockBinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator Token
}

func (m *mockBinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", m.Left.String(), m.Operator.String(), m.Right.String())
}

func (m *mockBinaryExpression) Position() Position {
	return Position{Line: 1, Column: 1}
}

type mockUnaryExpression struct {
	Operand  Expression
	Operator Token
}

func (m *mockUnaryExpression) String() string {
	return fmt.Sprintf("%s%s", m.Operator.String(), m.Operand.String())
}

func (m *mockUnaryExpression) Position() Position {
	return Position{Line: 1, Column: 1}
}

type mockLiteralExpression struct {
	Value Value
}

func (m *mockLiteralExpression) String() string {
	return fmt.Sprintf("%v", m.Value)
}

func (m *mockLiteralExpression) Position() Position {
	return Position{Line: 1, Column: 1}
}

type mockVariableExpression struct {
	Name string
}

func (m *mockVariableExpression) String() string {
	return m.Name
}

func (m *mockVariableExpression) Position() Position {
	return Position{Line: 1, Column: 1}
}

type mockCallExpression struct {
	Function  string
	Arguments []Expression
}

func (m *mockCallExpression) String() string {
	return fmt.Sprintf("%s(...)", m.Function)
}

func (m *mockCallExpression) Position() Position {
	return Position{Line: 1, Column: 1}
}

func TestNewExpressionOptimizer(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)

	if optimizer == nil {
		t.Error("Expected optimizer to be created")
	}

	if optimizer.maxCacheSize != 100 {
		t.Errorf("Expected max cache size 100, got %d", optimizer.maxCacheSize)
	}

	if len(optimizer.constantCache) != 0 {
		t.Errorf("Expected empty constant cache, got %d entries", len(optimizer.constantCache))
	}

	if len(optimizer.subexprCache) != 0 {
		t.Errorf("Expected empty subexpression cache, got %d entries", len(optimizer.subexprCache))
	}
}

func TestConstantFolder_FoldConstants_Literal(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test literal expression
	literal := &Literal{Value: 42, pos: Position{Line: 1, Column: 1}}
	result, ok := folder.FoldConstants(literal)

	if !ok {
		t.Error("Expected literal folding to succeed")
	}

	if result != 42 {
		t.Errorf("Expected result 42, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinaryAddition(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test binary addition with integer literals
	left := &Literal{Value: 10, pos: Position{Line: 1, Column: 1}}
	right := &Literal{Value: 20, pos: Position{Line: 1, Column: 1}}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
		pos:      Position{Line: 1, Column: 1},
	}

	result, ok := folder.FoldConstants(binary)

	if !ok {
		t.Error("Expected binary addition folding to succeed")
	}

	if result != 30 {
		t.Errorf("Expected result 30, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinarySubtraction(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test binary subtraction with integer literals
	left := &Literal{Value: 30}
	right := &Literal{Value: 10}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: MINUS,
	}

	result, ok := folder.FoldConstants(binary)

	if !ok {
		t.Error("Expected binary subtraction folding to succeed")
	}

	if result != 20 {
		t.Errorf("Expected result 20, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinaryMultiplication(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test binary multiplication with integer literals
	left := &Literal{Value: 5}
	right := &Literal{Value: 6}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: TIMES,
	}

	result, ok := folder.FoldConstants(binary)

	if !ok {
		t.Error("Expected binary multiplication folding to succeed")
	}

	if result != 30 {
		t.Errorf("Expected result 30, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinaryDivision(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test binary division with integer literals
	left := &Literal{Value: 20}
	right := &Literal{Value: 4}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: DIVIDE,
	}

	result, ok := folder.FoldConstants(binary)

	if !ok {
		t.Error("Expected binary division folding to succeed")
	}

	if result != 5 {
		t.Errorf("Expected result 5, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinaryDivisionByZero(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test binary division by zero
	left := &Literal{Value: 20}
	right := &Literal{Value: 0}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: DIVIDE,
	}

	_, ok := folder.FoldConstants(binary)

	if ok {
		t.Error("Expected division by zero to fail")
	}
}

func TestConstantFolder_FoldConstants_BinaryStringConcatenation(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test string concatenation
	left := &Literal{Value: "hello", pos: Position{Line: 1, Column: 1}}
	right := &Literal{Value: " world", pos: Position{Line: 1, Column: 1}}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
		pos:      Position{Line: 1, Column: 1},
	}

	result, ok := folder.FoldConstants(binary)

	if !ok {
		t.Error("Expected string concatenation folding to succeed")
	}

	if result != "hello world" {
		t.Errorf("Expected result 'hello world', got %v", result)
	}
}

func TestConstantFolder_FoldConstants_BinaryComparison(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	tests := []struct {
		left     Value
		right    Value
		operator Token
		expected Value
	}{
		{10, 20, LT, true},
		{20, 10, LT, false},
		{10, 10, LTE, true},
		{20, 10, GT, true},
		{10, 20, GT, false},
		{10, 10, GTE, true},
		{10, 10, EQUAL, true},
		{10, 20, EQUAL, false},
		{10, 20, NOTEQUAL, true},
		{10, 10, NOTEQUAL, false},
	}

	for _, test := range tests {
		left := &Literal{Value: test.left, pos: Position{Line: 1, Column: 1}}
		right := &Literal{Value: test.right, pos: Position{Line: 1, Column: 1}}
		binary := &Binary{
			Left:     left,
			Right:    right,
			Operator: test.operator,
			pos:      Position{Line: 1, Column: 1},
		}

		result, ok := folder.FoldConstants(binary)

		if !ok {
			t.Errorf("Expected %v %s %v folding to succeed", test.left, test.operator, test.right)
			continue
		}

		if result != test.expected {
			t.Errorf("Expected %v %s %v = %v, got %v", test.left, test.operator, test.right, test.expected, result)
		}
	}
}

func TestConstantFolder_FoldConstants_UnaryNegation(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test unary negation with integer
	operand := &Literal{Value: 42, pos: Position{Line: 1, Column: 1}}
	unary := &Unary{
		Operand:  operand,
		Operator: MINUS,
		pos:      Position{Line: 1, Column: 1},
	}

	result, ok := folder.FoldConstants(unary)

	if !ok {
		t.Error("Expected unary negation folding to succeed")
	}

	if result != -42 {
		t.Errorf("Expected result -42, got %v", result)
	}
}

func TestConstantFolder_FoldConstants_UnaryLogicalNot(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Test unary logical not with boolean
	operand := &Literal{Value: true, pos: Position{Line: 1, Column: 1}}
	unary := &Unary{
		Operand:  operand,
		Operator: NOT,
		pos:      Position{Line: 1, Column: 1},
	}

	result, ok := folder.FoldConstants(unary)

	if !ok {
		t.Error("Expected unary logical not folding to succeed")
	}

	if result != false {
		t.Errorf("Expected result false, got %v", result)
	}
}

func TestConstantFolder_Caching(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Create the same expression twice
	left := &Literal{Value: 10}
	right := &Literal{Value: 20}
	binary1 := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}
	binary2 := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}

	// First call should compute and cache
	result1, ok1 := folder.FoldConstants(binary1)
	if !ok1 || result1 != 30 {
		t.Error("First call should succeed")
	}

	// Second call should use cache
	result2, ok2 := folder.FoldConstants(binary2)
	if !ok2 || result2 != 30 {
		t.Error("Second call should succeed")
	}

	// Check optimization stats
	hits, _, _ := optimizer.GetOptimizationStats()
	if hits < 1 {
		t.Error("Expected at least one cache hit")
	}
}

func TestCommonSubexpressionEliminator_GenerateSignature(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	cse := NewCommonSubexpressionEliminator(optimizer)

	// Test literal signature
	literal := &Literal{Value: 42, pos: Position{Line: 1, Column: 1}}
	signature := cse.generateExpressionSignature(literal)
	if !strings.Contains(signature, "42") {
		t.Errorf("Expected signature to contain '42', got '%s'", signature)
	}

	// Test variable signature
	variable := &Variable{Name: "x", pos: Position{Line: 1, Column: 1}}
	signature = cse.generateExpressionSignature(variable)
	if !strings.Contains(signature, "x") {
		t.Errorf("Expected signature to contain 'x', got '%s'", signature)
	}

	// Test unary signature
	operand := &Literal{Value: 42, pos: Position{Line: 1, Column: 1}}
	unary := &Unary{
		Operator: MINUS,
		Operand:  operand,
		pos:      Position{Line: 1, Column: 1},
	}
	signature = cse.generateExpressionSignature(unary)
	if !strings.Contains(signature, "42") {
		t.Errorf("Expected signature to contain '42', got '%s'", signature)
	}
}

func TestCommonSubexpressionEliminator_CacheSubexpression(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	cse := NewCommonSubexpressionEliminator(optimizer)

	expr := &Literal{Value: 42}
	result := 42

	// Cache the subexpression
	cse.CacheSubexpression(expr, result)

	// Check if it's cached
	cachedResult, found := cse.EliminateCommonSubexpressions(expr)
	if !found {
		t.Error("Expected subexpression to be found in cache")
	}

	if cachedResult != result {
		t.Errorf("Expected cached result %v, got %v", result, cachedResult)
	}
}

func TestExpressionOptimizer_CacheEviction(t *testing.T) {
	optimizer := NewExpressionOptimizer(2) // Small cache size
	folder := NewConstantFolder(optimizer)

	// Fill cache beyond capacity
	for i := 0; i < 5; i++ {
		left := &Literal{Value: i}
		right := &Literal{Value: i + 1}
		binary := &Binary{
			Left:     left,
			Right:    right,
			Operator: PLUS,
		}
		folder.FoldConstants(binary)
	}

	// Cache should have been partially cleared
	if len(optimizer.constantCache) > 2 {
		t.Errorf("Expected cache size <= 2, got %d", len(optimizer.constantCache))
	}
}

func TestExpressionOptimizer_GetOptimizationStats(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)

	// Perform some operations
	left := &Literal{Value: 10}
	right := &Literal{Value: 20}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}

	// First call (miss)
	folder.FoldConstants(binary)

	// Second call (hit)
	folder.FoldConstants(binary)

	hits, misses, hitRatio := optimizer.GetOptimizationStats()

	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}

	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}

	expectedRatio := 0.5
	if hitRatio != expectedRatio {
		t.Errorf("Expected hit ratio %f, got %f", expectedRatio, hitRatio)
	}
}

func TestExpressionOptimizer_ClearCaches(t *testing.T) {
	optimizer := NewExpressionOptimizer(100)
	folder := NewConstantFolder(optimizer)
	cse := NewCommonSubexpressionEliminator(optimizer)

	// Add some entries to caches
	left := &Literal{Value: 10}
	right := &Literal{Value: 20}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}

	folder.FoldConstants(binary)
	cse.CacheSubexpression(left, 10)

	// Clear caches
	optimizer.ClearCaches()

	// Check if caches are empty
	if len(optimizer.constantCache) != 0 {
		t.Errorf("Expected empty constant cache, got %d entries", len(optimizer.constantCache))
	}

	if len(optimizer.subexprCache) != 0 {
		t.Errorf("Expected empty subexpression cache, got %d entries", len(optimizer.subexprCache))
	}

	hits, misses, _ := optimizer.GetOptimizationStats()
	if hits != 0 || misses != 0 {
		t.Error("Expected stats to be reset")
	}
}

// Benchmark tests
func BenchmarkConstantFolder_FoldConstants(b *testing.B) {
	optimizer := NewExpressionOptimizer(1000)
	folder := NewConstantFolder(optimizer)

	left := &Literal{Value: 10}
	right := &Literal{Value: 20}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		folder.FoldConstants(binary)
	}
}

func BenchmarkCommonSubexpressionEliminator_GenerateSignature(b *testing.B) {
	optimizer := NewExpressionOptimizer(1000)
	cse := NewCommonSubexpressionEliminator(optimizer)

	left := &Literal{Value: 10}
	right := &Literal{Value: 20}
	binary := &Binary{
		Left:     left,
		Right:    right,
		Operator: PLUS,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cse.generateExpressionSignature(binary)
	}
}

// Test concurrent access
func TestExpressionOptimizer_ConcurrentAccess(t *testing.T) {
	optimizer := NewExpressionOptimizer(1000)
	folder := NewConstantFolder(optimizer)

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				// Use modulo to create duplicate expressions that will result in cache hits
				left := &Literal{Value: id % 3}
				right := &Literal{Value: j % 5}
				binary := &Binary{
					Left:     left,
					Right:    right,
					Operator: PLUS,
				}
				folder.FoldConstants(binary)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Check that some optimizations occurred
	hits, _, _ := optimizer.GetOptimizationStats()
	if hits == 0 {
		t.Error("Expected some optimization hits")
	}
}
