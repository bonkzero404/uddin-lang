package main

import (
	"fmt"
	"log"

	uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
	fmt.Println("=== AST Conversion Examples ===")

	// Create a new engine
	engine := uddin.New()

	// Example 1: Simple code to JSON AST conversion
	fmt.Println("\n1. Simple Assignment to JSON AST:")
	simpleCode := "x = 42"
	jsonData, err := engine.ConvertStringToJSON(simpleCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original code: %s\n", simpleCode)
	fmt.Printf("JSON AST: %s\n", string(jsonData))

	// Convert back to source code
	convertedCode, err := engine.ConvertJSONToString(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Converted back: %s\n", convertedCode)

	// Example 2: Function definition conversion
	fmt.Println("\n2. Function Definition to JSON AST:")
	functionCode := `fun add(a, b):
    return a + b
end`
	jsonData, err = engine.ConvertStringToJSON(functionCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original function:\n%s\n", functionCode)
	fmt.Printf("JSON AST: %s\n", string(jsonData))

	convertedCode, err = engine.ConvertJSONToString(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Converted back:\n%s\n", convertedCode)

	// Example 3: Complex program with control flow
	fmt.Println("\n3. Complex Program to JSON AST:")
	complexCode := `x = 10
y = 20
if (x < y) then:
    print("x is smaller")
else:
    print("y is smaller")
end`
	jsonData, err = engine.ConvertStringToJSON(complexCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original program:\n%s\n", complexCode)
	fmt.Printf("JSON AST: %s\n", string(jsonData))

	convertedCode, err = engine.ConvertJSONToString(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Converted back:\n%s\n", convertedCode)

	// Example 4: Using convenience functions
	fmt.Println("\n4. Using Convenience Functions:")
	loopCode := `numbers = [1, 2, 3, 4, 5]
sum = 0
for (i in numbers):
    sum = sum + i
end
print("Sum:", sum)`

	// Using convenience functions (no need to create engine)
	jsonData, err = uddin.ConvertStringToJSON(loopCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original loop:\n%s\n", loopCode)
	fmt.Printf("JSON AST: %s\n", string(jsonData))

	convertedCode, err = uddin.ConvertJSONToString(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Converted back:\n%s\n", convertedCode)

	// Example 5: Round-trip conversion with execution
	fmt.Println("\n5. Round-trip Conversion with Execution:")
	executableCode := `fun factorial(n):
    if (n <= 1) then:
        return 1
    else:
        return n * factorial(n - 1)
    end
end

result = factorial(5)
print("Factorial of 5:", result)`

	// Convert to JSON and back
	jsonData, err = uddin.ConvertStringToJSON(executableCode)
	if err != nil {
		log.Fatal(err)
	}

	convertedCode, err = uddin.ConvertJSONToString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	// Execute both original and converted code
	fmt.Println("Executing original code:")
	_, err = engine.ExecuteString(executableCode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nExecuting converted code:")
	_, err = engine.ExecuteString(convertedCode)
	if err != nil {
		log.Fatal(err)
	}

	// Example 6: Byte array operations
	fmt.Println("\n6. Byte Array Operations:")
	byteCode := []byte("greeting = \"Hello, World!\"\nprint(greeting)")
	jsonBytes, err := uddin.ConvertToJSON(byteCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original bytes: %s\n", string(byteCode))
	fmt.Printf("JSON AST: %s\n", string(jsonBytes))

	restoredCode, err := uddin.ConvertFromJSON(jsonBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Restored code: %s\n", restoredCode)

	fmt.Println("\n=== AST Conversion examples completed ===")
}