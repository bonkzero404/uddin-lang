package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the application.
	// This should be set via build flags: -ldflags="-X github.com/bonkzero404/uddin-lang/internal/version.Version=v1.0.0"
	Version = "dev"

	// BuildTime is the build time of the application.
	// This should be set via build flags: -ldflags="-X github.com/bonkzero404/uddin-lang/internal/version.BuildTime=..."
	BuildTime = "unknown"

	// GitCommit is the git commit hash of the build.
	// This should be set via build flags: -ldflags="-X github.com/bonkzero404/uddin-lang/internal/version.GitCommit=..."
	GitCommit = "unknown"

	// GitTag is the git tag of the build (if any).
	// This should be set via build flags: -ldflags="-X github.com/bonkzero404/uddin-lang/internal/version.GitTag=..."
	GitTag = ""

	// GoVersion is the Go version used to build the application.
	GoVersion = runtime.Version()
)

// GetVersion returns the full version information string.
func GetVersion() string {
	version := Version
	if GitTag != "" {
		version = GitTag
	}

	info := fmt.Sprintf("Uddin-Lang %s", version)
	if BuildTime != "unknown" {
		info += fmt.Sprintf(" (built: %s)", BuildTime)
	}
	if GitCommit != "unknown" {
		shortCommit := GitCommit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		info += fmt.Sprintf(" (commit: %s)", shortCommit)
	}
	info += fmt.Sprintf(" (go: %s)", GoVersion)

	return info
}

// GetVersionShort returns the short version string.
func GetVersionShort() string {
	if GitTag != "" {
		return GitTag
	}
	return Version
}

