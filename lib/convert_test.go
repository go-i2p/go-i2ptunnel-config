package i2pconv

import (
	"fmt"
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

// TestConvertCommandOutputFile tests custom output file specification
func TestConvertCommandOutputFile(t *testing.T) {
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
		name           string
		args           []string
		expectedOutput string
		description    string
	}{
		{
			name:           "positional output file argument",
			args:           []string{"go-i2ptunnel-config", inputFile, "custom-positional.yaml"},
			expectedOutput: "custom-positional.yaml",
			description:    "Should use positional argument for output file",
		},
		{
			name:           "output flag short form",
			args:           []string{"go-i2ptunnel-config", "-o", "custom-flag-short.yaml", inputFile},
			expectedOutput: "custom-flag-short.yaml",
			description:    "Should use -o flag for output file",
		},
		{
			name:           "output flag long form",
			args:           []string{"go-i2ptunnel-config", "--output", "custom-flag-long.yaml", inputFile},
			expectedOutput: "custom-flag-long.yaml",
			description:    "Should use --output flag for output file",
		},
		{
			name:           "flag takes precedence over positional",
			args:           []string{"go-i2ptunnel-config", "--output", "flag-wins.yaml", inputFile, "positional-loses.yaml"},
			expectedOutput: "flag-wins.yaml",
			description:    "Flag should take precedence over positional argument",
		},
		{
			name:           "relative path output",
			args:           []string{"go-i2ptunnel-config", "-o", "subdir/output.yaml", inputFile},
			expectedOutput: "subdir/output.yaml",
			description:    "Should handle relative paths in output",
		},
		{
			name:           "different output format with custom name",
			args:           []string{"go-i2ptunnel-config", "--out-format", "ini", "-o", "custom.conf", inputFile},
			expectedOutput: "custom.conf",
			description:    "Should use custom name even with different format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subdir if needed for relative path test
			if strings.Contains(tt.expectedOutput, "/") {
				dir := filepath.Dir(filepath.Join(tempDir, tt.expectedOutput))
				os.MkdirAll(dir, 0755)
			}

			// Save current directory and change to temp directory
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)
			os.Chdir(tempDir)

			app := &cli.App{
				Name: "go-i2ptunnel-config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out-format",
						Value: "yaml",
					},
					&cli.StringFlag{
						Name: "output, o",
					},
				},
				Action: ConvertCommand,
			}

			err := app.Run(tt.args)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			// Verify the output file exists
			expectedPath := filepath.Join(tempDir, tt.expectedOutput)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("%s: expected output file %q was not created", tt.description, tt.expectedOutput)
			} else {
				// Clean up
				os.Remove(expectedPath)
				// Clean up subdir if it was created
				if strings.Contains(tt.expectedOutput, "/") {
					dir := filepath.Dir(expectedPath)
					os.Remove(dir) // Will only remove if empty
				}
			}
		})
	}
}

// TestConvertCommandOutputValidation tests output file validation scenarios
func TestConvertCommandOutputValidation(t *testing.T) {
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

	// Create a read-only directory to test permission errors
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.Mkdir(readOnlyDir, 0444)
	if err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Ensure cleanup can happen

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		description string
	}{
		{
			name:        "invalid output directory",
			args:        []string{"go-i2ptunnel-config", "-o", "/nonexistent/dir/output.yaml", inputFile},
			expectError: true,
			errorMsg:    "failed to write output file",
			description: "Should fail when output directory doesn't exist",
		},
		{
			name:        "read-only directory",
			args:        []string{"go-i2ptunnel-config", "-o", filepath.Join(readOnlyDir, "output.yaml"), inputFile},
			expectError: true,
			errorMsg:    "failed to write output file",
			description: "Should fail when output directory is read-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current directory and change to temp directory
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)
			os.Chdir(tempDir)

			app := &cli.App{
				Name: "go-i2ptunnel-config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out-format",
						Value: "yaml",
					},
					&cli.StringFlag{
						Name: "output, o",
					},
				},
				Action: ConvertCommand,
			}

			err := app.Run(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("%s: expected error to contain %q, got: %q", tt.description, tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
			}
		})
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

