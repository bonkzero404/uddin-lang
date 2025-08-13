package interpreter

import (
	"strings"
	"testing"
)

// TestMemoryLayoutIntegration tests the integration of memory layout optimizations
func TestMemoryLayoutIntegration(t *testing.T) {
	// Test with default configuration (memory layout disabled)
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		config.Stdout = &strings.Builder{}
		
		if config.MemoryLayout.EnableTaggedValues {
			t.Error("Expected TaggedValues to be disabled in default config")
		}
		
		if config.MemoryLayout.EnableCompactEnvironment {
			t.Error("Expected CompactEnvironment to be disabled in default config")
		}
		
		if config.MemoryLayout.EnableCacheFriendlyStructures {
			t.Error("Expected CacheFriendlyStructures to be disabled in default config")
		}
		
		// Test basic interpreter functionality
		program := `
			x = 42
			y = [1, 2, 3]
			z = {"key": "value"}
			print(x + len(y) + len(z))
		`
		
		ast, err := ParseProgram([]byte(program))
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		
		_, err = Execute(ast, config)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
	})
	
	// Test with experimental configuration (memory layout enabled)
	t.Run("ExperimentalConfig", func(t *testing.T) {
		config := ExperimentalConfig()
		config.Stdout = &strings.Builder{}
		
		if !config.MemoryLayout.EnableTaggedValues {
			t.Error("Expected TaggedValues to be enabled in experimental config")
		}
		
		if !config.MemoryLayout.EnableCompactEnvironment {
			t.Error("Expected CompactEnvironment to be enabled in experimental config")
		}
		
		if !config.MemoryLayout.EnableCacheFriendlyStructures {
			t.Error("Expected CacheFriendlyStructures to be enabled in experimental config")
		}
		
		// Test basic interpreter functionality with memory layout optimizations
		program := `
			x = 42
			y = [1, 2, 3]
			z = {"key": "value"}
			print(x + len(y) + len(z))
		`
		
		prog, err := ParseProgram([]byte(program))
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		
		_, err = Execute(prog, config)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
	})
}

// TestMemoryLayoutConfigurationAPI tests the memory layout configuration API
func TestMemoryLayoutConfigurationAPI(t *testing.T) {
	// Test default configuration
	defaultConfig := DefaultMemoryLayoutConfig()
	if defaultConfig.EnableTaggedValues {
		t.Error("Expected TaggedValues to be disabled by default")
	}
	if defaultConfig.EnableCompactEnvironment {
		t.Error("Expected CompactEnvironment to be disabled by default")
	}
	if defaultConfig.EnableCacheFriendlyStructures {
		t.Error("Expected CacheFriendlyStructures to be disabled by default")
	}
	
	// Test experimental configuration
	experimentalConfig := ExperimentalMemoryLayoutConfig()
	if !experimentalConfig.EnableTaggedValues {
		t.Error("Expected TaggedValues to be enabled in experimental config")
	}
	if !experimentalConfig.EnableCompactEnvironment {
		t.Error("Expected CompactEnvironment to be enabled in experimental config")
	}
	if !experimentalConfig.EnableCacheFriendlyStructures {
		t.Error("Expected CacheFriendlyStructures to be enabled in experimental config")
	}
	
	// Test global configuration setting
	SetGlobalMemoryLayoutConfig(experimentalConfig)
	if !IsTaggedValuesEnabled() {
		t.Error("Expected TaggedValues to be enabled globally")
	}
	if !IsCompactEnvironmentEnabled() {
		t.Error("Expected CompactEnvironment to be enabled globally")
	}
	if !IsCacheFriendlyStructuresEnabled() {
		t.Error("Expected CacheFriendlyStructures to be enabled globally")
	}
	
	// Reset to default
	SetGlobalMemoryLayoutConfig(defaultConfig)
	if IsTaggedValuesEnabled() {
		t.Error("Expected TaggedValues to be disabled after reset")
	}
	if IsCompactEnvironmentEnabled() {
		t.Error("Expected CompactEnvironment to be disabled after reset")
	}
	if IsCacheFriendlyStructuresEnabled() {
		t.Error("Expected CacheFriendlyStructures to be disabled after reset")
	}
}

// TestMemoryLayoutPerformance tests performance with and without memory layout optimizations
func TestMemoryLayoutPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	program := `
		// Create arrays and maps
		arr = []
		for (i in range(1000)):
			append(arr, i)
		end
		
		map_data = {}
		for (i in range(500)):
			key = "key_" + str(i)
			map_data[key] = i * 2
		end
		
		// Access data
		sum = 0
		for (val in arr):
			sum = sum + val
		end
		
		for (i in range(500)):
			key = "key_" + str(i)
			if (key in map_data) then:
				sum = sum + map_data[key]
			end
		end
		
		print("Sum: " + str(sum))
	`
	
	ast, err := ParseProgram([]byte(program))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	
	// Test with default configuration
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		config.Stdout = &strings.Builder{}
		
		_, err := Execute(ast, config)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
	})
	
	// Test with experimental configuration
	t.Run("ExperimentalConfig", func(t *testing.T) {
		config := ExperimentalConfig()
		config.Stdout = &strings.Builder{}
		
		_, err := Execute(ast, config)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
	})
}

// BenchmarkMemoryLayoutDefault benchmarks interpreter with default configuration
func BenchmarkMemoryLayoutDefault(b *testing.B) {
	program := `
		arr = []
		for (i in range(100)):
			append(arr, i)
		end
		sum = 0
		for (val in arr):
			sum = sum + val
		end
	`
	
	prog, err := ParseProgram([]byte(program))
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}
	
	config := DefaultConfig()
	config.Stdout = &strings.Builder{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Execute(prog, config)
		if err != nil {
			b.Fatalf("Execute error: %v", err)
		}
	}
}

// BenchmarkMemoryLayoutExperimental benchmarks interpreter with experimental memory layout
func BenchmarkMemoryLayoutExperimental(b *testing.B) {
	program := `
		arr = []
		for (i in range(100)):
			append(arr, i)
		end
		sum = 0
		for (val in arr):
			sum = sum + val
		end
	`
	
	prog, err := ParseProgram([]byte(program))
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}
	
	config := ExperimentalConfig()
	config.Stdout = &strings.Builder{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Execute(prog, config)
		if err != nil {
			b.Fatalf("Execute error: %v", err)
		}
	}
}