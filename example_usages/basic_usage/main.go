package main

import (
	"fmt"
	"log"

	uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
	fmt.Println("=== Basic UDDIN-LANG Library Usage ===")

	// Create a new engine
	engine := uddin.New()

	// Example 1: Simple arithmetic
	fmt.Println("\n1. Simple Arithmetic:")
	value, stats, err := engine.EvaluateString("2 + 3 * 4")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %v\n", value)
	fmt.Printf("Operations executed: %d\n", stats.Ops)

	// Example 2: Variable assignment and usage
	fmt.Println("\n2. Variable Assignment:")
	stats, err = engine.ExecuteString("x = 10\ny = 20\nresult = x + y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Operations executed: %d\n", stats.Ops)

	// Example 3: Function definition and call
	fmt.Println("\n3. Function Definition and Call:")
	stats, err = engine.ExecuteString(`
		fun add(a, b):
			return a + b
		end

		result = add(15, 25)
		print("Sum:", result)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Example 4: Control flow
	fmt.Println("\n4. Control Flow:")
	stats, err = engine.ExecuteString(`
		number = 42
		if (number > 0) then:
			print("Number is positive")
		else:
			print("Number is not positive")
		end
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Example 5: Arrays and loops
	fmt.Println("\n5. Arrays and Loops:")
	stats, err = engine.ExecuteString(`
		numbers = [1, 2, 3, 4, 5]
		sum = 0

		for (num in numbers):
			sum = sum + num
		end

		print("Array sum:", sum)
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Basic usage completed ===")
}
