package main

import (
	"fmt"
	"log"
	"os"

	uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
	fmt.Println("=== Advanced UDDIN-LANG Library Usage ===")

	// Create a new engine
	engine := uddin.New()

	// Example 1: Setting custom variables
	fmt.Println("\n1. Setting Custom Variables:")
	engine.SetVariable("pi", 3.14159)
	engine.SetVariable("app_name", "My UDDIN App")
	engine.SetVariable("debug_mode", true)

	// Set multiple variables at once
	vars := map[string]any{
		"x": 10,
		"y": 20,
		"config": map[string]any{
			"timeout": 30,
			"retries": 3,
		},
	}
	engine.SetVariables(vars)

	// Use the variables
	stats, err := engine.ExecuteString(`
		print("App:", app_name)
		print("Pi value:", pi)
		print("Debug mode:", debug_mode)
		print("Coordinates:", x, y)
	`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Operations executed: %d\n", stats.Ops)

	// Example 2: File execution
	fmt.Println("\n2. File Execution:")
	// Create a temporary script file
	scriptContent := `
		fun greet(name):
			return "Hello, " + name + "!"
		end

		message = greet("World")
		print(message)

		// Math operations
		result = pi * 2
		print("2 * Pi =", result)
	`

	tempFile, err := os.CreateTemp("", "uddin_script_*.din")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(scriptContent)
	if err != nil {
		log.Fatal(err)
	}
	tempFile.Close()

	stats, err = engine.ExecuteFile(tempFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File execution - Operations: %d, User calls: %d, Builtin calls: %d\n",
		stats.Ops, stats.UserCalls, stats.BuiltinCalls)

	// Example 3: Expression parsing and evaluation
	fmt.Println("\n3. Expression Parsing and Evaluation:")
	expressions := []string{
		"2 + 3 * 4",
		"x * y + 5",
		"pi * 2",
	}

	for _, expr := range expressions {
		value, stats, err := engine.EvaluateString(expr)
		if err != nil {
			fmt.Printf("Error evaluating '%s': %v\n", expr, err)
			continue
		}
		fmt.Printf("%s = %v (ops: %d)\n", expr, value, stats.Ops)
	}

	// Example 4: Program parsing
	fmt.Println("\n4. Program Parsing:")
	programSource := `
		fun fibonacci(n):
			if (n <= 1) then:
				return n
			else:
				return fibonacci(n - 1) + fibonacci(n - 2)
			end
		end

		result = fibonacci(8)
		print("Fibonacci(8) =", result)
	`

	program, err := engine.ParseProgram([]byte(programSource))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Program parsed successfully. AST: %s\n", program.String())

	// Execute the parsed program
	stats, err = engine.ExecuteString(programSource)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fibonacci execution - Operations: %d, User calls: %d\n", stats.Ops, stats.UserCalls)

	// Example 5: Error handling
	fmt.Println("\n5. Error Handling:")
	errorCases := []string{
		"invalid_variable",
		"1 / 0",
		"undefined_function()",
		"x +", // Syntax error
	}

	for _, errorCase := range errorCases {
		_, err := engine.ExecuteString(errorCase)
		if err != nil {
			fmt.Printf("Expected error for '%s': %v\n", errorCase, err)
		} else {
			fmt.Printf("Unexpected success for '%s'\n", errorCase)
		}
	}

	// Example 6: Complex data structures
	fmt.Println("\n6. Complex Data Structures:")
	complexCode := `
		// Create a complex data structure
		employees = [
			{"name": "Alice", "age": 30, "department": "Engineering"},
			{"name": "Bob", "age": 25, "department": "Design"},
			{"name": "Charlie", "age": 35, "department": "Engineering"}
		]

		// Function to filter employees by department
		fun filter_by_department(employees, dept):
			result = []
			for (emp in employees):
				if (emp["department"] == dept) then:
					append(result, emp)
				end
			end
			return result
		end

		// Filter and display
		engineers = filter_by_department(employees, "Engineering")
		print("Engineers:")
		for (eng in engineers):
			print("  -", eng["name"], "(age", str(eng["age"]) + ")")
		end
	`

	stats, err = engine.ExecuteString(complexCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Complex data operations - Total operations: %d\n", stats.Ops+stats.UserCalls+stats.BuiltinCalls)

	fmt.Println("\n=== Advanced usage examples completed ===")
}
