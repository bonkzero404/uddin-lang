package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

// CLI represents the command line interface
type CLI struct {
	args                       []string
	profile                    bool
	analyze                    bool
	toJson                     bool
	fromJson                   bool
	// DEPRECATED: Use memoryOptimizeStable or memoryOptimizeExperimental instead
	memoryOptimize             bool
	memoryOptimizeStable       bool
	memoryOptimizeExperimental bool
}

// NewCLI creates a new CLI instance
func NewCLI(args []string) *CLI {
	return &CLI{args: args, profile: false, analyze: false, toJson: false}
}

// Run executes the CLI with the given arguments
func (c *CLI) Run() error {
	if len(c.args) < 2 {
		c.printUsage()
		return nil
	}

	// Check for flags first
	for i := len(c.args) - 1; i >= 0; i-- {
		arg := c.args[i]
		switch arg {
		case "--profile", "-p":
			c.profile = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--analyze", "-a":
			c.analyze = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--to_json":
			c.toJson = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--from_json":
			c.fromJson = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--memory-optimize", "-m":
			// DEPRECATED: Use --memory-optimize-stable or --memory-optimize-experimental instead
			c.memoryOptimize = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--memory-optimize-stable":
			c.memoryOptimizeStable = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		case "--memory-optimize-experimental":
			c.memoryOptimizeExperimental = true
			// Remove the flag from args
			c.args = append(c.args[:i], c.args[i+1:]...)
		}
	}

	// Re-check args length after potential flag removal
	if len(c.args) < 2 {
		c.printUsage()
		return nil
	}

	arg := c.args[1]
	switch arg {
	case "--help", "-h":
		c.printUsage()
		return nil
	case "--version", "-v":
		c.printVersion()
		return nil
	case "--examples", "-e":
		return c.listExamples()
	default:
		return c.runScript(arg)
	}
}

func (c *CLI) printUsage() {
	fmt.Println("Uddin-Lang Interpreter")
	fmt.Println("Usage:")
	fmt.Println("  uddinlang <filename.din>   - Run a Uddin-Lang script file")
	fmt.Println("  uddinlang --profile <filename.din> - Run with performance profiling")
	fmt.Println("  uddinlang --analyze <filename.din> - Analyze syntax without execution")
	fmt.Println("  uddinlang --memory-optimize <filename.din> - Run with memory optimizations (DEPRECATED)")
	fmt.Println("  uddinlang --memory-optimize-stable <filename.din> - Run with stable memory optimizations")
	fmt.Println("  uddinlang --memory-optimize-experimental <filename.din> - Run with experimental memory optimizations")
	fmt.Println("  uddinlang --to_json <filename.din> - Convert Uddin-Lang code to JSON AST")
	fmt.Println("  uddinlang --from_json <filename.json> - Convert JSON AST back to Uddin-Lang code")
	fmt.Println("  uddinlang --examples       - List available example files")
	fmt.Println("  uddinlang --version        - Show version information")
	fmt.Println("  uddinlang --help           - Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --profile, -p                    - Enable performance profiling output")
	fmt.Println("  --analyze, -a                    - Analyze syntax only (no execution)")
	fmt.Println("  --memory-optimize, -m            - Enable experimental memory optimizations (DEPRECATED)")
	fmt.Println("  --memory-optimize-stable         - Enable stable memory optimizations (production-ready)")
	fmt.Println("  --memory-optimize-experimental   - Enable experimental memory optimizations (may be unstable)")
	fmt.Println("  --to_json                        - Convert source code to JSON AST representation")
}

func (c *CLI) printVersion() {
	fmt.Println(interpreter.GetVersionInfo())
}

func (c *CLI) listExamples() error {
	fmt.Println("Available example files:")
	examplesDir := "./examples"

	files, err := os.ReadDir(examplesDir)
	if err != nil {
		return fmt.Errorf("error reading examples directory: %w", err)
	}

	count := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".din" {
			fmt.Printf("  %s\n", file.Name())
			count++
		}
	}

	if count == 0 {
		fmt.Println("  No example files found")
	} else {
		fmt.Println("\nRun an example with: uddinlang examples/<filename>")
	}

	return nil
}

func (c *CLI) runScript(filename string) error {
	// Handle transpiler flags
	if c.toJson {
		return c.convertToJson(filename)
	}

	if c.fromJson {
		return c.convertFromJson(filename)
	}

	// Validate file extension for normal execution
	if filepath.Ext(filename) != ".din" {
		return fmt.Errorf("file must have .din extension")
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// If analyze flag is set, only perform syntax analysis
	if c.analyze {
		success, output := interpreter.AnalyzeSyntax(string(content))
		fmt.Print(output)
		if !success {
			return fmt.Errorf("syntax analysis failed")
		}
		return nil
	}

	// Create options based on flags
	options := &interpreter.RunProgramOptions{
		ShowProfiling:              c.profile,
		MemoryOptimize:             c.memoryOptimize, // DEPRECATED: kept for backward compatibility
		MemoryOptimizeStable:       c.memoryOptimizeStable,
		MemoryOptimizeExperimental: c.memoryOptimizeExperimental,
	}

	// Execute the program with options
	success, output := interpreter.RunProgramWithOptions(string(content), options)

	if success {
		fmt.Print(output)
		return nil
	} else {
		return fmt.Errorf("execution failed:\n%s", output)
	}
}

// convertToJson converts a .din file to JSON AST representation
func (c *CLI) convertToJson(filename string) error {
	// Validate file extension
	if filepath.Ext(filename) != ".din" {
		return fmt.Errorf("file must have .din extension for --to_json")
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Parse the source code to AST using the correct interface
	program, err := interpreter.ParseProgram(content)
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}

	// Convert AST to JSON
	jsonData, err := json.MarshalIndent(program, "", "  ")
	if err != nil {
		return fmt.Errorf("error converting to JSON: %w", err)
	}

	// Print JSON to stdout
	fmt.Println(string(jsonData))
	return nil
}

func (c *CLI) convertFromJson(filename string) error {
	// Validate file extension (allow stdin)
	if filename != "/dev/stdin" && filepath.Ext(filename) != ".json" {
		return fmt.Errorf("file must have .json extension for --from_json")
	}

	// Read the JSON file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Parse JSON into Program struct
	var program interpreter.Program
	if err := json.Unmarshal(content, &program); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	// Convert AST back to source code
	sourceCode := program.String()

	// Print the generated source code
	fmt.Println(sourceCode)
	return nil
}

func main() {
	cli := NewCLI(os.Args)
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}
