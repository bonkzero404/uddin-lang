package interpreter

import (
	"fmt"
	"sync"
)

// ExpressionOptimizer provides optimization for expressions
type ExpressionOptimizer struct {
	constantCache    map[string]Value
	subexprCache     map[string]Value
	mutex            sync.RWMutex
	maxCacheSize     int
	optimizationHits int64
	optimizationMiss int64
}

// NewExpressionOptimizer creates a new expression optimizer
func NewExpressionOptimizer(maxCacheSize int) *ExpressionOptimizer {
	return &ExpressionOptimizer{
		constantCache: make(map[string]Value),
		subexprCache:  make(map[string]Value),
		maxCacheSize:  maxCacheSize,
	}
}

// Global expression optimizer
var globalExpressionOptimizer = NewExpressionOptimizer(1000)

// GetGlobalExpressionOptimizer returns the global expression optimizer
func GetGlobalExpressionOptimizer() *ExpressionOptimizer {
	return globalExpressionOptimizer
}

// ConstantFolder performs constant folding optimization
type ConstantFolder struct {
	optimizer *ExpressionOptimizer
}

// NewConstantFolder creates a new constant folder
func NewConstantFolder(optimizer *ExpressionOptimizer) *ConstantFolder {
	return &ConstantFolder{optimizer: optimizer}
}

// FoldConstants attempts to fold constant expressions
func (cf *ConstantFolder) FoldConstants(expr Expression) (Value, bool) {
	switch e := expr.(type) {
	case *Binary:
		return cf.foldBinaryExpression(e)
	case *Unary:
		return cf.foldUnaryExpression(e)
	case *Literal:
		return e.Value, true
	default:
		return nil, false
	}
}

// foldBinaryExpression folds binary expressions with constant operands
func (cf *ConstantFolder) foldBinaryExpression(expr *Binary) (Value, bool) {
	// Check if both operands are literals
	leftLit, leftOk := expr.Left.(*Literal)
	rightLit, rightOk := expr.Right.(*Literal)
	
	if !leftOk || !rightOk {
		return nil, false
	}
	
	leftVal := leftLit.Value
	rightVal := rightLit.Value
	
	// Generate cache key
	cacheKey := fmt.Sprintf("binary:%v:%s:%v", leftVal, expr.Operator.String(), rightVal)
	
	// Check cache first
	cf.optimizer.mutex.RLock()
	if cached, exists := cf.optimizer.constantCache[cacheKey]; exists {
		cf.optimizer.optimizationHits++
		cf.optimizer.mutex.RUnlock()
		return cached, true
	}
	cf.optimizer.optimizationMiss++
	cf.optimizer.mutex.RUnlock()
	
	// Perform constant folding
	result, ok := cf.evaluateConstantBinary(leftVal, expr.Operator, rightVal)
	if !ok {
		return nil, false
	}
	
	// Cache the result
	cf.optimizer.mutex.Lock()
	defer cf.optimizer.mutex.Unlock()
	
	// Check cache size limit
	if len(cf.optimizer.constantCache) >= cf.optimizer.maxCacheSize {
		cf.clearHalfCache(cf.optimizer.constantCache)
	}
	
	cf.optimizer.constantCache[cacheKey] = result

	return result, true
}

// foldUnaryExpression folds unary expressions with constant operands
func (cf *ConstantFolder) foldUnaryExpression(expr *Unary) (Value, bool) {
	// Check if operand is a literal
	operandLit, ok := expr.Operand.(*Literal)
	if !ok {
		return nil, false
	}
	
	operandVal := operandLit.Value
	
	// Generate cache key
	cacheKey := fmt.Sprintf("unary:%s:%v", expr.Operator.String(), operandVal)
	
	// Check cache first
	cf.optimizer.mutex.RLock()
	if cached, exists := cf.optimizer.constantCache[cacheKey]; exists {
		cf.optimizer.optimizationHits++
		cf.optimizer.mutex.RUnlock()
		return cached, true
	}
	cf.optimizer.optimizationMiss++
	cf.optimizer.mutex.RUnlock()
	
	// Perform constant folding
	result, ok := cf.evaluateConstantUnary(expr.Operator, operandVal)
	if !ok {
		return nil, false
	}
	
	// Cache the result
	cf.optimizer.mutex.Lock()
	defer cf.optimizer.mutex.Unlock()
	
	// Check cache size limit
	if len(cf.optimizer.constantCache) >= cf.optimizer.maxCacheSize {
		cf.clearHalfCache(cf.optimizer.constantCache)
	}
	
	cf.optimizer.constantCache[cacheKey] = result

	return result, true
}

