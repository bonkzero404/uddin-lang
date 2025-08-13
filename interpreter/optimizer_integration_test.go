package interpreter

import (
	"testing"
)

func TestConstantFoldingIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Value
	}{
		{"arithmetic", "2 + 3 * 4", Value(14)},
		{"string concat", `"hello" + " " + "world"`, Value("hello world")},
		{"boolean logic", "true and false", Value(false)},
		{"comparison", "5 > 3", Value(true)},
		{"unary", "-42", Value(-42)},
		{"nested", "(2 + 3) * (4 - 1)", Value(15)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse expression
			expr, err := ParseExpression([]byte(tt.input))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Create environment and evaluator
			config := TestConfig()
			env := NewEnvironment(config)
			stats := NewStats()
			interp := newInterpreter(config)
			evaluator := NewEvaluator(env, stats, interp)

			// Evaluate expression
			result := evaluator.EvaluateExpression(expr)

			// Check result
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestOptimizerCacheHit(t *testing.T) {
	// Parse expression
	expr, err := ParseExpression([]byte("2 + 3"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create environment and evaluator
	config := TestConfig()
	env := NewEnvironment(config)
	stats := NewStats()
	interp := newInterpreter(config)
	evaluator := NewEvaluator(env, stats, interp)

	// Get initial cache stats
	optimizer := GetGlobalExpressionOptimizer()
	initialHits, _, _ := optimizer.GetOptimizationStats()

	// Evaluate same expression multiple times
	for i := 0; i < 3; i++ {
		evaluator.EvaluateExpression(expr)
	}

	// Check cache hits increased
	finalHits, _, _ := optimizer.GetOptimizationStats()
	if finalHits <= initialHits {
		t.Errorf("Expected cache hits to increase, initial: %d, final: %d", initialHits, finalHits)
	}
}

func TestOptimizerDisabled(t *testing.T) {
	// Parse expression
	expr, err := ParseExpression([]byte("2 + 3"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create environment and evaluator with nil optimizer
	config := TestConfig()
	env := NewEnvironment(config)
	stats := NewStats()
	interp := newInterpreter(config)
	evaluator := &Evaluator{
		env:       env,
		stats:     stats,
		interp:    interp,
		optimizer: nil, // Disable optimizer
	}

	// Evaluate expression
	result := evaluator.EvaluateExpression(expr)

	// Check result is still correct
	expected := Value(5)
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}