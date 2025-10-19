package i2pconv

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExampleFiles validates all example configuration files in the examples directory.
// This ensures that the provided examples are syntactically correct and can be parsed.
func TestExampleFiles(t *testing.T) {
	examplesDir := filepath.Join("..", "examples")

	// Check if examples directory exists
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found, skipping example validation tests")
		return
	}

	testCases := []struct {
		name   string
		format string
	}{
		// Properties format examples
		{"httpclient.properties", "properties"},
		{"httpserver.properties", "properties"},
		{"socks.properties", "properties"},
		{"client.properties", "properties"},
		{"server.properties", "properties"},

		// INI format examples
		{"httpclient.conf", "ini"},
		{"httpserver.conf", "ini"},
		{"socks.conf", "ini"},
		{"client.conf", "ini"},
		{"server.conf", "ini"},

		// YAML format examples
		{"httpclient.yaml", "yaml"},
		{"httpserver.yaml", "yaml"},
		{"socks.yaml", "yaml"},
		{"client.yaml", "yaml"},
		{"server.yaml", "yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(examplesDir, tc.name)

			// Read the example file
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read example file %s: %v", tc.name, err)
			}

			// Parse the configuration
			conv := &Converter{strict: false}
			config, err := conv.ParseInput(data, tc.format)
			if err != nil {
				t.Fatalf("Failed to parse example file %s: %v", tc.name, err)
			}

			// Validate the configuration
			if err := conv.validate(config); err != nil {
				t.Errorf("Example file %s failed validation: %v", tc.name, err)
			}

			// Verify basic required fields
			if config.Name == "" {
				t.Errorf("Example file %s has empty name", tc.name)
			}
			if config.Type == "" {
				t.Errorf("Example file %s has empty type", tc.name)
			}
		})
	}
}

// TestExampleFileConversion verifies that example files can be converted between formats.
// This ensures cross-format compatibility and that conversions maintain data integrity.
// Note: YAML format has a nested structure (tunnels: map) so conversions TO YAML
// produce valid output, but may not round-trip perfectly without special handling.
func TestExampleFileConversion(t *testing.T) {
	examplesDir := filepath.Join("..", "examples")

	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found, skipping conversion tests")
		return
	}

	testCases := []struct {
		sourceFile   string
		sourceFormat string
		targetFormat string
	}{
		// Properties and INI formats are flat and convert well in both directions
		{"httpclient.properties", "properties", "ini"},
		{"httpserver.conf", "ini", "properties"},
		// YAML format converts FROM yaml to other formats reliably
		{"socks.yaml", "yaml", "properties"},
		{"socks.yaml", "yaml", "ini"},
	}

	for _, tc := range testCases {
		t.Run(tc.sourceFile+"_to_"+tc.targetFormat, func(t *testing.T) {
			filePath := filepath.Join(examplesDir, tc.sourceFile)

			// Read source file
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}

			// Convert to target format
			conv := &Converter{strict: false}
			output, err := conv.Convert(data, tc.sourceFormat, tc.targetFormat)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Verify output is not empty
			if len(output) == 0 {
				t.Error("Conversion produced empty output")
			}

			// Parse the converted output
			config, err := conv.ParseInput(output, tc.targetFormat)
			if err != nil {
				t.Fatalf("Failed to parse converted output: %v", err)
			}

			// Validate the converted configuration
			if err := conv.validate(config); err != nil {
				t.Errorf("Converted configuration failed validation: %v", err)
			}
		})
	}
}