// evaluateConstantBinary evaluates binary operations on constants
func (cf *ConstantFolder) evaluateConstantBinary(left Value, operator Token, right Value) (Value, bool) {
	switch operator {
	case PLUS:
		return cf.evaluateAddition(left, right)
	case MINUS:
		return cf.evaluateSubtraction(left, right)
	case TIMES:
		return cf.evaluateMultiplication(left, right)
	case DIVIDE:
		return cf.evaluateDivision(left, right)
	case MODULO:
		return cf.evaluateModulo(left, right)
	case EQUAL:
		return cf.evaluateEquality(left, right)
	case NOTEQUAL:
		result, ok := cf.evaluateEquality(left, right)
		if !ok {
			return nil, false
		}
		if boolResult, isBool := result.(bool); isBool {
			return !boolResult, true
		}
		return nil, false
	case LT:
		return cf.evaluateLessThan(left, right)
	case LTE:
		return cf.evaluateLessEqual(left, right)
	case GT:
		return cf.evaluateGreaterThan(left, right)
	case GTE:
		return cf.evaluateGreaterEqual(left, right)
	case AND:
		return cf.evaluateLogicalAnd(left, right)
	case OR:
		return cf.evaluateLogicalOr(left, right)
	default:
		return nil, false
	}
}

// evaluateConstantUnary evaluates unary operations on constants
func (cf *ConstantFolder) evaluateConstantUnary(operator Token, operand Value) (Value, bool) {
	switch operator {
	case MINUS:
		switch v := operand.(type) {
		case int:
			return -v, true
		case float64:
			return -v, true
		default:
			return nil, false
		}
	case NOT:
		switch v := operand.(type) {
		case bool:
			return !v, true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

// Helper methods for binary operations
func (cf *ConstantFolder) evaluateAddition(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l + r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) + r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l + r, true
		}
		if r, ok := right.(int); ok {
			return l + float64(r), true
		}
	case string:
		if r, ok := right.(string); ok {
			return l + r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateSubtraction(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l - r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) - r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l - r, true
		}
		if r, ok := right.(int); ok {
			return l - float64(r), true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateMultiplication(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l * r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) * r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l * r, true
		}
		if r, ok := right.(int); ok {
			return l * float64(r), true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateDivision(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			if r == 0 {
				return nil, false // Division by zero
			}
			return l / r, true
		}
		if r, ok := right.(float64); ok {
			if r == 0.0 {
				return nil, false // Division by zero
			}
			return float64(l) / r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			if r == 0.0 {
				return nil, false // Division by zero
			}
			return l / r, true
		}
		if r, ok := right.(int); ok {
			if r == 0 {
				return nil, false // Division by zero
			}
			return l / float64(r), true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateModulo(left, right Value) (Value, bool) {
	if l, ok := left.(int); ok {
		if r, ok := right.(int); ok {
			if r == 0 {
				return nil, false // Division by zero
			}
			return l % r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateEquality(left, right Value) (Value, bool) {
	return left == right, true
}

func (cf *ConstantFolder) evaluateLessThan(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l < r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) < r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l < r, true
		}
		if r, ok := right.(int); ok {
			return l < float64(r), true
		}
	case string:
		if r, ok := right.(string); ok {
			return l < r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateLessEqual(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l <= r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) <= r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l <= r, true
		}
		if r, ok := right.(int); ok {
			return l <= float64(r), true
		}
	case string:
		if r, ok := right.(string); ok {
			return l <= r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateGreaterThan(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l > r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) > r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l > r, true
		}
		if r, ok := right.(int); ok {
			return l > float64(r), true
		}
	case string:
		if r, ok := right.(string); ok {
			return l > r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateGreaterEqual(left, right Value) (Value, bool) {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l >= r, true
		}
		if r, ok := right.(float64); ok {
			return float64(l) >= r, true
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l >= r, true
		}
		if r, ok := right.(int); ok {
			return l >= float64(r), true
		}
	case string:
		if r, ok := right.(string); ok {
			return l >= r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateLogicalAnd(left, right Value) (Value, bool) {
	if l, ok := left.(bool); ok {
		if r, ok := right.(bool); ok {
			return l && r, true
		}
	}
	return nil, false
}

func (cf *ConstantFolder) evaluateLogicalOr(left, right Value) (Value, bool) {
	if l, ok := left.(bool); ok {
		if r, ok := right.(bool); ok {
			return l || r, true
		}
	}
	return nil, false
}

// CommonSubexpressionEliminator eliminates common subexpressions
type CommonSubexpressionEliminator struct {
	optimizer *ExpressionOptimizer
}

// NewCommonSubexpressionEliminator creates a new CSE
func NewCommonSubexpressionEliminator(optimizer *ExpressionOptimizer) *CommonSubexpressionEliminator {
	return &CommonSubexpressionEliminator{optimizer: optimizer}
}

// EliminateCommonSubexpressions attempts to eliminate common subexpressions
func (cse *CommonSubexpressionEliminator) EliminateCommonSubexpressions(expr Expression) (Value, bool) {
	// Generate expression signature
	signature := cse.generateExpressionSignature(expr)
	if signature == "" {
		return nil, false
	}
	
	// Check cache
	cse.optimizer.mutex.RLock()
	if cached, exists := cse.optimizer.subexprCache[signature]; exists {
		cse.optimizer.optimizationHits++
		cse.optimizer.mutex.RUnlock()
		return cached, true
	}
	cse.optimizer.mutex.RUnlock()
	
	return nil, false
}

// CacheSubexpression caches a subexpression result
func (cse *CommonSubexpressionEliminator) CacheSubexpression(expr Expression, result Value) {
	signature := cse.generateExpressionSignature(expr)
	if signature == "" {
		return
	}
	
	cse.optimizer.mutex.Lock()
	defer cse.optimizer.mutex.Unlock()
	
	// Check cache size limit
	if len(cse.optimizer.subexprCache) >= cse.optimizer.maxCacheSize {
		cse.clearHalfCache(cse.optimizer.subexprCache)
	}
	
	cse.optimizer.subexprCache[signature] = result
}

// generateExpressionSignature generates a signature for an expression
func (cse *CommonSubexpressionEliminator) generateExpressionSignature(expr Expression) string {
	switch e := expr.(type) {
	case *Binary:
		leftSig := cse.generateExpressionSignature(e.Left)
		rightSig := cse.generateExpressionSignature(e.Right)
		return fmt.Sprintf("binary:%s:%s:%s", leftSig, e.Operator, rightSig)
	case *Unary:
		operandSig := cse.generateExpressionSignature(e.Operand)
		return fmt.Sprintf("unary:%s:%s", e.Operator, operandSig)
	case *Variable:
		return fmt.Sprintf("var:%s", e.Name)
	case *Literal:
		return fmt.Sprintf("lit:%v", e.Value)
	case *Call:
		argSigs := make([]string, len(e.Arguments))
		for i, arg := range e.Arguments {
			argSigs[i] = cse.generateExpressionSignature(arg)
		}
		return fmt.Sprintf("call:%s:%v", e.Function, argSigs)
	default:
		return ""
	}
}

// clearHalfCache clears half of the cache entries
func (cf *ConstantFolder) clearHalfCache(cache map[string]Value) {
	count := 0
	target := len(cache) / 2
	for k := range cache {
		delete(cache, k)
		count++
		if count >= target {
			break
		}
	}
}

// clearHalfCache clears half of the cache entries
func (cse *CommonSubexpressionEliminator) clearHalfCache(cache map[string]Value) {
	count := 0
	target := len(cache) / 2
	for k := range cache {
		delete(cache, k)
		count++
		if count >= target {
			break
		}
	}
}

// GetOptimizationStats returns optimization statistics
func (eo *ExpressionOptimizer) GetOptimizationStats() (hits, misses int64, hitRatio float64) {
	eo.mutex.RLock()
	defer eo.mutex.RUnlock()
	
	hits = eo.optimizationHits
	misses = eo.optimizationMiss
	total := hits + misses
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}
	return
}

// ClearCaches clears all optimization caches
func (eo *ExpressionOptimizer) ClearCaches() {
	eo.mutex.Lock()
	defer eo.mutex.Unlock()
	
	eo.constantCache = make(map[string]Value)
	eo.subexprCache = make(map[string]Value)
	eo.optimizationHits = 0
	eo.optimizationMiss = 0
}