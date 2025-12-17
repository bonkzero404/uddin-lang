# Release Guide

This document explains how to create releases and manage versioning for UDDIN-LANG.

## Versioning Scheme

UDDIN-LANG uses [Semantic Versioning](https://semver.org/) (SemVer):

-   **MAJOR** version (1.0.0) - Breaking changes
-   **MINOR** version (0.1.0) - New features, backward compatible
-   **PATCH** version (0.0.1) - Bug fixes, backward compatible

## Creating a Release

### 1. Update Version

Before creating a release, ensure all changes are committed:

```bash
git status
git add .
git commit -m "Prepare for release v1.0.0"
```

### 2. Create Git Tag

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Or use Makefile
make tag-release VERSION=v1.0.0
```

### 3. Push Tag to GitHub

```bash
git push origin v1.0.0
# Or push all tags
git push --tags
```

### 4. Build Release Binary

```bash
# Build with version info
make release VERSION=v1.0.0
```

### 5. Verify Version

```bash
./uddinlang --version
# Should output: Uddin-Lang v1.0.0 (built: ...) (commit: ...) (go: ...)
```

## Installing as a Library

Users can install UDDIN-LANG as a Go library using:

### Install Latest Version

```bash
go get github.com/bonkzero404/uddin-lang@latest
```

### Install Specific Version

```bash
go get github.com/bonkzero404/uddin-lang@v1.0.0
```

### Install Latest from Main Branch

```bash
go get github.com/bonkzero404/uddin-lang@main
```

## Using as a Library

After installation, import the package:

```go
import (
    "github.com/bonkzero404/uddin-lang/interpreter"
    uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
    // Use the engine API
    engine := uddin.New()
    result, err := engine.ExecuteString(`
        fun main():
            print("Hello from UDDIN-LANG!")
        end
    `)

    // Or use interpreter directly
    program, err := interpreter.ParseProgram([]byte(code))
    stats, err := interpreter.Execute(program, &interpreter.Config{})
}
```

## Version Information in Code

Version information is automatically injected during build via ldflags:

```bash
go build -ldflags "-X github.com/bonkzero404/uddin-lang/internal/version.Version=v1.0.0 \
                   -X github.com/bonkzero404/uddin-lang/internal/version.BuildTime=..." \
     -o uddinlang ./cmd/uddin-lang
```

The Makefile handles this automatically when using `make build` or `make install`.

## Checking Current Version

```bash
# Check version from git
git describe --tags --always

# Check version in binary
./uddinlang --version

# Check version using Makefile
make version
```

## CI/CD Integration

For automated releases, GitHub Actions can:

1. Build binaries with version info
2. Create git tags
3. Publish to GitHub Releases
4. Update version in go.mod (if needed)

## Best Practices

1. **Always tag releases**: Use annotated tags with meaningful messages
2. **Follow SemVer**: Increment version appropriately for changes
3. **Test before tagging**: Ensure all tests pass before creating a tag
4. **Update CHANGELOG**: Document changes in CHANGELOG.md or RELEASE_NOTES.md
5. **Sign tags** (optional): Use `git tag -s` for signed tags

## Troubleshooting

### Version shows "dev"

If version shows "dev", it means:

-   No git tag exists
-   Binary was built without version flags
-   Build process didn't inject version info

**Solution**: Use `make build` or `make install` instead of direct `go build`.

### go get fails

If `go get` fails:

-   Ensure the tag exists on GitHub
-   Check that go.mod is properly configured
-   Verify module path is correct

### Version mismatch

If installed version doesn't match tag:

-   Clear module cache: `go clean -modcache`
-   Reinstall: `go get -u github.com/bonkzero404/uddin-lang@v1.0.0`
