package interpreter

import (
	"sync"
)

// FastNumericEvaluator provides specialized evaluation for numeric operations
// to reduce interface{} boxing/unboxing and heap allocations
type FastNumericEvaluator struct {
	intPool   sync.Pool
	floatPool sync.Pool
}

// NewFastNumericEvaluator creates a new fast numeric evaluator
func NewFastNumericEvaluator() *FastNumericEvaluator {
	return &FastNumericEvaluator{
		intPool: sync.Pool{
			New: func() any {
				v := int(0)
				return &v
			},
		},
		floatPool: sync.Pool{
			New: func() any {
				v := float64(0)
				return &v
			},
		},
	}
}

// Global instance for reuse
var globalFastEvaluator = NewFastNumericEvaluator()

// FastEvalPlus provides optimized addition without interface{} boxing for common cases
func (fne *FastNumericEvaluator) FastEvalPlus(l, r Value) (Value, bool) {
	// Fast path for int + int
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			// Direct int arithmetic - no heap allocation
			return Value(li + ri), true
		}
		if rf, ok := r.(float64); ok {
			// int + float conversion - minimal allocation
			return Value(float64(li) + rf), true
		}
	}

	// Fast path for float + float
	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			// Direct float arithmetic - no heap allocation
			return Value(lf + rf), true
		}
		if ri, ok := r.(int); ok {
			// float + int conversion - minimal allocation
			return Value(lf + float64(ri)), true
		}
	}

	// Not a numeric operation, fallback to regular evaluation
	return nil, false
}

// FastEvalMinus provides optimized subtraction
func (fne *FastNumericEvaluator) FastEvalMinus(l, r Value) (Value, bool) {
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return Value(li - ri), true
		}
		if rf, ok := r.(float64); ok {
			return Value(float64(li) - rf), true
		}
	}

	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf - rf), true
		}
		if ri, ok := r.(int); ok {
			return Value(lf - float64(ri)), true
		}
	}

	return nil, false
}

// FastEvalTimes provides optimized multiplication
func (fne *FastNumericEvaluator) FastEvalTimes(l, r Value) (Value, bool) {
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return Value(li * ri), true
		}
		if rf, ok := r.(float64); ok {
			return Value(float64(li) * rf), true
		}
	}

	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf * rf), true
		}
		if ri, ok := r.(int); ok {
			return Value(lf * float64(ri)), true
		}
	}

	return nil, false
}

// FastEvalDivide provides optimized division
func (fne *FastNumericEvaluator) FastEvalDivide(l, r Value) (Value, bool) {
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			if ri == 0 {
				return nil, false // Let regular evaluator handle division by zero
			}
			return Value(li / ri), true
		}
		if rf, ok := r.(float64); ok {
			if rf == 0 {
				return nil, false
			}
			return Value(float64(li) / rf), true
		}
	}

	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			if rf == 0 {
				return nil, false
			}
			return Value(lf / rf), true
		}
		if ri, ok := r.(int); ok {
			if ri == 0 {
				return nil, false
			}
			return Value(lf / float64(ri)), true
		}
	}

	return nil, false
}

// FastEvalEqual provides optimized equality comparison
func (fne *FastNumericEvaluator) FastEvalEqual(l, r Value) (Value, bool) {
	// Fast path for same type comparisons
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return Value(li == ri), true
		}
		if rf, ok := r.(float64); ok {
			return Value(float64(li) == rf), true
		}
	}

	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf == rf), true
		}
		if ri, ok := r.(int); ok {
			return Value(lf == float64(ri)), true
		}
	}

	// Fast path for string comparison
	if ls, ok := l.(string); ok {
		if rs, ok := r.(string); ok {
			return Value(ls == rs), true
		}
	}

	// Fast path for bool comparison
	if lb, ok := l.(bool); ok {
		if rb, ok := r.(bool); ok {
			return Value(lb == rb), true
		}
	}

	return nil, false
}

// FastEvalLess provides optimized less-than comparison
func (fne *FastNumericEvaluator) FastEvalLess(l, r Value) (Value, bool) {
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return Value(li < ri), true
		}
		if rf, ok := r.(float64); ok {
			return Value(float64(li) < rf), true
		}
	}

	if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf < rf), true
		}
		if ri, ok := r.(int); ok {
			return Value(lf < float64(ri)), true
		}
	}

	// Fast path for string comparison
	if ls, ok := l.(string); ok {
		if rs, ok := r.(string); ok {
			return Value(ls < rs), true
		}
	}

	return nil, false
}

// GetFastEvaluator returns the global fast evaluator instance
func GetFastEvaluator() *FastNumericEvaluator {
	return globalFastEvaluator
}