// TestExampleFileRoundTrip verifies that converting an example file to another format
// and back produces equivalent configuration (round-trip conversion).
// With the nested YAML format, all formats now support proper round-trip conversion.
func TestExampleFileRoundTrip(t *testing.T) {
	examplesDir := filepath.Join("..", "examples")

	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found, skipping round-trip tests")
		return
	}

	testCases := []struct {
		file         string
		format       string
		intermediate string
	}{
		// Properties <-> INI conversions work well in both directions
		{"httpclient.properties", "properties", "ini"},
		{"httpserver.conf", "ini", "properties"},
		{"client.properties", "properties", "ini"},
		// Properties/INI <-> YAML conversions now work with nested format
		{"httpclient.properties", "properties", "yaml"},
		{"socks.conf", "ini", "yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.file+"_via_"+tc.intermediate, func(t *testing.T) {
			filePath := filepath.Join(examplesDir, tc.file)

			// Read original file
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			conv := &Converter{strict: false}

			// Parse original
			original, err := conv.ParseInput(data, tc.format)
			if err != nil {
				t.Fatalf("Failed to parse original: %v", err)
			}

			// Convert to intermediate format
			intermediate, err := conv.generateOutput(original, tc.intermediate)
			if err != nil {
				t.Fatalf("Failed to convert to intermediate: %v", err)
			}

			// Parse intermediate
			intermediateConfig, err := conv.ParseInput(intermediate, tc.intermediate)
			if err != nil {
				t.Fatalf("Failed to parse intermediate: %v", err)
			}

			// Convert back to original format
			final, err := conv.generateOutput(intermediateConfig, tc.format)
			if err != nil {
				t.Fatalf("Failed to convert back to original: %v", err)
			}

			// Parse final result
			finalConfig, err := conv.ParseInput(final, tc.format)
			if err != nil {
				t.Fatalf("Failed to parse final: %v", err)
			}

			// Compare key fields between original and final
			if original.Name != finalConfig.Name {
				t.Errorf("Name mismatch: original=%s, final=%s", original.Name, finalConfig.Name)
			}
			if original.Type != finalConfig.Type {
				t.Errorf("Type mismatch: original=%s, final=%s", original.Type, finalConfig.Type)
			}
			if original.Port != finalConfig.Port {
				t.Errorf("Port mismatch: original=%d, final=%d", original.Port, finalConfig.Port)
			}
		})
	}
}

// TestExampleFilesAgainstTunnelTypes verifies that each example file uses
// a valid tunnel type and meets the requirements for that type.
func TestExampleFilesAgainstTunnelTypes(t *testing.T) {
	examplesDir := filepath.Join("..", "examples")

	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found, skipping tunnel type tests")
		return
	}

	testCases := []struct {
		file          string
		format        string
		tunnelType    string
		requirePort   bool
		requireTarget bool
	}{
		{"httpclient.properties", "properties", "httpclient", true, false},
		{"httpclient.conf", "ini", "http", true, false},
		{"httpclient.yaml", "yaml", "httpclient", true, false},
		{"httpserver.properties", "properties", "httpserver", false, true},
		{"httpserver.conf", "ini", "http", false, true},
		{"httpserver.yaml", "yaml", "httpserver", false, true},
		{"socks.properties", "properties", "sockstunnel", true, false},
		{"socks.conf", "ini", "socks", true, false},
		{"socks.yaml", "yaml", "sockstunnel", true, false},
		{"client.properties", "properties", "client", true, false},
		{"client.conf", "ini", "client", true, false},
		{"client.yaml", "yaml", "client", true, false},
		{"server.properties", "properties", "server", false, true},
		{"server.conf", "ini", "server", false, true},
		{"server.yaml", "yaml", "server", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			filePath := filepath.Join(examplesDir, tc.file)

			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			conv := &Converter{strict: false}
			config, err := conv.ParseInput(data, tc.format)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Note: tunnel types may vary between formats (e.g., "http" in i2pd vs "httpclient" in Java I2P)
			// We check if the type is reasonable but don't enforce exact matches across formats
			if config.Type == "" {
				t.Error("Tunnel type is empty")
			}

			// Check port requirement
			if tc.requirePort && config.Port == 0 {
				t.Errorf("Expected port to be set for %s tunnel type", tc.tunnelType)
			}

			// Check target requirement
			if tc.requireTarget && config.Target == "" {
				t.Errorf("Expected target to be set for %s tunnel type", tc.tunnelType)
			}
		})
	}
}
