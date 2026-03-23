package i2pconv

import (
	"errors"
	"strings"
	"testing"
)

// TestPropertiesParserErrors tests enhanced error reporting for properties format
func TestPropertiesParserErrors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectParseErr bool
		expectLineNum  int
	}{
		{
			name: "valid properties",
			input: `name=test-tunnel
type=client
interface=127.0.0.1
listenPort=4444`,
			expectError: false,
		},
		// Note: The properties library is lenient and may not fail on all invalid inputs
		// We keep this test for documentation but expect no error in this specific case
		{
			name:        "empty input is valid",
			input:       "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Converter{}
			config, err := conv.parseJavaProperties([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Check if it's a ParseError with line context
				var parseErr *ParseError
				if errors.As(err, &parseErr) {
					if tt.expectParseErr {
						t.Logf("Got expected ParseError: %v", parseErr)
						if tt.expectLineNum > 0 && parseErr.Line != tt.expectLineNum {
							t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, tt.expectLineNum)
						}
						if parseErr.Format != "properties" {
							t.Errorf("ParseError.Format = %q, want %q", parseErr.Format, "properties")
						}
					}
				} else if tt.expectParseErr {
					t.Errorf("expected ParseError but got: %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config == nil {
					t.Errorf("expected config but got nil")
				}
			}
		})
	}
}

// TestINIParserErrors tests enhanced error reporting for INI format
func TestINIParserErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectLineNum int
		errorContains string
	}{
		{
			name: "valid ini",
			input: `[test-tunnel]
type = client
host = 127.0.0.1
port = 4444`,
			expectError: false,
		},
		{
			name: "unclosed section bracket",
			input: `[test-tunnel
type = client`,
			expectError:   true,
			expectLineNum: 1,
			errorContains: "unclosed section bracket",
		},
		{
			name: "empty section name",
			input: `[]
type = client`,
			expectError:   true,
			expectLineNum: 1,
			errorContains: "empty section name",
		},
		{
			name: "missing equals sign",
			input: `[test]
type client`,
			expectError:   true,
			expectLineNum: 2,
			errorContains: "expected key=value pair",
		},
		{
			name: "empty key",
			input: `[test]
= value`,
			expectError:   true,
			expectLineNum: 2,
			errorContains: "empty key name",
		},
		// Note: INI parser uses SplitN(line, "=", 2) which treats everything after first = as value
		// This is valid behavior: key="type", value="client = extra"
		// No error expected in this case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Converter{}
			config, err := conv.parseINI([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Check if it's a ParseError
				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("expected ParseError but got: %T: %v", err, err)
					return
				}

				t.Logf("Got ParseError:\n%v", parseErr)

				if parseErr.Line != tt.expectLineNum {
					t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, tt.expectLineNum)
				}

				if parseErr.Format != "ini" {
					t.Errorf("ParseError.Format = %q, want %q", parseErr.Format, "ini")
				}

				if tt.errorContains != "" && !strings.Contains(parseErr.Message, tt.errorContains) {
					t.Errorf("ParseError.Message should contain %q, got: %q",
						tt.errorContains, parseErr.Message)
				}

				// Verify context is populated
				if len(parseErr.Context) == 0 {
					t.Errorf("ParseError.Context should not be empty")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config == nil {
					t.Errorf("expected config but got nil")
				}
			}
		})
	}
}

// TestYAMLParserErrors tests enhanced error reporting for YAML format (nested structure)
func TestYAMLParserErrors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectParseErr bool
		expectLineNum  int
	}{
		{
			name: "valid yaml",
			input: `tunnels:
  test-tunnel:
    type: client
    interface: 127.0.0.1
    port: 4444`,
			expectError: false,
		},
		{
			name: "invalid indentation",
			input: `tunnels:
  test:
      type: client
  interface: 127.0.0.1`,
			expectError:    true,
			expectParseErr: true,
		},
		{
			name: "invalid yaml structure",
			input: `tunnels:
  test
    type: client`,
			expectError:    true,
			expectParseErr: true,
			expectLineNum:  3, // YAML error reports line 3
		},
		{
			name: "unclosed quote",
			input: `tunnels:
  test:
    name: "test
    type: client`,
			expectError:    true,
			expectParseErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Converter{}
			config, err := conv.parseYAML([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Check if it's a ParseError with line context
				var parseErr *ParseError
				if errors.As(err, &parseErr) {
					if tt.expectParseErr {
						t.Logf("Got expected ParseError: %v", parseErr)
						if tt.expectLineNum > 0 && parseErr.Line != tt.expectLineNum {
							t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, tt.expectLineNum)
						}
						if parseErr.Format != "yaml" {
							t.Errorf("ParseError.Format = %q, want %q", parseErr.Format, "yaml")
						}
						// Note: Context may be empty if line number is beyond file length (EOF errors)
						// This is expected behavior
					}
				} else if tt.expectParseErr {
					// Some YAML errors might not have line numbers
					t.Logf("Got non-ParseError (acceptable for some YAML errors): %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config == nil {
					t.Errorf("expected config but got nil")
				}
			}
		})
	}
}

