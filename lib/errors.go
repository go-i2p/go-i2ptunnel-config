package i2pconv

import (
	"fmt"
	"strings"
)

// ConversionError represents an error during conversion
type ConversionError struct {
	Op  string
	Err error
}

// ValidationError represents an error during validation
type ValidationError struct {
	Config *TunnelConfig
	Err    error
}

// ParseError represents a parsing error with line context information.
// It provides detailed error reporting including the line number, context lines,
// and a clear error message to help users identify and fix configuration issues.
type ParseError struct {
	Line    int      // Line number where error occurred (1-indexed)
	Column  int      // Column number where error occurred (1-indexed, 0 if unknown)
	Content string   // The problematic line content
	Context []string // Surrounding lines for context (up to 2 lines before and after)
	Message string   // The error message
	Format  string   // The format being parsed (properties, ini, yaml)
}

func (e *ValidationError) Error() string {
	return "validation: " + e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (e *ConversionError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *ConversionError) Unwrap() error {
	return e.Err
}

// Error implements the error interface for ParseError with enhanced formatting.
// It returns a multi-line error message with context showing:
// - Error location (line and optionally column)
// - Surrounding context lines (numbered)
// - The problematic line highlighted with ">>>"
// - The specific error message
func (e *ParseError) Error() string {
	var sb strings.Builder

	// Write header with location
	if e.Column > 0 {
		sb.WriteString(fmt.Sprintf("parse error at line %d, column %d", e.Line, e.Column))
	} else {
		sb.WriteString(fmt.Sprintf("parse error at line %d", e.Line))
	}

	if e.Format != "" {
		sb.WriteString(fmt.Sprintf(" (%s format)", e.Format))
	}
	sb.WriteString(":\n")

	// Write context lines with line numbers
	if len(e.Context) > 0 {
		sb.WriteString("\n")
		startLine := e.Line - len(e.Context)/2
		if startLine < 1 {
			startLine = 1
		}

		for i, line := range e.Context {
			lineNum := startLine + i
			prefix := "   "
			if lineNum == e.Line {
				prefix = ">>>"
			}
			sb.WriteString(fmt.Sprintf("%s %4d | %s\n", prefix, lineNum, line))
		}
		sb.WriteString("\n")
	}

	// Write error message
	sb.WriteString(e.Message)

	return sb.String()
}

// Unwrap returns the underlying error if this ParseError wraps another error.
// This enables error chain inspection using errors.Is and errors.As.
func (e *ParseError) Unwrap() error {
	// ParseError doesn't wrap other errors by default, but we provide
	// this method for future extensibility
	return nil
}

// extractContext extracts surrounding lines from the input for error context.
// It returns up to contextLines before and after the target line.
//
// Parameters:
//   - input: The full input content as byte slice
//   - targetLine: The line number to extract context for (1-indexed)
//   - contextLines: Number of lines to extract before and after
//
// Returns:
//   - []string: Array of context lines including the target line
//   - string: The target line content (for convenience)
func extractContext(input []byte, targetLine, contextLines int) ([]string, string) {
	lines := strings.Split(string(input), "\n")

	// Ensure target line is valid
	if targetLine < 1 || targetLine > len(lines) {
		return nil, ""
	}

	// Calculate range (1-indexed converted to 0-indexed)
	targetIdx := targetLine - 1
	start := targetIdx - contextLines
	end := targetIdx + contextLines + 1

	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	context := lines[start:end]
	targetContent := lines[targetIdx]

	return context, targetContent
}

// newParseError creates a new ParseError with context extracted from the input.
// This is a convenience function for creating well-formatted parse errors.
//
// Parameters:
//   - input: The full input content being parsed
//   - line: Line number where error occurred (1-indexed)
//   - column: Column number where error occurred (1-indexed, use 0 if unknown)
//   - format: The format being parsed (properties, ini, yaml)
//   - message: The error message explaining what went wrong
//
// Returns:
//   - *ParseError: A fully populated ParseError with context
func newParseError(input []byte, line, column int, format, message string) *ParseError {
	context, content := extractContext(input, line, 2)

	return &ParseError{
		Line:    line,
		Column:  column,
		Content: content,
		Context: context,
		Message: message,
		Format:  format,
	}
}
