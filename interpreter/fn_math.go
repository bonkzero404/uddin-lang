package interpreter

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

// Helper function to convert Value to float64
func toFloat64(pos Position, v Value, funcName string) (float64, Value) {
	switch val := v.(type) {
	case int:
		return float64(val), Value(nil)
	case float64:
		return val, Value(nil)
	default:
		return 0, Value(typeError(pos, "%s() requires a number, got %s", funcName, typeName(v)))
	}
}

// Helper function to convert Value to int
func toInt(pos Position, v Value, funcName string) (int, Value) {
	switch val := v.(type) {
	case int:
		return val, Value(nil)
	case float64:
		return int(val), Value(nil)
	default:
		return 0, Value(typeError(pos, "%s() requires a number, got %s", funcName, typeName(v)))
	}
}

// ========================================
// Basic Math Operations
// ========================================

// absFunc implements the abs() built-in function
// Returns the absolute value of a number
func absFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "abs", args, 1); err != nil {
		return err
	}
	switch val := args[0].(type) {
	case int:
		if val < 0 {
			return Value(-val)
		}
		return Value(val)
	case float64:
		return Value(math.Abs(val))
	default:
		panic(typeError(pos, "abs() requires a number, got %s", typeName(args[0])))
	}
}

// maxFunc implements the max() built-in function
// Returns the maximum value from multiple arguments or from an array
func maxFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		panic(typeError(pos, "max() requires at least 1 argument"))
	}

	// If first argument is an array, find max within the array
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "max() cannot be applied to empty array"))
		}
		maxVal := (*arr)[0]
		for i := 1; i < len(*arr); i++ {
			if evalLess(pos, maxVal, (*arr)[i]).(bool) {
			maxVal = (*arr)[i]
			}
		}
		return maxVal
	}

	// Otherwise, find max among all arguments
	maxVal := args[0]
	for i := 1; i < len(args); i++ {
		if evalLess(pos, maxVal, args[i]).(bool) {
			maxVal = args[i]
		}
	}
	return maxVal
}

// minFunc implements the min() built-in function
// Returns the minimum value from multiple arguments or from an array
func minFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		panic(typeError(pos, "min() requires at least 1 argument"))
	}

	// If first argument is an array, find min within the array
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "min() cannot be applied to empty array"))
		}
		minVal := (*arr)[0]
		for i := 1; i < len(*arr); i++ {
			if evalLess(pos, (*arr)[i], minVal).(bool) {
			minVal = (*arr)[i]
			}
		}
		return minVal
	}

	// Otherwise, find min among all arguments
	minVal := args[0]
	for i := 1; i < len(args); i++ {
		if evalLess(pos, args[i], minVal).(bool) {
			minVal = args[i]
		}
	}
	return minVal
}

// powFunc implements the pow() built-in function
// Returns base raised to the power of exponent
func powFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "pow", args, 2); err != nil {
		return err
	}
	base, err1 := toFloat64(pos, args[0], "pow")
	if err1 != nil {
		return err1
	}
	exp, err2 := toFloat64(pos, args[1], "pow")
	if err2 != nil {
		return err2
	}
	result := math.Pow(base, exp)

	// Return int if result is a whole number and both inputs were ints
	if _, baseIsInt := args[0].(int); baseIsInt {
		if _, expIsInt := args[1].(int); expIsInt {
			if result == math.Trunc(result) {
				return Value(int(result))
			}
		}
	}
	return Value(result)
}

// sqrtFunc implements the sqrt() built-in function
// Returns the square root of a number
func sqrtFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "sqrt", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "sqrt")
	if err != nil {
		return err
	}
	if val < 0 {
		panic(valueError(pos, "sqrt() of negative number"))
	}
	return Value(math.Sqrt(val))
}

// cbrtFunc implements the cbrt() built-in function
// Returns the cube root of a number
func cbrtFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "cbrt", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "cbrt")
	if err != nil {
		return err
	}
	return Value(math.Cbrt(val))
}

// ========================================
// Rounding Functions
// ========================================