// TestProcessBatchWithGlob tests the glob pattern functionality using filepath.Glob directly
func TestProcessBatchWithGlob(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test properties files
	validProperties := `name=test-tunnel
type=httpclient
interface=127.0.0.1
listenPort=8080
`

	// Write test files
	testFiles := map[string]string{
		"tunnel1.properties": validProperties,
		"tunnel2.config":     validProperties,
		"tunnel3.properties": validProperties,
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name        string
		pattern     string
		expectFiles int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid glob pattern for properties",
			pattern:     filepath.Join(tempDir, "*.properties"),
			expectFiles: 2,
			expectError: false,
		},
		{
			name:        "valid glob pattern for config files",
			pattern:     filepath.Join(tempDir, "*.config"),
			expectFiles: 1,
			expectError: false,
		},
		{
			name:        "no matching files",
			pattern:     filepath.Join(tempDir, "*.nonexistent"),
			expectFiles: 0,
			expectError: false,
		},
		{
			name:        "all files",
			pattern:     filepath.Join(tempDir, "*"),
			expectFiles: 3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := filepath.Glob(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got: %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(files) != tt.expectFiles {
				t.Errorf("expected %d files, got %d", tt.expectFiles, len(files))
			}
		})
	}
}

// TestProcessSingleFile tests the extracted single file processing logic
func TestProcessSingleFile(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	validProperties := `name=test-tunnel
type=httpclient
interface=127.0.0.1
listenPort=8080
`

	invalidProperties := `name=
type=invalid-type
`

	// Write test files
	validFile := filepath.Join(tempDir, "valid.properties")
	invalidFile := filepath.Join(tempDir, "invalid.properties")

	err := os.WriteFile(validFile, []byte(validProperties), 0644)
	if err != nil {
		t.Fatalf("failed to create valid test file: %v", err)
	}

	err = os.WriteFile(invalidFile, []byte(invalidProperties), 0644)
	if err != nil {
		t.Fatalf("failed to create invalid test file: %v", err)
	}

	converter := &Converter{strict: false}

	tests := []struct {
		name         string
		inputFile    string
		outputFile   string
		inputFormat  string
		outputFormat string
		validateOnly bool
		dryRun       bool
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid file conversion",
			inputFile:    validFile,
			inputFormat:  "properties",
			outputFormat: "yaml",
			dryRun:       true,
			expectError:  false,
		},
		{
			name:         "valid file validation only",
			inputFile:    validFile,
			inputFormat:  "properties",
			outputFormat: "yaml",
			validateOnly: true,
			expectError:  false,
		},
		{
			name:         "invalid file validation",
			inputFile:    invalidFile,
			inputFormat:  "properties",
			outputFormat: "yaml",
			validateOnly: true,
			expectError:  true,
			errorMsg:     "validation error",
		},
		{
			name:         "auto-detect format",
			inputFile:    validFile,
			inputFormat:  "", // Auto-detect
			outputFormat: "yaml",
			dryRun:       true,
			expectError:  false,
		},
		{
			name:        "nonexistent file",
			inputFile:   filepath.Join(tempDir, "nonexistent.properties"),
			inputFormat: "properties",
			expectError: true,
			errorMsg:    "failed to read input file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processSingleFile(tt.inputFile, tt.outputFile, tt.inputFormat,
				tt.outputFormat, tt.validateOnly, tt.dryRun, converter)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestReportBatchResults tests the batch results reporting functionality
func TestReportBatchResults(t *testing.T) {
	tests := []struct {
		name         string
		results      []BatchResult
		validateOnly bool
		dryRun       bool
		expectError  bool
	}{
		{
			name: "all successful conversions",
			results: []BatchResult{
				{
					InputFile:    "file1.properties",
					OutputFile:   "file1.yaml",
					InputFormat:  "properties",
					OutputFormat: "yaml",
					Success:      true,
				},
				{
					InputFile:    "file2.properties",
					OutputFile:   "file2.yaml",
					InputFormat:  "properties",
					OutputFormat: "yaml",
					Success:      true,
				},
			},
			expectError: false,
		},
		{
			name: "mixed success and failure",
			results: []BatchResult{
				{
					InputFile:    "file1.properties",
					OutputFile:   "file1.yaml",
					InputFormat:  "properties",
					OutputFormat: "yaml",
					Success:      true,
				},
				{
					InputFile: "file2.properties",
					Success:   false,
					Error:     fmt.Errorf("validation failed"),
				},
			},
			expectError: true,
		},
		{
			name: "validation only mode",
			results: []BatchResult{
				{
					InputFile:   "file1.properties",
					InputFormat: "properties",
					Success:     true,
				},
			},
			validateOnly: true,
			expectError:  false,
		},
		{
			name: "dry run mode",
			results: []BatchResult{
				{
					InputFile:    "file1.properties",
					InputFormat:  "properties",
					OutputFormat: "yaml",
					Success:      true,
				},
			},
			dryRun:      true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reportBatchResults(tt.results, tt.validateOnly, tt.dryRun)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
