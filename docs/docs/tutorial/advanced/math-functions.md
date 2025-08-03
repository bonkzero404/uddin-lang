---
sidebar_position: 2
---

# Mathematical Functions

Uddin-Lang provides a comprehensive mathematical library with over 50 mathematical functions, ranging from basic operations to advanced statistical and trigonometric functions. This extensive collection makes Uddin-Lang suitable for scientific computing, data analysis, and mathematical modeling.

## Basic Math Functions

Fundamental mathematical operations that form the foundation of numerical computing.

### Core Operations

```uddin
fun main():
    // abs - absolute value
    print(abs(-5))    // 5
    print(abs(3.14))  // 3.14

    // max/min - maximum/minimum values
    print(max(5, 10, 3, 8))  // 10
    print(min(5, 10, 3, 8))  // 3

    // pow - exponentiation
    print(pow(2, 3))   // 8 (2^3)
    print(pow(4, 0.5)) // 2 (square root of 4)

    // sqrt - square root
    print(sqrt(16))  // 4
    print(sqrt(2))   // 1.4142135623730951

    // cbrt - cube root
    print(cbrt(27))  // 3
    print(cbrt(8))   // 2
end
```

### Rounding Functions

Precise control over decimal precision and rounding behavior.

```uddin
fun main():
    // round - rounds to nearest integer
    print(round(3.14159))  // 3
    print(round(3.7))      // 4
    print(round(3.5))      // 4 (rounds half up)

    // floor - rounds down to nearest integer
    print(floor(3.9))   // 3
    print(floor(-2.1))  // -3

    // ceil - rounds up to nearest integer
    print(ceil(3.1))   // 4
    print(ceil(-2.9))  // -2

    // trunc - truncates decimal part
    print(trunc(3.9))   // 3
    print(trunc(-2.9))  // -2
end
```

## Trigonometric Functions

Comprehensive trigonometric functions for geometric calculations and wave analysis.

### Basic Trigonometric Functions

```uddin
fun main():
    // Angles in radians
    angle = 3.14159 / 4  // 45 degrees in radians

    print(sin(angle))  // 0.7071067811865476
    print(cos(angle))  // 0.7071067811865476
    print(tan(angle))  // 1.0000000000000002

    // Common angles
    print(sin(0))           // 0 (0 degrees)
    print(cos(3.14159))     // -1 (180 degrees)
    print(tan(3.14159/4))   // 1 (45 degrees)
end
```

### Inverse Trigonometric Functions

```uddin
fun main():
    // asin, acos, atan - inverse trigonometric functions
    print(asin(0.5))  // 0.5235987755982989 (30 degrees)
    print(acos(0.5))  // 1.0471975511965979 (60 degrees)
    print(atan(1))    // 0.7853981633974483 (45 degrees)

    // atan2 - two-argument arctangent
    print(atan2(1, 1))   // 0.7853981633974483 (45 degrees)
    print(atan2(-1, 1))  // -0.7853981633974483 (-45 degrees)
end
```

### Hyperbolic Functions

```uddin
fun main():
    // sinh, cosh, tanh - hyperbolic functions
    print(sinh(1))  // 1.1752011936438014
    print(cosh(1))  // 1.5430806348152437
    print(tanh(1))  // 0.7615941559557649

    // Hyperbolic identities
    x = 2.0
    print(cosh(x) * cosh(x) - sinh(x) * sinh(x))  // Should be 1
end
```

### Angle Conversion

```uddin
fun main():
    // degrees - radians to degrees
    print(degrees(3.14159))  // 180.0
    print(degrees(3.14159/2)) // 90.0

    // radians - degrees to radians
    print(radians(180))  // 3.141592653589793
    print(radians(90))   // 1.5707963267948966
end
```

## Logarithmic and Exponential Functions

Powerful functions for exponential growth, decay, and logarithmic calculations.