// roundFunc implements the round() built-in function
// Rounds to nearest integer or to specified decimal places
func roundFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 1 {
		val, err := toFloat64(pos, args[0], "round")
		if err != nil {
			return err
		}
		return Value(int(math.Round(val)))
	} else if len(args) == 2 {
		val, err1 := toFloat64(pos, args[0], "round")
		if err1 != nil {
			return err1
		}
		places, err2 := toInt(pos, args[1], "round")
		if err2 != nil {
			return err2
		}
		if places < 0 {
			panic(valueError(pos, "round() decimal places must not be negative"))
		}
		multiplier := math.Pow(10, float64(places))
		return Value(math.Round(val*multiplier) / multiplier)
	}
	panic(typeError(pos, "round() requires 1 or 2 arguments, got %d", len(args)))
}

// floorFunc implements the floor() built-in function
// Returns the largest integer less than or equal to the number
func floorFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "floor", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "floor")
	if err != nil {
		return err
	}
	return Value(int(math.Floor(val)))
}

// ceilFunc implements the ceil() built-in function
// Returns the smallest integer greater than or equal to the number
func ceilFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "ceil", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "ceil")
	if err != nil {
		return err
	}
	return Value(int(math.Ceil(val)))
}

// truncFunc implements the trunc() built-in function
// Returns the integer part of a number (truncates decimal part)
func truncFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "trunc", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "trunc")
	if err != nil {
		return err
	}
	return Value(int(math.Trunc(val)))
}

// ========================================
// Trigonometric Functions
// ========================================

// sinFunc implements the sin() built-in function
func sinFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "sin", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "sin")
	if err != nil {
		return err
	}
	return Value(math.Sin(val))
}

// cosFunc implements the cos() built-in function
func cosFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "cos", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "cos")
	if err != nil {
		return err
	}
	return Value(math.Cos(val))
}

// tanFunc implements the tan() built-in function
func tanFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tan", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "tan")
	if err != nil {
		return err
	}
	return Value(math.Tan(val))
}

// asinFunc implements the asin() built-in function
func asinFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "asin", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "asin")
	if err != nil {
		return err
	}
	if val < -1 || val > 1 {
		panic(valueError(pos, "asin() input must be between -1 and 1"))
	}
	return Value(math.Asin(val))
}

// acosFunc implements the acos() built-in function
func acosFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "acos", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "acos")
	if err != nil {
		return err
	}
	if val < -1 || val > 1 {
		panic(valueError(pos, "acos() input must be between -1 and 1"))
	}
	return Value(math.Acos(val))
}

// atanFunc implements the atan() built-in function
func atanFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "atan", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "atan")
	if err != nil {
		return err
	}
	return Value(math.Atan(val))
}

// atan2Func implements the atan2() built-in function
func atan2Func(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "atan2", args, 2); err != nil {
		return err
	}
	y, err1 := toFloat64(pos, args[0], "atan2")
	if err1 != nil {
		return err1
	}
	x, err2 := toFloat64(pos, args[1], "atan2")
	if err2 != nil {
		return err2
	}
	return Value(math.Atan2(y, x))
}

// ========================================
// Hyperbolic Functions
// ========================================

// sinhFunc implements the sinh() built-in function
func sinhFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "sinh", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "sinh")
	if err != nil {
		return err
	}
	return Value(math.Sinh(val))
}

// coshFunc implements the cosh() built-in function
func coshFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "cosh", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "cosh")
	if err != nil {
		return err
	}
	return Value(math.Cosh(val))
}

// tanhFunc implements the tanh() built-in function
func tanhFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tanh", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "tanh")
	if err != nil {
		return err
	}
	return Value(math.Tanh(val))
}

// ========================================
// Logarithmic Functions
// ========================================

// logFunc implements the log() built-in function (natural logarithm)
func logFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "log", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "log")
	if err != nil {
		return err
	}
	if val <= 0 {
		panic(valueError(pos, "log() of non-positive number"))
	}
	return Value(math.Log(val))
}

// log10Func implements the log10() built-in function
func log10Func(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "log10", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "log10")
	if err != nil {
		return err
	}
	if val <= 0 {
		panic(valueError(pos, "log10() of non-positive number"))
	}
	return Value(math.Log10(val))
}

// log2Func implements the log2() built-in function
func log2Func(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "log2", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "log2")
	if err != nil {
		return err
	}
	if val <= 0 {
		panic(valueError(pos, "log2() of non-positive number"))
	}
	return Value(math.Log2(val))
}

