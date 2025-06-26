package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"uddin-lang/interpreter"
)

// CLI represents the command line interface
type CLI struct {
	args []string
}

// NewCLI creates a new CLI instance
func NewCLI(args []string) *CLI {
	return &CLI{args: args}
}

// Run executes the CLI with the given arguments
func (c *CLI) Run() error {
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
	fmt.Println("  uddinlang --examples       - List available example files")
	fmt.Println("  uddinlang --version        - Show version information")
	fmt.Println("  uddinlang --help           - Show this help message")
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
	// Validate file extension
	if filepath.Ext(filename) != ".din" {
		return fmt.Errorf("file must have .din extension")
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Execute the program
	success, output := interpreter.RunProgram(string(content))

	if success {
		fmt.Print(output)
		return nil
	} else {
		return fmt.Errorf("execution failed:\n%s", output)
	}
}

func main() {
	cli := NewCLI(os.Args)

	if err := cli.Run(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
