package i2pconv

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

// ConvertCommand handles the CLI command for converting tunnel configurations
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