// logbFunc implements the logb() built-in function (logarithm with custom base)
func logbFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "logb", args, 2); err != nil {
		return err
	}
	val, err1 := toFloat64(pos, args[0], "logb")
	if err1 != nil {
		return err1
	}
	base, err2 := toFloat64(pos, args[1], "logb")
	if err2 != nil {
		return err2
	}
	if val <= 0 {
		panic(valueError(pos, "logb() value must be positive"))
	}
	if base <= 0 || base == 1 {
		panic(valueError(pos, "logb() base must be positive and not equal to 1"))
	}
	return Value(math.Log(val) / math.Log(base))
}

// expFunc implements the exp() built-in function (e^x)
func expFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "exp", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "exp")
	if err != nil {
		return err
	}
	return Value(math.Exp(val))
}

// exp2Func implements the exp2() built-in function (2^x)
func exp2Func(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "exp2", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "exp2")
	if err != nil {
		return err
	}
	return Value(math.Exp2(val))
}

// ========================================
// Statistical Functions
// ========================================

// sumFunc implements the sum() built-in function
func sumFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "sum", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		var total float64
		var hasFloat bool

		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
				hasFloat = true
			default:
				panic(typeError(pos, "sum() array must contain only numbers"))
			}
		}

		if hasFloat {
			return Value(total)
		}
		return Value(int(total))
	}
	panic(typeError(pos, "sum() requires an array"))
}

// meanFunc implements the mean() built-in function
func meanFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "mean", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "mean() of empty array"))
		}

		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "mean() array must contain only numbers"))
			}
		}

		return Value(total / float64(len(*arr)))
	}
	panic(typeError(pos, "mean() requires an array"))
}

// medianFunc implements the median() built-in function
func medianFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "median", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "median() of empty array"))
		}

		// Convert to float slice and sort
		nums := make([]float64, len(*arr))
		for i, v := range *arr {
			switch val := v.(type) {
			case int:
				nums[i] = float64(val)
			case float64:
				nums[i] = val
			default:
				panic(typeError(pos, "median() array must contain only numbers"))
			}
		}

		sort.Float64s(nums)
		n := len(nums)

		if n%2 == 0 {
			// Even number of elements - return average of middle two
			return Value((nums[n/2-1] + nums[n/2]) / 2)
		} else {
			// Odd number of elements - return middle element
			return Value(nums[n/2])
		}
	}
	panic(typeError(pos, "median() requires an array"))
}

// modeFunc implements the mode() built-in function
func modeFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "mode", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "mode() of empty array"))
		}

		frequency := make(map[string]int)
		valueMap := make(map[string]Value)

		// Count frequencies
		for _, v := range *arr {
			key := toString(v, true)
			frequency[key]++
			valueMap[key] = v
		}

		// Find the most frequent value
		maxCount := 0
		var modeKey string
		for key, count := range frequency {
			if count > maxCount {
				maxCount = count
				modeKey = key
			}
		}

		return valueMap[modeKey]
	}
	panic(typeError(pos, "mode() requires an array"))
}

// stdDevFunc implements the std_dev() built-in function
func stdDevFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "std_dev", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) <= 1 {
			return Value(0.0)
		}

		// Calculate mean
		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "std_dev() array must contain only numbers"))
			}
		}
		mean := total / float64(len(*arr))

		// Calculate variance
		var sumSquaredDiff float64
		for _, v := range *arr {
			val, err := toFloat64(pos, v, "std_dev")
			if err != nil {
				return err
			}
			diff := val - mean
			sumSquaredDiff += diff * diff
		}
		variance := sumSquaredDiff / float64(len(*arr)-1)

		return Value(math.Sqrt(variance))
	}
	panic(typeError(pos, "std_dev() requires an array"))
}

// varianceFunc implements the variance() built-in function
func varianceFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "variance", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) <= 1 {
			return Value(0.0)
		}

		// Calculate mean
		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "variance() array must contain only numbers"))
			}
		}
		mean := total / float64(len(*arr))

		// Calculate variance
		var sumSquaredDiff float64
		for _, v := range *arr {
			val, err := toFloat64(pos, v, "variance")
			if err != nil {
				return err
			}
			diff := val - mean
			sumSquaredDiff += diff * diff
		}

		return Value(sumSquaredDiff / float64(len(*arr)-1))
	}
	panic(typeError(pos, "variance() requires an array"))
}

