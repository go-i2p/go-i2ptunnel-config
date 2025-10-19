package i2pconv

import (
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		contains []string // Strings that should appear in the error message
	}{
		{
			name: "basic parse error with line number",
			err: &ParseError{
				Line:    5,
				Message: "invalid syntax",
				Format:  "properties",
			},
			contains: []string{
				"parse error at line 5",
				"properties format",
				"invalid syntax",
			},
		},
		{
			name: "parse error with line and column",
			err: &ParseError{
				Line:    10,
				Column:  25,
				Message: "unexpected character",
				Format:  "ini",
			},
			contains: []string{
				"parse error at line 10, column 25",
				"ini format",
				"unexpected character",
			},
		},
		{
			name: "parse error with context",
			err: &ParseError{
				Line:    3,
				Content: "invalid line",
				Context: []string{"line 1", "line 2", "invalid line", "line 4", "line 5"},
				Message: "malformed key=value pair",
				Format:  "ini",
			},
			contains: []string{
				"parse error at line 3",
				">>>    3 | invalid line",
				"malformed key=value pair",
				"line 1",
				"line 2",
				"line 4",
				"line 5",
			},
		},
		{
			name: "parse error without format",
			err: &ParseError{
				Line:    1,
				Message: "generic error",
			},
			contains: []string{
				"parse error at line 1",
				"generic error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			for _, expected := range tt.contains {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("error message missing expected substring:\n  expected: %q\n  message: %q",
						expected, errMsg)
				}
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	err := &ParseError{
		Line:    1,
		Message: "test error",
	}

	if unwrapped := err.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() should return nil for ParseError, got: %v", unwrapped)
	}
}

func TestExtractContext(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		targetLine   int
		contextLines int
		wantContext  []string
		wantContent  string
		wantNilCtx   bool
	}{
		{
			name: "extract context from middle of file",
			input: `line 1
line 2
line 3
line 4
line 5`,
			targetLine:   3,
			contextLines: 1,
			wantContext:  []string{"line 2", "line 3", "line 4"},
			wantContent:  "line 3",
		},
		{
			name: "extract context at beginning of file",
			input: `line 1
line 2
line 3`,
			targetLine:   1,
			contextLines: 2,
			wantContext:  []string{"line 1", "line 2", "line 3"},
			wantContent:  "line 1",
		},
		{
			name: "extract context at end of file",
			input: `line 1
line 2
line 3`,
			targetLine:   3,
			contextLines: 2,
			wantContext:  []string{"line 1", "line 2", "line 3"},
			wantContent:  "line 3",
		},
		{
			name: "extract with larger context window",
			input: `line 1
line 2
line 3
line 4
line 5
line 6
line 7`,
			targetLine:   4,
			contextLines: 2,
			wantContext:  []string{"line 2", "line 3", "line 4", "line 5", "line 6"},
			wantContent:  "line 4",
		},
		{
			name:         "invalid line number (too high)",
			input:        "line 1\nline 2",
			targetLine:   10,
			contextLines: 1,
			wantNilCtx:   true,
			wantContent:  "",
		},
		{
			name:         "invalid line number (zero)",
			input:        "line 1\nline 2",
			targetLine:   0,
			contextLines: 1,
			wantNilCtx:   true,
			wantContent:  "",
		},
		{
			name:         "invalid line number (negative)",
			input:        "line 1\nline 2",
			targetLine:   -1,
			contextLines: 1,
			wantNilCtx:   true,
			wantContent:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, content := extractContext([]byte(tt.input), tt.targetLine, tt.contextLines)

			if tt.wantNilCtx {
				if ctx != nil {
					t.Errorf("expected nil context, got: %v", ctx)
				}
				if content != tt.wantContent {
					t.Errorf("content = %q, want %q", content, tt.wantContent)
				}
				return
			}

			if len(ctx) != len(tt.wantContext) {
				t.Errorf("context length = %d, want %d\ncontext: %v\nwant: %v",
					len(ctx), len(tt.wantContext), ctx, tt.wantContext)
				return
			}

			for i, line := range ctx {
				if line != tt.wantContext[i] {
					t.Errorf("context[%d] = %q, want %q", i, line, tt.wantContext[i])
				}
			}

			if content != tt.wantContent {
				t.Errorf("content = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestNewParseError(t *testing.T) {
	input := []byte(`line 1
line 2
problematic line
line 4
line 5`)

	err := newParseError(input, 3, 10, "ini", "invalid character at position")

	if err.Line != 3 {
		t.Errorf("Line = %d, want 3", err.Line)
	}

	if err.Column != 10 {
		t.Errorf("Column = %d, want 10", err.Column)
	}

	if err.Format != "ini" {
		t.Errorf("Format = %q, want %q", err.Format, "ini")
	}

	if err.Message != "invalid character at position" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid character at position")
	}

	if err.Content != "problematic line" {
		t.Errorf("Content = %q, want %q", err.Content, "problematic line")
	}

	if len(err.Context) != 5 {
		t.Errorf("Context length = %d, want 5", len(err.Context))
	}

	// Verify the error message contains expected components
	errMsg := err.Error()
	expectedParts := []string{
		"parse error at line 3, column 10",
		"ini format",
		">>>    3 | problematic line",
		"invalid character at position",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("error message missing %q:\n%s", part, errMsg)
		}
	}
}

func TestConversionError_ErrorInterface(t *testing.T) {
	baseErr := newParseError([]byte("test"), 1, 0, "properties", "test error")
	convErr := &ConversionError{
		Op:  "parse",
		Err: baseErr,
	}

	// Test Error() method
	errMsg := convErr.Error()
	if !strings.Contains(errMsg, "parse:") {
		t.Errorf("ConversionError.Error() should contain operation, got: %s", errMsg)
	}

	// Test Unwrap() method
	if unwrapped := convErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("ConversionError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestValidationError_ErrorInterface(t *testing.T) {
	config := &TunnelConfig{Name: "test", Type: "client"}
	baseErr := newParseError([]byte("test"), 1, 0, "yaml", "validation failed")
	valErr := &ValidationError{
		Config: config,
		Err:    baseErr,
	}

	// Test Error() method
	errMsg := valErr.Error()
	if !strings.Contains(errMsg, "validation:") {
		t.Errorf("ValidationError.Error() should contain 'validation:', got: %s", errMsg)
	}

	// Test Unwrap() method
	if unwrapped := valErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("ValidationError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}