```uddin
fun main():
    // log - natural logarithm (ln)
    print(log(2.71828))  // 1.0 (approximately)
    print(log(exp(5)))   // 5.0 (log and exp are inverse)

    // log10 - base-10 logarithm
    print(log10(100))   // 2.0
    print(log10(1000))  // 3.0

    // log2 - base-2 logarithm
    print(log2(8))   // 3.0
    print(log2(16))  // 4.0

    // logb - custom base logarithm
    print(logb(125, 5))  // 3.0 (5^3 = 125)
    print(logb(81, 3))   // 4.0 (3^4 = 81)

    // exp - e raised to power x
    print(exp(1))  // 2.718281828459045 (Euler's number)
    print(exp(0))  // 1.0

    // exp2 - 2 raised to power x
    print(exp2(3))  // 8.0 (2^3)
    print(exp2(10)) // 1024.0 (2^10)
end
```

## Statistical Functions

Comprehensive statistical analysis functions for data science and analytics.

### Basic Statistics

```uddin
fun range_stat(data):
    return max(data) - min(data)
end

fun main():
    data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

    // sum - total sum
    print(sum(data))  // 55

    // mean - arithmetic average
    print(mean(data))  // 5.5

    // median - middle value
    print(median(data))  // 5.5

    // mode - most frequent value
    frequent_data = [1, 2, 2, 3, 3, 3, 4]
    print(mode(frequent_data))  // 3

    // range - difference between max and min
    print(range_stat(data))  // 9 (10 - 1)
end
```

### Advanced Statistics

```uddin
fun quartile(data, q):
    // Create a copy and sort the data
    sorted_data = data
    sort(sorted_data)
    n = len(sorted_data)

    if (q == 1) then:
        // Calculate Q1 (25th percentile)
        q1_pos = (n + 1) / 4
        q1_index = floor(q1_pos) - 1
        if (q1_index < 0) then:
            q1_index = 0
        end
        if (q1_index >= n) then:
            q1_index = n - 1
        end
        return sorted_data[q1_index]
    else if (q == 2) then:
        // Calculate Q2 (50th percentile - median)
        return median(sorted_data)
    else if (q == 3) then:
        // Calculate Q3 (75th percentile)
        q3_pos = 3 * (n + 1) / 4
        q3_index = floor(q3_pos) - 1
        if (q3_index < 0) then:
            q3_index = 0
        end
        if (q3_index >= n) then:
            q3_index = n - 1
        end
        return sorted_data[q3_index]
    end
end

fun main():
    data = [2, 4, 6, 8, 10]

    // variance - measure of data spread
    print(variance(data))  // 8.0

    // std_dev - standard deviation
    print(std_dev(data))  // 2.8284271247461903

    // coefficient of variation
    cv = std_dev(data) / mean(data)
    print(cv)  // 0.4714045207910317

    // quartiles for data distribution
    larger_data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    print(quartile(larger_data, 1))  // Q1: 3.0
    print(quartile(larger_data, 2))  // Q2 (median): 5.5
    print(quartile(larger_data, 3))  // Q3: 8.0

    // quartiles - Q1, Q2, Q3
    print("Q1: " + str(quartile(data, 1)))  // First quartile
    print("Q2: " + str(quartile(data, 2)))  // Second quartile (median)
    print("Q3: " + str(quartile(data, 3)))  // Third quartile
end
```

## Number Theory Functions

Mathematical functions for integer operations and number theory applications.

### Integer Operations

```uddin
fun main():
    // gcd - greatest common divisor
    print(gcd(48, 18))  // 6
    print(gcd(100, 25)) // 25

    // lcm - least common multiple
    print(lcm(12, 18))  // 36
    print(lcm(15, 20))  // 60

    // factorial - factorial calculation
    print(factorial(5))  // 120
    print(factorial(0))  // 1 (by definition)
    print(factorial(7))  // 5040

    // fibonacci - fibonacci sequence
    print(fibonacci(10))  // 55
    print(fibonacci(15))  // 610
end
```

### Prime Number Functions