// ========================================
// Number Theory Functions
// ========================================

// gcdFunc implements the gcd() built-in function
func gcdFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "gcd", args, 2); err != nil {
		return err
	}
	a, err := toInt(pos, args[0], "gcd")
	if err != nil {
		return err
	}
	b, err := toInt(pos, args[1], "gcd")
	if err != nil {
		return err
	}

	a = int(math.Abs(float64(a)))
	b = int(math.Abs(float64(b)))

	for b != 0 {
		a, b = b, a%b
	}
	return Value(a)
}

// lcmFunc implements the lcm() built-in function
func lcmFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "lcm", args, 2); err != nil {
		return err
	}
	a, err := toInt(pos, args[0], "lcm")
	if err != nil {
		return err
	}
	b, err := toInt(pos, args[1], "lcm")
	if err != nil {
		return err
	}

	if a == 0 || b == 0 {
		return Value(0)
	}

	// Calculate GCD first
	gcdArgs := []Value{Value(a), Value(b)}
	gcdResult := gcdFunc(interp, pos, gcdArgs)
	gcd := gcdResult.(int)

	return Value(int(math.Abs(float64(a*b))) / gcd)
}

// factorialFunc implements the factorial() built-in function
func factorialFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "factorial", args, 1); err != nil {
		return err
	}
	n, err := toInt(pos, args[0], "factorial")
	if err != nil {
		return err
	}

	if n < 0 {
		panic(valueError(pos, "factorial() of negative number"))
	}
	if n > 20 {
		panic(valueError(pos, "factorial() argument too large (max 20)"))
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return Value(result)
}

// fibonacciFunc implements the fibonacci() built-in function
func fibonacciFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "fibonacci", args, 1); err != nil {
		return err
	}
	n, err := toInt(pos, args[0], "fibonacci")
	if err != nil {
		return err
	}

	if n < 0 {
		panic(valueError(pos, "fibonacci() of negative number"))
	}
	if n > 92 {
		panic(valueError(pos, "fibonacci() argument too large (max 92)"))
	}

	// Check memoization cache
	// EXPERIMENTAL: Using experimental memoization for fibonacci function
	memoKey := getMemoKey("fibonacci", args)
	if cached, exists := interp.getMemoValue(memoKey); exists {
		return cached
	}

	var result Value
	if n <= 1 {
		result = Value(n)
	} else {
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		result = Value(b)
	}

	// Store in memoization cache
	// EXPERIMENTAL: Storing result in experimental memoization cache
	interp.setMemoValue(memoKey, result)
	return result
}

// isPrimeFunc implements the is_prime() built-in function
func isPrimeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "is_prime", args, 1)
	n, err := toInt(pos, args[0], "is_prime")
	if err != nil {
		return err
	}

	if n <= 1 {
		return Value(false)
	}
	if n <= 3 {
		return Value(true)
	}
	if n%2 == 0 || n%3 == 0 {
		return Value(false)
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return Value(false)
		}
	}
	return Value(true)
}

// primeFactorsFunc implements the prime_factors() built-in function
func primeFactorsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "prime_factors", args, 1)
	n, err := toInt(pos, args[0], "prime_factors")
	if err != nil {
		return err
	}

	if n <= 1 {
		factors := make([]Value, 0)
		return Value(&factors)
	}

	factors := make([]Value, 0)

	// Handle factor of 2
	for n%2 == 0 {
		factors = append(factors, Value(2))
		n /= 2
	}

	// Handle odd factors
	for i := 3; i*i <= n; i += 2 {
		for n%i == 0 {
			factors = append(factors, Value(i))
			n /= i
		}
	}

	// If n is still > 1, then it's a prime
	if n > 1 {
		factors = append(factors, Value(n))
	}

	return Value(&factors)
}

// ========================================
// Random Number Functions
// ========================================

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// randomFunc implements the random() built-in function
func randomFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "random", args, 0)
	return Value(rng.Float64())
}

