package main

import (
	"log"
	"os"

	i2pconv "github.com/go-i2p/go-i2ptunnel-config/lib"
	"github.com/urfave/cli"
)

// CLI implementation
func main() {
	cmd := &cli.App{
		Name:    "go-i2ptunnel-config",
		Usage:   "Convert I2P tunnel configurations between formats",
		Version: "1.0.0",
		Description: `A command line utility to convert I2P tunnel configurations between Java I2P, i2pd, and go-i2p formats.

Supports automatic format detection based on file extensions:
  - .config, .properties, .prop -> Java I2P properties format
  - .conf, .ini -> i2pd INI format  
  - .yaml, .yml -> go-i2p YAML format

Examples:
  # Auto-detect input format, output as YAML
  go-i2ptunnel-config tunnel.config

  # Specify output format
  go-i2ptunnel-config -out-format ini tunnel.properties

  # Specify custom output file
  go-i2ptunnel-config -o /path/to/output.yaml tunnel.config
  go-i2ptunnel-config tunnel.config custom-name.yaml

  # Dry run to validate without writing
  go-i2ptunnel-config -dry-run tunnel.config

  # Specify both input and output formats explicitly
  go-i2ptunnel-config -in-format properties -out-format yaml tunnel.txt`,
		ArgsUsage: "<input-file> [output-file]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "in-format, if",
				Usage: "Input format (properties|ini|yaml) - auto-detected if not specified",
			},
			&cli.StringFlag{
				Name:  "out-format, of",
				Usage: "Output format (properties|ini|yaml) - defaults to yaml",
				Value: "yaml",
			},
			&cli.StringFlag{
				Name:  "output, o",
				Usage: "Output file path - auto-generated if not specified",
			},
			&cli.BoolFlag{
				Name:  "validate",
				Usage: "Validate input without conversion",
			},
			&cli.BoolFlag{
				Name:  "strict",
				Usage: "Enable strict validation",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print output to console instead of writing to file",
			},
		},
		Action: i2pconv.ConvertCommand,
	}

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