```uddin
fun main():
    // is_prime - check if number is prime
    print(is_prime(17))  // true
    print(is_prime(15))  // false
    print(is_prime(2))   // true (smallest prime)

    // prime_factors - find prime factors
    print(prime_factors(60))   // [2, 2, 3, 5]
    print(prime_factors(100))  // [2, 2, 5, 5]

end
```

## Random Number Functions

Pseudo-random number generation for simulations, games, and statistical sampling.

```uddin
fun main():
    // random - random number between 0 and 1
    print(random())  // 0.6394267984578837

    // random_int - random integer within range
    print(random_int(1, 10))   // 7 (inclusive range)
    print(random_int(50, 100)) // 73

    // random_float - random float within range
    print(random_float(1.0, 5.0))   // 3.2847
    print(random_float(-10.0, 10.0)) // -4.567

    // random_choice - pick random element from array
    options = ["apple", "banana", "orange"]
    print(random_choice(options))  // "banana"

    colors = ["red", "green", "blue", "yellow"]
    print(random_choice(colors))   // "green"

    // shuffle - randomize array order
    numbers = [1, 2, 3, 4, 5]
    shuffled = shuffle(numbers)
    print(shuffled)  // [3, 1, 5, 2, 4]

    // seed_random - set seed for reproducibility
    seed_random(42)
    print(random())  // will always be same with same seed
    print(random())  // predictable sequence
end
```

## Utility Math Functions

General-purpose mathematical utilities for common programming tasks.

```uddin
fun main():
    // sign - get sign of number
    print(sign(5))   // 1
    print(sign(-3))  // -1
    print(sign(0))   // 0

    // clamp - constrain value within range
    print(clamp(15, 1, 10))  // 10 (clamped to max)
    print(clamp(-5, 1, 10))  // 1  (clamped to min)
    print(clamp(7, 1, 10))   // 7  (within range)

    // lerp - linear interpolation
    print(lerp(0, 10, 0.5))   // 5.0  (50% between 0 and 10)
    print(lerp(20, 30, 0.3))  // 23.0 (30% between 20 and 30)
    print(lerp(100, 200, 0.75)) // 175.0

    // is_nan - check for NaN (Not a Number)
    print(is_nan(5))    // false
    print(is_nan(0)) // true

    // is_infinite - check for infinity
    print(is_infinite(1/1))   // true
    print(is_infinite(100))   // false
end
```

## Practical Usage Examples

Real-world applications demonstrating the power of mathematical functions in Uddin-Lang.

### Statistical Calculator

```uddin
fun calculate_stats(data):
    if (len(data) == 0) then:
        return "Error: Empty data set"
    end

    count_val = len(data)
    sum_val = sum(data)
    mean_val = mean(data)
    median_val = median(data)
    min_val = min(data)
    max_val = max(data)
    variance_val = variance(data)
    std_dev_val = std_dev(data)

    print("Statistical Analysis:")
    print("Count: " + str(count_val))
    print("Sum: " + str(sum_val))
    print("Mean: " + str(mean_val))
    print("Median: " + str(median_val))
    print("Min: " + str(min_val))
    print("Max: " + str(max_val))
    print("Variance: " + str(variance_val))
    print("Std Dev: " + str(std_dev_val))
end

fun main():
    scores = [85, 92, 78, 96, 88, 91, 84, 89]
    calculate_stats(scores)
end
```

### Geometry Calculator

```uddin
fun circle_area(radius):
    return 3.14159 * pow(radius, 2)
end

fun sphere_volume(radius):
    return (4.0 / 3.0) * 3.14159 * pow(radius, 3)
end

fun distance_2d(x1, y1, x2, y2):
    return sqrt(pow(x2 - x1, 2) + pow(y2 - y1, 2))
end

fun triangle_area(base, height):
    return 0.5 * base * height
end

fun main():
    // Circle calculations
    radius = 5
    area = circle_area(radius)
    print("Circle area (r=5): " + str(area))

    // Sphere calculations
    sphere_r = 3
    volume = sphere_volume(sphere_r)
    print("Sphere volume (r=3): " + str(volume))

    // Distance calculation
    dist = distance_2d(0, 0, 3, 4)
    print("Distance from (0,0) to (3,4): " + str(dist))

    // Triangle area
    tri_area = triangle_area(10, 8)
    print("Triangle area (base=10, height=8): " + str(tri_area))
end
```

