package i2pconv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

// TestGenerateOutputFilename tests the generateOutputFilename function with various inputs
func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		format    string
		expected  string
	}{
		{
			name:      "properties format with .config extension",
			inputFile: "tunnel.config",
			format:    "properties",
			expected:  "tunnel.properties",
		},
		{
			name:      "ini format with .properties extension",
			inputFile: "tunnel.properties",
			format:    "ini",
			expected:  "tunnel.conf",
		},
		{
			name:      "yaml format with .ini extension",
			inputFile: "tunnel.ini",
			format:    "yaml",
			expected:  "tunnel.yaml",
		},
		{
			name:      "file without extension",
			inputFile: "tunnel",
			format:    "yaml",
			expected:  "tunnel.yaml",
		},
		{
			name:      "file with multiple dots",
			inputFile: "tunnel.backup.config",
			format:    "ini",
			expected:  "tunnel.backup.conf",
		},
		{
			name:      "unknown format",
			inputFile: "tunnel.config",
			format:    "unknown",
			expected:  "tunnel.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOutputFilename(tt.inputFile, tt.format)
			if result != tt.expected {
				t.Errorf("generateOutputFilename(%q, %q) = %q, want %q", tt.inputFile, tt.format, result, tt.expected)
			}
		})
	}
}

// TestConvertCommandValidation tests argument validation in ConvertCommand
func TestConvertCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no arguments provided",
			args:        []string{"go-i2ptunnel-config"},
			expectError: true,
			errorMsg:    "input file is required",
		},
		{
			name:        "non-existent input file",
			args:        []string{"go-i2ptunnel-config", "nonexistent.config"},
			expectError: true,
			errorMsg:    "failed to read input file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.App{
				Name:   "go-i2ptunnel-config",
				Action: ConvertCommand,
			}

			err := app.Run(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, but got: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConvertCommandWithValidInput tests ConvertCommand with actual file conversion
func TestConvertCommandWithValidInput(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "test.properties")
	testContent := `name=testTunnel
type=httpclient
interface=127.0.0.1
listenPort=8080
`

	err := os.WriteFile(inputFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name            string
		args            []string
		expectError     bool
		checkOutputFile bool
		expectedOutput  string
	}{
		{
			name:        "validate only mode",
			args:        []string{"go-i2ptunnel-config", "--validate", inputFile},
			expectError: false,
		},
		{
			name:        "dry run mode",
			args:        []string{"go-i2ptunnel-config", "--dry-run", inputFile},
			expectError: false,
		},
		{
			name:            "convert with auto-detection",
			args:            []string{"go-i2ptunnel-config", inputFile},
			expectError:     false,
			checkOutputFile: true,
		},
		{
			name:            "convert to specific format",
			args:            []string{"go-i2ptunnel-config", "--out-format", "ini", inputFile},
			expectError:     false,
			checkOutputFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current directory and change to temp directory for test
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)
			os.Chdir(tempDir)

			app := &cli.App{
				Name: "go-i2ptunnel-config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "in-format",
						Usage: "Input format",
					},
					&cli.StringFlag{
						Name:  "out-format",
						Usage: "Output format",
						Value: "yaml",
					},
					&cli.BoolFlag{
						Name:  "validate",
						Usage: "Validate only",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Dry run",
					},
					&cli.BoolFlag{
						Name:  "strict",
						Usage: "Strict validation",
					},
				},
				Action: ConvertCommand,
			}

			err := app.Run(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check if output file was created when expected
				if tt.checkOutputFile {
					var expectedFile string
					if strings.Contains(strings.Join(tt.args, " "), "--out-format ini") {
						expectedFile = "test.conf"
					} else {
						expectedFile = "test.yaml" // default format
					}

					if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
						t.Errorf("expected output file %q was not created", expectedFile)
					} else {
						// Clean up the created file
						os.Remove(expectedFile)
					}
				}
			}
		})
	}
}

// TestFormatAutoDetection tests the integration of format auto-detection in CLI
func TestFormatAutoDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Test files with different extensions
	testFiles := map[string]string{
		"test.properties": `name=test\ntype=httpclient\n`,
		"test.config":     `name=test\ntype=httpclient\n`,
		"test.conf":       `name = test\ntype = httpclient\n`,
		"test.yaml":       `tunnels:\n  test:\n    name: test\n    type: httpclient\n`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}

		// Test that format auto-detection works
		converter := &Converter{}
		format, err := converter.DetectFormat(filePath)
		if err != nil {
			t.Errorf("failed to detect format for %s: %v", filename, err)
			continue
		}

		// Verify the detected format makes sense
		switch filepath.Ext(filename) {
		case ".properties", ".config":
			if format != "properties" {
				t.Errorf("expected properties format for %s, got %s", filename, format)
			}
		case ".conf":
			if format != "ini" {
				t.Errorf("expected ini format for %s, got %s", filename, format)
			}
		case ".yaml":
			if format != "yaml" {
				t.Errorf("expected yaml format for %s, got %s", filename, format)
			}
		}
	}
}

// TestConvertCommandErrorHandling tests various error conditions
func TestConvertCommandErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file with unsupported extension
	unsupportedFile := filepath.Join(tempDir, "test.unsupported")
	err := os.WriteFile(unsupportedFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a file with invalid content
	invalidFile := filepath.Join(tempDir, "test.properties")
	err = os.WriteFile(invalidFile, []byte("invalid content without required fields"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		errorMsg string
	}{
		{
			name:     "unsupported file extension",
			args:     []string{"go-i2ptunnel-config", unsupportedFile},
			errorMsg: "failed to detect input format",
		},
		{
			name:     "invalid configuration content",
			args:     []string{"go-i2ptunnel-config", invalidFile},
			errorMsg: "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.App{
				Name: "go-i2ptunnel-config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out-format",
						Value: "yaml",
					},
				},
				Action: ConvertCommand,
			}

			err := app.Run(tt.args)
			if err == nil {
				t.Errorf("expected error but got none")
			} else if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("expected error to contain %q, got: %q", tt.errorMsg, err.Error())
			}
		})
	}
}
