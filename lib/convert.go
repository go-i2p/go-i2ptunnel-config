package i2pconv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

// BatchResult represents the result of processing a single file in batch mode
type BatchResult struct {
	InputFile    string
	OutputFile   string
	InputFormat  string
	OutputFormat string
	Success      bool
	Error        error
}

// ProcessBatch processes multiple files using glob patterns and returns results for each file.
// It continues processing even if some files fail, collecting all results for reporting.
//
// Parameters:
//   - pattern: Glob pattern to match input files
//   - c: CLI context containing flags and options
//
// Returns:
//   - []BatchResult: Results for each processed file
//   - error: Fatal error that prevented batch processing from starting
func ProcessBatch(pattern string, c *cli.Context) ([]BatchResult, error) {
	// Expand glob pattern to get list of files
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files match pattern '%s'", pattern)
	}

	// Get flags
	inputFormat := c.String("in-format")
	outputFormat := c.String("out-format")
	validateOnly := c.Bool("validate")
	strict := c.Bool("strict")
	dryRun := c.Bool("dry-run")

	// Process each file individually
	results := make([]BatchResult, 0, len(files))
	converter := &Converter{strict: strict}

	for _, inputFile := range files {
		result := BatchResult{
			InputFile:    inputFile,
			OutputFormat: outputFormat,
		}

		// Process single file using existing logic
		err := processSingleFile(inputFile, "", inputFormat, outputFormat, validateOnly, dryRun, converter)
		if err != nil {
			result.Success = false
			result.Error = err
		} else {
			result.Success = true
			result.OutputFile = generateOutputFilename(inputFile, outputFormat)

			// Detect input format for reporting
			if inputFormat == "" {
				detectedFormat, err := converter.DetectFormat(inputFile)
				if err == nil {
					result.InputFormat = detectedFormat
				}
			} else {
				result.InputFormat = inputFormat
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// reportBatchResults prints a summary of batch processing results and returns appropriate error.
// It reports both successful and failed file processing, providing clear feedback to users.
//
// Parameters:
//   - results: Slice of BatchResult containing processing results for each file
//   - validateOnly: Whether the operation was validation-only
//   - dryRun: Whether the operation was a dry-run
//
// Returns:
//   - error: Non-nil if any files failed processing, nil if all succeeded
func reportBatchResults(results []BatchResult, validateOnly, dryRun bool) error {
	successCount := 0
	failureCount := 0

	// Report individual results
	for _, result := range results {
		if result.Success {
			successCount++
			if validateOnly {
				fmt.Printf("✓ Configuration in '%s' is valid (%s format)\n",
					result.InputFile, result.InputFormat)
			} else if dryRun {
				// Individual dry-run output already printed during processing
				fmt.Printf("✓ Dry-run conversion '%s' (%s -> %s)\n",
					result.InputFile, result.InputFormat, result.OutputFormat)
			} else {
				fmt.Printf("✓ Converted '%s' (%s) -> '%s' (%s)\n",
					result.InputFile, result.InputFormat, result.OutputFile, result.OutputFormat)
			}
		} else {
			failureCount++
			fmt.Printf("✗ Failed to process '%s': %v\n", result.InputFile, result.Error)
		}
	}

	// Print summary
	total := len(results)
	if validateOnly {
		fmt.Printf("\nValidation summary: %d/%d files valid, %d failed\n", successCount, total, failureCount)
	} else if dryRun {
		fmt.Printf("\nDry-run summary: %d/%d files processed, %d failed\n", successCount, total, failureCount)
	} else {
		fmt.Printf("\nBatch conversion summary: %d/%d files converted, %d failed\n", successCount, total, failureCount)
	}

	// Return error if any files failed
	if failureCount > 0 {
		return fmt.Errorf("%d of %d files failed processing", failureCount, total)
	}

	return nil
}

// processSingleFile handles the conversion of a single file with the given parameters.
// This function contains the core conversion logic extracted from ConvertCommand to enable reuse
// in both single-file and batch processing modes.
//
// Parameters:
//   - inputFile: Path to input configuration file
//   - outputFile: Path for output file (empty string for auto-generation)
//   - inputFormat: Input format (empty string for auto-detection)
//   - outputFormat: Output format
//   - validateOnly: Whether to only validate without conversion
//   - dryRun: Whether to print output instead of writing to file
//   - converter: Converter instance with configuration
//
// Returns:
//   - error: Any error that occurred during processing
func processSingleFile(inputFile, outputFile, inputFormat, outputFormat string, validateOnly, dryRun bool, converter *Converter) error {
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
	if err := converter.validateWithFormat(config, inputFormat); err != nil {
		return fmt.Errorf("validation error in '%s': %w", inputFile, err)
	}

	// If validate-only mode, we're done
	if validateOnly {
		return nil
	}

	// Generate output
	outputData, err := converter.generateOutput(config, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to generate %s output: %w", outputFormat, err)
	}

	// Handle output - either print to console or write to file
	if dryRun {
		fmt.Printf("# Converted '%s' from %s to %s format:\n", inputFile, inputFormat, outputFormat)
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

	return nil
}

// ConvertCommand converts the input configuration file to the specified output format.
// It supports automatic format detection based on file extensions and provides comprehensive
// argument validation. The command can validate-only, perform dry-run conversions, write
// output to specified files, or process multiple files in batch mode.
//
// Parameters:
//   - c (*cli.Context): The CLI context containing the command-line arguments and flags.
//
// Arguments:
//   - input-file: Path to the input configuration file or glob pattern (required)
//   - output-file: Path for output file (optional, defaults based on input name and format)
//
// Flags:
//   - in-format: Input format (properties|ini|yaml) - auto-detected if not specified
//   - out-format: Output format (properties|ini|yaml) - defaults to yaml
//   - output: Output file path - takes precedence over positional output-file argument
//   - validate: Validate input without performing conversion
//   - strict: Enable strict validation of the configuration
//   - dry-run: Print output to console instead of writing to file
//   - batch: Process multiple files using glob patterns
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
// Batch Processing:
//   - When --batch flag is used, the input argument is treated as a glob pattern
//   - Multiple files are processed independently with individual success/failure reporting
//   - Processing continues even if some files fail
//
// Related:
//   - Converter.DetectFormat, Converter.ParseInput, Converter.validate, Converter.generateOutput
//   - ProcessBatch, processSingleFile
func ConvertCommand(c *cli.Context) error {
	// Validate required arguments
	if c.NArg() < 1 {
		return fmt.Errorf("input file is required\nUsage: %s <input-file> [output-file]", c.App.Name)
	}

	inputArg := c.Args().Get(0)
	outputFile := c.Args().Get(1)

	// Get flags with proper defaults
	inputFormat := c.String("in-format")
	outputFormat := c.String("out-format")
	outputFlag := c.String("output")
	validateOnly := c.Bool("validate")
	strict := c.Bool("strict")
	dryRun := c.Bool("dry-run")
	batchMode := c.Bool("batch")

	// Handle output file priority: --output flag takes precedence over positional argument
	if outputFlag != "" {
		outputFile = outputFlag
	}

	// Check for incompatible options in batch mode
	if batchMode {
		if outputFile != "" {
			return fmt.Errorf("cannot specify output file in batch mode - files are auto-generated")
		}

		// Process batch
		results, err := ProcessBatch(inputArg, c)
		if err != nil {
			return fmt.Errorf("batch processing failed: %w", err)
		}

		// Report results
		return reportBatchResults(results, validateOnly, dryRun)
	}

	// Single file processing (original behavior)
	inputFile := inputArg
	converter := &Converter{strict: strict}

	// Use extracted single file processing logic
	err := processSingleFile(inputFile, outputFile, inputFormat, outputFormat, validateOnly, dryRun, converter)
	if err != nil {
		return err
	}

	// Report success for single file (only if not validate-only or dry-run, as those print their own messages)
	if !validateOnly && !dryRun {
		// Auto-detect format for reporting if not specified
		reportInputFormat := inputFormat
		if reportInputFormat == "" {
			if detected, err := converter.DetectFormat(inputFile); err == nil {
				reportInputFormat = detected
			}
		}

		// Generate output filename for reporting if not specified
		reportOutputFile := outputFile
		if reportOutputFile == "" {
			reportOutputFile = generateOutputFilename(inputFile, outputFormat)
		}

		fmt.Printf("✓ Converted '%s' (%s) -> '%s' (%s)\n", inputFile, reportInputFormat, reportOutputFile, outputFormat)
	} else if validateOnly {
		// Auto-detect format for validation reporting if not specified
		reportInputFormat := inputFormat
		if reportInputFormat == "" {
			if detected, err := converter.DetectFormat(inputFile); err == nil {
				reportInputFormat = detected
			}
		}
		fmt.Printf("✓ Configuration in '%s' is valid (%s format)\n", inputFile, reportInputFormat)
	}

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