### Prime Number Generator

```uddin
fun generate_primes(limit):
    primes = []
    i = 2
    while (i <= limit):
        if (is_prime(i)) then:
            push(primes, i)
        end
        i = i + 1
    end
    return primes
end

fun analyze_prime_factors(n):
    factors = prime_factors(n)
    print("Prime factorization of " + str(n) + ":")

    // Count occurrences of each factor
    i = 0
    while (i < len(factors)):
        factor = factors[i]
        count = 1
        j = i + 1
        while (j < len(factors) and factors[j] == factor):
            count = count + 1
            j = j + 1
        end

        if (count == 1) then:
            print(str(factor))
        else:
            print(str(factor) + "^" + str(count))
        end

        i = j
    end
end

fun main():
    // Generate primes up to 50
    primes = generate_primes(50)
    print("Prime numbers up to 50:")
    print(str(primes))

    // Analyze prime factorization
    analyze_prime_factors(60)
    analyze_prime_factors(100)
    analyze_prime_factors(144)
end
```

## Best Practices

1. **Choose the Right Function**: Use appropriate mathematical functions for your specific use case
2. **Handle Edge Cases**: Always check for division by zero, negative inputs for sqrt, etc.
3. **Performance Considerations**: For large datasets, consider the computational complexity
4. **Precision**: Be aware of floating-point precision limitations
5. **Combine Functions**: Leverage multiple functions together for complex calculations

This comprehensive guide covers all essential mathematical functions available in Uddin-Lang, providing you with the tools needed for scientific computing, data analysis, and mathematical problem-solving.

### Monte Carlo Simulation

```uddin
fun estimate_pi(iterations):
    inside_circle = 0
    i = 0

    while i < iterations:
        x = random_float(-1, 1)
        y = random_float(-1, 1)

        if sqrt(x*x + y*y) <= 1:
            inside_circle = inside_circle + 1
        end
        i = i + 1
    end

    return 4.0 * inside_circle / iterations
end

fun main():
    // Estimate pi value using Monte Carlo method
    seed_random(42)  // for consistent results
    estimated_pi = estimate_pi(100000)
    actual_pi = 3.14159

    print("Estimated π: " + str(estimated_pi))
    print("Actual π: " + str(actual_pi))
    print("Error: " + str(abs(estimated_pi - actual_pi)))
end
```

## Advanced Tips and Best Practices

1. **Floating Point Precision**: Be careful with floating point comparisons, use tolerance values
2. **Function Domains**: Some functions have limited domains (e.g., sqrt for positive numbers)
3. **Performance**: For intensive calculations, consider caching results
4. **Random Seed**: Use `seed_random()` for reproducible results in testing
5. **Overflow Prevention**: Be cautious with functions like `factorial()` for large inputs
6. **Error Handling**: Always validate inputs before mathematical operations

```uddin
fun is_equal(a, b, tolerance):
    return abs(a - b) < tolerance
end

fun safe_sqrt(x):
    if (x < 0) then:
        print("Error: Cannot calculate square root of negative number")
        return 0
    end
    return sqrt(x)
end

fun main():
    // Safe floating point comparison
    result = sqrt(2) * sqrt(2)
    print(is_equal(result, 2.0, 0.0001))  // true

    // Safe square root calculation
    print(safe_sqrt(16))   // 4.0
    print(safe_sqrt(-4))   // Error message + 0
end
```

With this comprehensive mathematical library, Uddin-Lang provides powerful tools for scientific computing, data analysis, statistical modeling, and mathematical problem-solving across various domains.
