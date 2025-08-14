---
sidebar_position: 1
---

# Installation

Get Uddin Programming Language up and running on your system.

## Prerequisites

Before installing Uddin-Lang, make sure you have:

-   **Go 1.23 or later** (for building from source)
-   **Git** (for cloning the repository)
-   **Terminal/Command Prompt** access

## Installation Methods

### Method 1: Global Installation with Go Install (Recommended)

The easiest way to install Uddin-Lang globally:

```bash
# Clone the repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Install globally
go install ./cmd/uddin-lang

# Now you can use uddin-lang from anywhere
uddin-lang --version
```

### Method 2: Build from Source

If you prefer to build a local binary:

```bash
# Clone the repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Build the interpreter
go build -o uddinlang cmd/uddin-lang/main.go

# Make it executable (Linux/macOS)
chmod +x uddinlang
```

### Method 3: Direct Execution with Go

If you prefer not to build a binary:

```bash
# Clone the repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Run directly with Go
go run cmd/uddin-lang/main.go your_script.din
```

## Verify Installation

Test your installation by checking the version:

```bash
# If you used go install (global installation)
uddin-lang --version

# If you built the binary locally
./uddinlang --version

# If using go run
go run cmd/uddin-lang/main.go --version
```

You should see output similar to:

```
Uddin-Lang v1.0.0
```

## Setting Up Your Environment

### Adding to PATH (Optional)

To use `uddinlang` from anywhere, add it to your system PATH:

#### Linux/macOS

```bash
# Copy to a directory in your PATH
sudo cp uddinlang /usr/local/bin/

# Or add the current directory to PATH
echo 'export PATH="$PATH:$(pwd)"' >> ~/.bashrc
source ~/.bashrc
```

#### Windows

1. Copy `uddinlang.exe` to a directory like `C:\uddin-lang\`
2. Add `C:\uddin-lang\` to your system PATH environment variable
3. Restart your command prompt

### IDE/Editor Setup

While Uddin-Lang doesn't have dedicated IDE plugins yet, you can use any text editor:

#### Recommended Editors

-   **VS Code** - Use "Plain Text" or "Go" syntax highlighting
-   **Vim/Neovim** - Use basic syntax highlighting
-   **Sublime Text** - Use "Plain Text" mode
-   **Atom** - Use "Plain Text" mode

#### File Extension

Uddin-Lang files use the `.din` extension:

```bash
# Example file names
hello.din
my_script.din
app.din
```

## Testing Your Installation

Create a simple test file to verify everything works:

### Create `test.din`

```uddin
// test.din - Installation verification
fun main():
    print("🎉 Uddin-Lang is working!")
    print("Installation successful!")

    // Test basic functionality
    name = "Developer"
    print("Hello, " + name + "!")
end
```

### Run the test

```bash
# If you used go install (global installation)
uddin-lang test.din

# If you built the binary locally
./uddinlang test.din

# If using go run
go run cmd/uddin-lang/main.go test.din
```

### Expected Output

```
🎉 Uddin-Lang is working!
Installation successful!
Hello, Developer!
```

## CLI Tools Verification

Test the development tools:

```bash
# If you used go install (global installation)
uddin-lang --analyze test.din
uddin-lang --profile test.din
uddin-lang --examples
uddin-lang --help

# If you built the binary locally
./uddinlang --analyze test.din
./uddinlang --profile test.din
./uddinlang --examples
./uddinlang --help
```

## Using Go Run (Development Mode)

For development or when you don't want to build a binary, you can use `go run` directly:

### Basic Usage

```bash
# Run any .din file
go run cmd/uddin-lang/main.go examples/basics/hello_world.din

# Run with profiling
go run cmd/uddin-lang/main.go --profile examples/basics/functional_demo.din

# Analyze syntax without execution
go run cmd/uddin-lang/main.go --analyze examples/basics/hello_world.din

