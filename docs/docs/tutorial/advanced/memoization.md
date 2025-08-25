---
sidebar_position: 9
---

# Function Memoization (Experimental)

:::warning Experimental Feature
Memoization is an experimental feature and is not production-ready. Use with caution as it may have thread safety issues and memory management limitations.
:::

Memoization is a powerful optimization technique that allows functions to cache their results to improve performance for expensive computations with repeated calls. This is particularly useful for recursive algorithms, mathematical computations, or any function that performs expensive operations.

## What is Memoization?

Memoization automatically stores the results of function calls and returns the cached result when the same inputs occur again. This can dramatically improve performance for functions that:

- Perform expensive calculations
- Have overlapping subproblems (like recursive algorithms)
- Are called frequently with the same parameters
- Have deterministic outputs (pure functions)

## Basic Syntax

To enable memoization for a function, use the `memo` keyword before `fun`:

```uddin
memo fun function_name(parameters):
    // Function body
end
```

## Performance Measurement with time_now()

To measure the performance benefits of memoization, Uddin-Lang provides the `time_now()` built-in function that returns the current timestamp in milliseconds since Unix epoch. This allows you to accurately measure execution time:

```uddin
// Basic performance measurement
start_time = time_now()
// ... your code here ...
end_time = time_now()
execution_time = end_time - start_time
print("Execution time: " + str(execution_time) + "ms")
```

## Examples

### Fibonacci Sequence Optimization

One of the classic examples where memoization provides dramatic performance improvements:

```uddin
// Function without memoization (exponential time complexity)
fun fibonacci_slow(n):
    if (n <= 1) then:
        return n
    end
    return fibonacci_slow(n - 1) + fibonacci_slow(n - 2)
end

// Function with memoization (linear time complexity)
memo fun fibonacci_fast(n):
    if (n <= 1) then:
        return n
    end
    return fibonacci_fast(n - 1) + fibonacci_fast(n - 2)
end

// Performance comparison
fun test_fibonacci_performance():
    print("Testing Fibonacci performance...")
    
    // Test with n=35
    start_time = time_now()
    result1 = fibonacci_slow(35)
    slow_time = time_now() - start_time
    
    start_time = time_now()
    result2 = fibonacci_fast(35)
    fast_time = time_now() - start_time
    
    print("Without memoization: " + str(slow_time) + "ms")
    print("With memoization: " + str(fast_time) + "ms")
    print("Performance improvement: " + str(round(slow_time / fast_time, 2)) + "x faster")
end
```

### Prime Factorization

```uddin
memo fun calculate_prime_factors(n):
    factors = []
    divisor = 2
    
    while (divisor * divisor <= n):
        while (n % divisor == 0):
            append(factors, divisor)
            n = n / divisor
        end
        divisor = divisor + 1
    end
    
    if (n > 1) then:
        append(factors, n)
    end
    
    return factors
end

// Usage
factors1 = calculate_prime_factors(100)  // Calculated and cached
factors2 = calculate_prime_factors(100)  // Retrieved from cache
```

### Complex Data Processing

```uddin
memo fun process_complex_data(data_set, algorithm_type):
    processed_data = []
    
    for (item in data_set):
        if (algorithm_type == "advanced") then:
            // Expensive computation
            result = perform_advanced_analysis(item)
        else if (algorithm_type == "standard") then:
            result = perform_standard_analysis(item)
        else:
            result = perform_basic_analysis(item)
        end
        append(processed_data, result)
    end
    
    return processed_data
end
```

## Best Practices

### 1. Use Memoization for Pure Functions

Memoization works best with pure functions that have no side effects:

```uddin
// ✓ Good: Pure function suitable for memoization
memo fun calculate_hash(input_string):
    hash_value = 0
    for (i in range(len(input_string))):
        char_code = ord(input_string[i])
        hash_value = (hash_value * 31 + char_code) % 1000000007
    end
    return hash_value
end

// ✗ Avoid: Function with side effects
// memo fun log_and_calculate(value):  // Don't memoize this!
//     print("Calculating for: " + str(value))  // Side effect
//     return value * value
// end
```

### 2. Separate Side Effects from Pure Computation

If you need both logging and memoization, separate them:

