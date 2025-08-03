---
sidebar_position: 1
---

# Installation

Get Uddin Programming Language up and running on your system.

## Prerequisites

Before installing Uddin-Lang, make sure you have:

- **Go 1.19 or later** (for building from source)
- **Git** (for cloning the repository)
- **Terminal/Command Prompt** access

## Installation Methods

### Method 1: Build from Source (Recommended)

This is currently the primary way to install Uddin-Lang:

```bash
# Clone the repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Build the interpreter
go build -o uddinlang main.go

# Make it executable (Linux/macOS)
chmod +x uddinlang
```

### Method 2: Direct Execution with Go

If you prefer not to build a binary:

```bash
# Clone the repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Run directly with Go
go run main.go your_script.din
```

## Verify Installation

Test your installation by checking the version:

```bash
# If you built the binary
./uddinlang --version

# If using go run
go run main.go --version
```

You should see output similar to:
```
Uddin-Lang v1.0
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

- **VS Code** - Use "Plain Text" or "Go" syntax highlighting
- **Vim/Neovim** - Use basic syntax highlighting
- **Sublime Text** - Use "Plain Text" mode
- **Atom** - Use "Plain Text" mode

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
# If you built the binary
./uddinlang test.din

# If using go run
go run main.go test.din
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
# Check syntax analysis
./uddinlang --analyze test.din

# Run with profiling
./uddinlang --profile test.din

# List available examples
./uddinlang --examples

# Show help
./uddinlang --help
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