// randomIntFunc implements the random_int() built-in function
func randomIntFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "random_int", args, 2); err != nil {
		return err
	}
	min, err := toInt(pos, args[0], "random_int")
	if err != nil {
		return err
	}
	max, err := toInt(pos, args[1], "random_int")
	if err != nil {
		return err
	}

	if min >= max {
		panic(valueError(pos, "random_int() min must be less than max"))
	}

	return Value(rng.Intn(max-min) + min)
}

// randomFloatFunc implements the random_float() built-in function
func randomFloatFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "random_float", args, 2); err != nil {
		return err
	}
	min, err := toFloat64(pos, args[0], "random_float")
	if err != nil {
		return err
	}
	max, err := toFloat64(pos, args[1], "random_float")
	if err != nil {
		return err
	}

	if min >= max {
		panic(valueError(pos, "random_float() min must be less than max"))
	}

	return Value(rng.Float64()*(max-min) + min)
}

// randomChoiceFunc implements the random_choice() built-in function
func randomChoiceFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "random_choice", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "random_choice() of empty array"))
		}

		index := rng.Intn(len(*arr))
		return (*arr)[index]
	}
	panic(typeError(pos, "random_choice() requires an array"))
}

// shuffleFunc implements the shuffle() built-in function
func shuffleFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "shuffle", args, 1); err != nil {
		return err
	}
	if arr, ok := args[0].(*[]Value); ok {
		// Fisher-Yates shuffle
		for i := len(*arr) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
		}
		return Value(nil)
	}
	panic(typeError(pos, "shuffle() requires an array"))
}

// seedRandomFunc implements the seed_random() built-in function
func seedRandomFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "seed_random", args, 1); err != nil {
		return err
	}
	seed, err := toInt(pos, args[0], "seed_random")
	if err != nil {
		return err
	}
	rng = rand.New(rand.NewSource(int64(seed)))
	return Value(nil)
}

// ========================================
// Utility Functions
// ========================================

// signFunc implements the sign() built-in function
func signFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "sign", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "sign")
	if err != nil {
		return err
	}

	if val > 0 {
		return Value(1)
	} else if val < 0 {
		return Value(-1)
	}
	return Value(0)
}

// clampFunc implements the clamp() built-in function
func clampFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "clamp", args, 3); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "clamp")
	if err != nil {
		return err
	}
	min, err := toFloat64(pos, args[1], "clamp")
	if err != nil {
		return err
	}
	max, err := toFloat64(pos, args[2], "clamp")
	if err != nil {
		return err
	}

	if min > max {
		panic(valueError(pos, "clamp() min must be less than or equal to max"))
	}

	if val < min {
		val = min
	} else if val > max {
		val = max
	}

	// Return int if all inputs were ints
	if _, ok := args[0].(int); ok {
		if _, ok := args[1].(int); ok {
			if _, ok := args[2].(int); ok {
				return Value(int(val))
			}
		}
	}
	return Value(val)
}

// lerpFunc implements the lerp() built-in function (linear interpolation)
func lerpFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "lerp", args, 3)
	a, err := toFloat64(pos, args[0], "lerp")
	if err != nil {
		return err
	}
	b, err := toFloat64(pos, args[1], "lerp")
	if err != nil {
		return err
	}
	t, err := toFloat64(pos, args[2], "lerp")
	if err != nil {
		return err
	}

	return Value(a + t*(b-a))
}

// degreesFunc implements the degrees() built-in function
func degreesFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "degrees", args, 1); err != nil {
		return err
	}
	radians, err := toFloat64(pos, args[0], "degrees")
	if err != nil {
		return err
	}
	return Value(radians * 180.0 / math.Pi)
}

// radiansFunc implements the radians() built-in function
func radiansFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "radians", args, 1); err != nil {
		return err
	}
	degrees, err := toFloat64(pos, args[0], "radians")
	if err != nil {
		return err
	}
	return Value(degrees * math.Pi / 180.0)
}

// isNanFunc implements the is_nan() built-in function
func isNanFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "is_nan", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "is_nan")
	if err != nil {
		return err
	}
	return Value(math.IsNaN(val))
}

// isInfiniteFunc implements the is_infinite() built-in function
func isInfiniteFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "is_infinite", args, 1); err != nil {
		return err
	}
	val, err := toFloat64(pos, args[0], "is_infinite")
	if err != nil {
		return err
	}
	return Value(math.IsInf(val, 0))
}
