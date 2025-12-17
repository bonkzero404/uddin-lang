//go:build !versioninit
// +build !versioninit

package interpreter

import (
	"github.com/bonkzero404/uddin-lang/internal/version"
)

func init() {
	// Set version info getter from version package
	getVersionInfo = func() string {
		return version.GetVersion()
	}
}

