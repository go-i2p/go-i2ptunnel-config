package i2pconv

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

// ConvertCommand converts the input configuration file to the specified output format.
// It reads the input file, parses it according to the specified input format, validates the configuration,
// and generates the output in the specified format. If the dry-run flag is set, it prints the output to the console
// instead of writing it to a file.
//
// Parameters:
//   - c (*cli.Context): The CLI context containing the command-line arguments and flags.
//
// Flags:
//   - input-format (string): The format of the input configuration file.
//   - output-format (string): The format of the output configuration file.
//   - strict (bool): If true, enables strict validation of the configuration.
//   - dry-run (bool): If true, prints the output to the console instead of writing it to a file.
//
// Returns:
//   - error: An error if any step of the conversion process fails, including reading the input file,
//     parsing the input, validating the configuration, or writing the output file.
//
// Errors:
//   - "failed to read input file": If the input file cannot be read.
//   - "failed to parse input": If the input data cannot be parsed according to the specified format.
//   - "validation error": If the configuration fails validation.
//   - "failed to generate output": If the output data cannot be generated according to the specified format.
//   - "failed to write output file": If the output file cannot be written.
//
// Related:
//   - Converter.parseInput
//   - Converter.validate
//   - Converter.generateOutput
func ConvertCommand(c *cli.Context) error {
	inputFormat := c.StringSlice("input-format")[0]
	outputFormat := c.StringSlice("output-format")[0]
	strict := c.Bool("strict")
	dryRun := c.Bool("dry-run")

	converter := &Converter{strict: strict}

	inputData, err := os.ReadFile(c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	config, err := converter.parseInput(inputData, inputFormat)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if err := converter.validate(config); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	outputData, err := converter.generateOutput(config, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	if dryRun {
		fmt.Println(string(outputData))
	} else {
		outputFile := c.Args().Get(1)
		if err := os.WriteFile(outputFile, outputData, 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}

	return nil
}
