# Versioning Guide

UDDIN-LANG uses [Semantic Versioning](https://semver.org/) for all releases.

## Version Format

Versions follow the format: `MAJOR.MINOR.PATCH`

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.0.1): Bug fixes, backward compatible

## Installing as a Library

### Install Latest Release

```bash
go get github.com/bonkzero404/uddin-lang@latest
```

### Install Specific Version

```bash
go get github.com/bonkzero404/uddin-lang@v1.0.0
```

### Install from Main Branch

```bash
go get github.com/bonkzero404/uddin-lang@main
```

### Update to Latest

```bash
go get -u github.com/bonkzero404/uddin-lang@latest
go mod tidy
```

## Using in Your Project

### Import the Package

```go
package main

import (
    "fmt"
    "github.com/bonkzero404/uddin-lang/interpreter"
    uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
    // Using Engine API
    engine := uddin.New()
    result, err := engine.ExecuteString(`
        fun main():
            print("Hello, World!")
        end
    `)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
    
    // Using Interpreter directly
    code := `print("Hello from UDDIN-LANG!")`
    program, err := interpreter.ParseProgram([]byte(code))
    if err != nil {
        fmt.Println("Parse error:", err)
        return
    }
    
    config := &interpreter.Config{
        Vars: make(map[string]interpreter.Value),
    }
    stats, err := interpreter.Execute(program, config)
    if err != nil {
        fmt.Println("Execution error:", err)
        return
    }
    fmt.Println("Execution stats:", stats)
}
```

### go.mod Example

```go
module my-project

go 1.24

require (
    github.com/bonkzero404/uddin-lang v1.0.0
)
```

## Checking Installed Version

### From Go Module

```bash
go list -m github.com/bonkzero404/uddin-lang
```

### From Binary

```bash
# If installed via go install
uddinlang --version

# Output example:
# Uddin-Lang v1.0.0 (built: 2025-12-18_10:30:00_UTC) (commit: abc1234) (go: go1.25.4)
```

## Building from Source

### Build with Version Info

```bash
# Using Makefile (recommended)
make build

# Or manually
go build -ldflags "-X github.com/bonkzero404/uddin-lang/internal/version.Version=v1.0.0" \
         -o uddinlang ./cmd/uddin-lang
```

### Install from Source

```bash
# Install to GOPATH/bin
make install

# Or directly
go install github.com/bonkzero404/uddin-lang/cmd/uddin-lang@latest
```

## Version Information in Code

Version information is available programmatically:

```go
import "github.com/bonkzero404/uddin-lang/internal/version"

func main() {
    fmt.Println(version.GetVersion())
    // Output: Uddin-Lang v1.0.0 (built: ...) (commit: ...) (go: ...)
    
    fmt.Println(version.GetVersionShort())
    // Output: v1.0.0
}
```

## Creating a Release (For Maintainers)

See [RELEASE.md](../RELEASE.md) for detailed release instructions.

Quick version:

```bash
# Create and tag release
./scripts/release.sh v1.0.0

# Push tag to GitHub
git push origin v1.0.0
```

## Version Compatibility

### Go Version Requirements

- Minimum Go version: **1.24**
- Tested with Go: **1.24+**

### Module Compatibility

UDDIN-LANG follows Go module versioning:
- `v0.x.x`: Development/alpha releases (may have breaking changes)
- `v1.x.x`: Stable releases (backward compatible within major version)
- `v2.x.x`: Future major versions (if breaking changes are needed)

## Troubleshooting

### Module Not Found

If you get `module not found` errors:

```bash
# Clear module cache
go clean -modcache

# Re-download
go get github.com/bonkzero404/uddin-lang@latest
```

### Version Mismatch

If the installed version doesn't match:

```bash
# Update to specific version
go get github.com/bonkzero404/uddin-lang@v1.0.0
go mod tidy
```

### Build Errors

If you encounter build errors:

1. Ensure Go version is 1.24+
2. Run `go mod tidy`
3. Clear cache: `go clean -modcache`
4. Re-download: `go get -u github.com/bonkzero404/uddin-lang@latest`

## Related Documentation

- [RELEASE.md](../RELEASE.md) - Release process
- [README.md](../README.md) - Project overview
- [Getting Started Guide](docs/getting-started/installation.md) - Installation guide