# Show help
go run cmd/uddin-lang/main.go --help
```

### Available Examples

Try running different example files:

```bash
# Basic examples
go run cmd/uddin-lang/main.go examples/basics/hello_world.din
go run cmd/uddin-lang/main.go examples/basics/functional_demo.din
go run cmd/uddin-lang/main.go examples/basics/loop_control_demo.din

# Data structures
go run cmd/uddin-lang/main.go examples/data-structures/array_methods_demo.din
go run cmd/uddin-lang/main.go examples/data-structures/linked_list_demo.din

# Math and calculations
go run cmd/uddin-lang/main.go examples/math-and-calculations/simple_calculator.din
go run cmd/uddin-lang/main.go examples/math-and-calculations/geometry_calculator.din
```

### Development Tips

- Use `go run` during development for faster iteration
- Build a binary (`go build`) for production or distribution
- The `go run` approach automatically uses the latest source code
- No need to rebuild when you make changes to the interpreter

## Using Go Install (Global Installation)

For a more convenient installation that makes `uddin-lang` available globally on your system, you can use `go install`:

### Install from Local Source

If you have cloned the repository:

```bash
# Navigate to the project directory
cd uddin-lang

# Install globally using go install
go install ./cmd/uddin-lang
```

After installation, you can use `uddin-lang` from anywhere:

```bash
# Check version
uddin-lang --version
# Output: Uddin-Lang v1.0.0

# Run any .din file from anywhere
uddin-lang /path/to/your/script.din

# Use all CLI features globally
uddin-lang --help
uddin-lang --analyze script.din
uddin-lang --profile script.din
```

### Install from Repository (Future)

*Note: Direct installation from GitHub repository is planned for future releases:*

```bash
# This will be available in future releases
go install github.com/bonkzero404/uddin-lang/cmd/uddin-lang@latest
```

### Verify Global Installation

To verify that `go install` worked correctly:

```bash
# Check if uddin-lang is in your PATH
which uddin-lang

# Test with version command
uddin-lang --version

# Test with a simple example
uddin-lang examples/basics/hello_world.din
```

### Benefits of Go Install

- **Global Access**: Use `uddin-lang` from any directory
- **No Path Setup**: Automatically installed to `$GOPATH/bin` or `$GOBIN`
- **Easy Updates**: Simply run `go install` again to update
- **Clean Installation**: No need to manage binary files manually

### Uninstalling

To remove the globally installed binary:

```bash
# Find the installation location
which uddin-lang

# Remove the binary (typically in $GOPATH/bin or $GOBIN)
rm $(which uddin-lang)
```

## Troubleshooting

### Common Issues

#### "Go not found" Error

**Problem**: `go: command not found`

**Solution**: Install Go from [golang.org](https://golang.org/dl/)

#### "Permission denied" Error

**Problem**: `./uddinlang: permission denied`

**Solution**: Make the file executable:

```bash
chmod +x uddinlang
```

#### Build Errors

**Problem**: Build fails with dependency errors

**Solution**: Ensure you have Go 1.19+ and run:

```bash
go mod tidy
go build -o uddinlang main.go
```

### Getting Help

If you encounter issues:

1. **Check the examples**: Run `./uddinlang --examples` to see working code
2. **GitHub Issues**: [Report bugs or ask questions](https://github.com/bonkzero404/uddin-lang/issues)
3. **Discussions**: [Community discussions](https://github.com/bonkzero404/uddin-lang/discussions)

## Development Setup

If you plan to contribute or modify Uddin-Lang:

```bash
# Clone with development intent
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Install development dependencies
go mod download

# Run tests
go test ./...

# Run specific tests
go test ./interpreter
```

## Next Steps

Now that Uddin-Lang is installed:

1. **[Quick Start Guide](quick-start.md)** - Learn the basics in 5 minutes
2. **[Your First Program](your-first-program.md)** - Write your first Uddin application
3. **[Tutorial Series](../tutorial/basics/introduction.md)** - Comprehensive learning path

---

**Next**: [Quick Start →](quick-start.md)