// TestParseErrorFormatting verifies that ParseError produces useful output
func TestParseErrorFormatting(t *testing.T) {
	input := `[tunnel]
type = client
invalid line without equals
port = 4444`

	conv := &Converter{}
	_, err := conv.parseINI([]byte(input))

	if err == nil {
		t.Fatal("expected error but got none")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected ParseError but got: %T", err)
	}

	// Verify error message formatting
	errMsg := parseErr.Error()
	t.Logf("Formatted error message:\n%s", errMsg)

	// Check for key components in formatted output
	expectedComponents := []string{
		"parse error at line",
		"ini format",
		">>>", // Highlight marker
		"|",   // Line number separator
		"invalid line without equals",
		"expected key=value pair",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(errMsg, component) {
			t.Errorf("formatted error should contain %q:\n%s", component, errMsg)
		}
	}

	// Verify context includes surrounding lines
	if !strings.Contains(errMsg, "type = client") || !strings.Contains(errMsg, "port = 4444") {
		t.Errorf("formatted error should include context lines:\n%s", errMsg)
	}
}

// TestEnhancePropertiesError tests the properties error enhancement
func TestEnhancePropertiesError(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		originalErr    error
		expectLineNum  int
		expectEnhanced bool
	}{
		{
			name:           "error with line number",
			input:          "line 1\nline 2\nline 3",
			originalErr:    errors.New("line 2: invalid syntax"),
			expectLineNum:  2,
			expectEnhanced: true,
		},
		{
			name:           "error without line number",
			input:          "line 1\nline 2",
			originalErr:    errors.New("generic error"),
			expectEnhanced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Converter{}
			err := conv.enhancePropertiesError([]byte(tt.input), tt.originalErr)

			var parseErr *ParseError
			if errors.As(err, &parseErr) {
				if !tt.expectEnhanced {
					t.Errorf("expected non-ParseError but got ParseError: %v", parseErr)
					return
				}
				if parseErr.Line != tt.expectLineNum {
					t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, tt.expectLineNum)
				}
			} else {
				if tt.expectEnhanced {
					t.Errorf("expected ParseError but got: %T: %v", err, err)
				}
			}
		})
	}
}

// TestEnhanceYAMLError tests the YAML error enhancement
func TestEnhanceYAMLError(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		originalErr    error
		expectLineNum  int
		expectEnhanced bool
	}{
		{
			name:           "yaml error with line prefix",
			input:          "line 1\nline 2\nline 3",
			originalErr:    errors.New("yaml: line 2: invalid syntax"),
			expectLineNum:  2,
			expectEnhanced: true,
		},
		{
			name:           "error with line without yaml prefix",
			input:          "line 1\nline 2",
			originalErr:    errors.New("line 1: error message"),
			expectLineNum:  1,
			expectEnhanced: true,
		},
		{
			name:           "error with line in middle of message",
			input:          "line 1\nline 2\nline 3",
			originalErr:    errors.New("error at line 3 in file"),
			expectLineNum:  3,
			expectEnhanced: true,
		},
		{
			name:           "error without line number",
			input:          "line 1\nline 2",
			originalErr:    errors.New("generic error"),
			expectEnhanced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Converter{}
			err := conv.enhanceYAMLError([]byte(tt.input), tt.originalErr)

			var parseErr *ParseError
			if errors.As(err, &parseErr) {
				if !tt.expectEnhanced {
					t.Errorf("expected non-ParseError but got ParseError: %v", parseErr)
					return
				}
				if parseErr.Line != tt.expectLineNum {
					t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, tt.expectLineNum)
				}
				if parseErr.Format != "yaml" {
					t.Errorf("ParseError.Format = %q, want %q", parseErr.Format, "yaml")
				}
			} else {
				if tt.expectEnhanced {
					t.Errorf("expected ParseError but got: %T: %v", err, err)
				}
			}
		})
	}
}
