package i2pconv

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// ConvertCommand converts the input configuration file to the specified output format.
// It supports automatic format detection based on file extensions and provides comprehensive
// argument validation. The command can validate-only, perform dry-run conversions, or
// write output to specified files.
//
// Parameters:
//   - c (*cli.Context): The CLI context containing the command-line arguments and flags.
//
// Arguments:
//   - input-file: Path to the input configuration file (required)
//   - output-file: Path for output file (optional, defaults based on input name and format)
//
// Flags:
//   - in-format: Input format (properties|ini|yaml) - auto-detected if not specified
//   - out-format: Output format (properties|ini|yaml) - defaults to yaml
//   - output: Output file path - takes precedence over positional output-file argument
//   - validate: Validate input without performing conversion
//   - strict: Enable strict validation of the configuration
//   - dry-run: Print output to console instead of writing to file
//
// Returns:
//   - error: An error if any step fails, including argument validation, file I/O,
//     format detection, parsing, validation, or output generation.
//
// Format Detection:
//
//	Auto-detects input format based on file extension:
//	- .config, .properties, .prop -> properties format
//	- .conf, .ini -> ini format
//	- .yaml, .yml -> yaml format
//
// Related:
//   - Converter.DetectFormat, Converter.ParseInput, Converter.validate, Converter.generateOutput
func ConvertCommand(c *cli.Context) error {
	// Validate required arguments
	if c.NArg() < 1 {
		return fmt.Errorf("input file is required\nUsage: %s <input-file> [output-file]", c.App.Name)
	}

	inputFile := c.Args().Get(0)
	outputFile := c.Args().Get(1)

	// Get flags with proper defaults
	inputFormat := c.String("in-format")
	outputFormat := c.String("out-format")
	outputFlag := c.String("output")
	validateOnly := c.Bool("validate")
	strict := c.Bool("strict")
	dryRun := c.Bool("dry-run")

	// Handle output file priority: --output flag takes precedence over positional argument
	if outputFlag != "" {
		outputFile = outputFlag
	}

	// Initialize converter with options
	converter := &Converter{strict: strict}

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s': %w", inputFile, err)
	}

	// Auto-detect input format if not specified
	if inputFormat == "" {
		inputFormat, err = converter.DetectFormat(inputFile)
		if err != nil {
			return fmt.Errorf("failed to detect input format for '%s': %w (try specifying --in-format)", inputFile, err)
		}
	}

	// Parse input configuration
	config, err := converter.ParseInput(inputData, inputFormat)
	if err != nil {
		return fmt.Errorf("failed to parse %s input from '%s': %w", inputFormat, inputFile, err)
	}

	// Validate configuration
	if err := converter.validate(config); err != nil {
		return fmt.Errorf("validation error in '%s': %w", inputFile, err)
	}

	// If validate-only mode, we're done
	if validateOnly {
		fmt.Printf("✓ Configuration in '%s' is valid (%s format)\n", inputFile, inputFormat)
		return nil
	}

	// Generate output
	outputData, err := converter.generateOutput(config, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to generate %s output: %w", outputFormat, err)
	}

	// Handle output - either print to console or write to file
	if dryRun {
		fmt.Printf("# Converted from %s to %s format:\n", inputFormat, outputFormat)
		fmt.Println(string(outputData))
		return nil
	}

	// Determine output file name if not specified
	if outputFile == "" {
		outputFile = generateOutputFilename(inputFile, outputFormat)
	}

	// Write output file
	if err := os.WriteFile(outputFile, outputData, 0o644); err != nil {
		return fmt.Errorf("failed to write output file '%s': %w", outputFile, err)
	}

	fmt.Printf("✓ Converted '%s' (%s) -> '%s' (%s)\n", inputFile, inputFormat, outputFile, outputFormat)
	return nil
}

// generateOutputFilename creates a default output filename based on input file and desired format.
// It replaces the input file extension with the appropriate extension for the output format.
//
// Parameters:
//   - inputFile: Path to the input file
//   - format: Desired output format (properties|ini|yaml)
//
// Returns:
//   - string: Generated output filename with appropriate extension
func generateOutputFilename(inputFile, format string) string {
	base := inputFile

	// Remove existing extension
	if idx := strings.LastIndex(base, "."); idx != -1 {
		base = base[:idx]
	}

	// Add appropriate extension based on format
	switch format {
	case "properties":
		return base + ".properties"
	case "ini":
		return base + ".conf"
	case "yaml":
		return base + ".yaml"
	default:
		return base + ".out"
	}
}