```uddin
// Separate the side effect from the pure computation
fun log_and_calculate(value):
    print("Calculating for: " + str(value))
    return calculate_square_memoized(value)
end

memo fun calculate_square_memoized(value):
    return value * value  // Pure function suitable for memoization
end
```

### 3. Consider Input Complexity

Memoization works best with simple, hashable input parameters:

```uddin
// ✓ Good: Simple parameters
memo fun calculate_distance(x1, y1, x2, y2):
    return sqrt((x2 - x1) ** 2 + (y2 - y1) ** 2)
end

// ✓ Good: String parameters
memo fun parse_expression(expression):
    // Complex parsing logic
    return parsed_result
end
```

## Built-in Functions for Performance Testing

### time_now() Function

The `time_now()` function is a built-in function specifically designed for performance measurement:

- **Returns**: Current Unix timestamp in milliseconds (integer)
- **Parameters**: None
- **Usage**: Ideal for measuring execution time of code blocks
- **Precision**: Millisecond-level accuracy

```uddin
// Example: Measuring function execution time
fun measure_execution_time(func_name, func_call):
    start = time_now()
    result = func_call
    duration = time_now() - start
    print(func_name + " took " + str(duration) + "ms")
    return result
end

// Usage
result = measure_execution_time("fibonacci_fast(30)", fibonacci_fast(30))
```

## Performance Considerations

### Memory Usage

Memoized functions consume memory to store cached results. Consider this when:

- Functions are called with many different parameter combinations
- Function results are large objects
- Applications run for extended periods

### Cache Behavior

Currently, the memoization cache:

- Stores results for the lifetime of the program
- Uses function name and parameters as cache keys
- Does not have automatic cache size limits

### When to Use Memoization

Memoization is most beneficial when:

- **High computation cost**: The function performs expensive operations
- **Repeated calls**: The function is called multiple times with the same parameters
- **Overlapping subproblems**: Recursive functions that solve the same subproblems
- **Deterministic results**: Pure functions that always return the same output for the same input

### When NOT to Use Memoization

Avoid memoization when:

- **Functions have side effects**: Logging, file I/O, network calls, etc.
- **Results change over time**: Functions that depend on external state
- **Memory constraints**: Limited memory environments
- **Unique parameters**: Functions rarely called with the same parameters

## Limitations and Warnings

:::caution Current Limitations
- **Thread Safety**: Memoization is not thread-safe in concurrent environments
- **Memory Management**: No automatic cache eviction or size limits
- **Debugging**: Cached results may make debugging more difficult
- **Testing**: May affect test isolation if not properly managed
:::

## Example: Complete Performance Test

```uddin
// Complete example demonstrating memoization benefits
fun memoization_demo():
    print("=== Memoization Performance Demo ===")
    
    // Test 1: Fibonacci sequence
    print("\nTest 1: Fibonacci(35)")
    
    start_time = time_now()
    result1 = fibonacci_slow(35)
    time1 = time_now() - start_time
    
    start_time = time_now()
    result2 = fibonacci_fast(35)
    time2 = time_now() - start_time
    
    print("Without memoization: " + str(time1) + "ms")
    print("With memoization: " + str(time2) + "ms")
    print("Speedup: " + str(round(time1 / time2, 2)) + "x")
    
    // Test 2: Repeated calls
    print("\nTest 2: Repeated prime factorization")
    
    numbers = [97, 101, 103, 107, 109, 97, 101, 103]  // Some repeated
    
    start_time = time_now()
    for (num in numbers):
        factors = calculate_prime_factors(num)
    end
    total_time = time_now() - start_time
    
    print("Total time for " + str(len(numbers)) + " calculations: " + str(total_time) + "ms")
    print("Average time per calculation: " + str(round(total_time / len(numbers), 2)) + "ms")
end

// Run the demo
memoization_demo()
```

## Conclusion

Memoization is a powerful optimization technique that can provide significant performance improvements for appropriate use cases. However, as an experimental feature, it should be used with caution and thorough testing.

**Key takeaways:**
- Use memoization for expensive, pure functions
- Use `time_now()` function to measure and verify performance improvements
- Be aware of memory usage implications
- Test thoroughly before using in production
- Consider the experimental nature and potential limitations

For more examples, see the [memoization examples](../../examples/index.md#-performance-and-optimization) in the examples section.