package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	
	// Should contain "Uddin-Lang"
	if !strings.Contains(version, "Uddin-Lang") {
		t.Errorf("Version should contain 'Uddin-Lang', got: %s", version)
	}
	
	// Should not be empty
	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetVersionShort(t *testing.T) {
	short := GetVersionShort()
	
	// Should not be empty
	if short == "" {
		t.Error("VersionShort should not be empty")
	}
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are accessible
	if Version == "" && GitTag == "" {
		// In dev mode, this is OK
		t.Log("Version variables are empty (dev build)")
	}
}

