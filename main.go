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
		Usage:   "Convert I2P tunnel configurations between Java I2P, i2pd, and go-i2p formats",
		Version: "0.33.0",
		Description: `A command line utility to convert I2P tunnel configurations between Java I2P, 
i2pd, and go-i2p formats with automatic format detection, validation, and 
batch processing capabilities.

SUPPORTED FORMATS:
  Java I2P   (.config, .properties, .prop) - Properties format used by Java I2P
  i2pd       (.conf, .ini)                  - INI format used by i2pd
  go-i2p     (.yaml, .yml)                  - YAML format used by go-i2p

FORMAT AUTO-DETECTION:
  Input format is automatically detected based on file extension. You can
  override this with the --in-format flag if needed (e.g., for non-standard
  extensions or stdin input).

COMMON USAGE PATTERNS:

  1. Simple conversion with auto-detection:
     $ go-i2ptunnel-config tunnel.config
     Converts tunnel.config to tunnel.yaml (default output format)

  2. Convert to specific format:
     $ go-i2ptunnel-config --out-format ini tunnel.properties
     Converts Java I2P properties to i2pd INI format

  3. Specify custom output filename:
     $ go-i2ptunnel-config -o my-tunnel.conf tunnel.yaml
     $ go-i2ptunnel-config tunnel.config custom-name.yaml
     Both flag and positional argument syntax supported

  4. Validate configuration without converting:
     $ go-i2ptunnel-config --validate tunnel.config
     $ go-i2ptunnel-config --validate --strict tunnel.yaml
     Use --strict for additional validation checks (port ranges, etc.)

  5. Preview conversion without writing files:
     $ go-i2ptunnel-config --dry-run tunnel.properties
     Shows converted output on console without creating files

  6. Batch convert multiple files:
     $ go-i2ptunnel-config --batch "*.config"
     $ go-i2ptunnel-config --batch --out-format ini "tunnels/*.properties"
     Processes all matching files, continues on errors

  7. Convert from non-standard extension:
     $ go-i2ptunnel-config --in-format properties --out-format yaml tunnel.txt
     Useful when file extension doesn't match actual format

  8. Migrate from Java I2P to i2pd:
     $ go-i2ptunnel-config --batch --out-format ini "~/.i2p/i2ptunnel.config.d/*.config"
     Converts all Java I2P tunnel configs to i2pd format

CONFIGURATION EXAMPLES:
  Ready-to-use configuration templates are available in the examples/ directory:
  - httpclient   : HTTP proxy for browsing I2P websites
  - httpserver   : Host your own I2P website (eepsite)
  - socks        : SOCKS proxy for I2P connections
  - client       : Generic client tunnel (forward local port to I2P destination)
  - server       : Generic server tunnel (expose local service on I2P)

  Each template includes detailed inline comments and is available in all three
  formats. See examples/README.md for more information.

VALIDATION MODES:
  --validate       : Checks required fields and tunnel type validity
  --validate --strict : Additional checks for port ranges, target formats, etc.

BATCH PROCESSING:
  When using --batch, the tool:
  - Accepts glob patterns (e.g., "*.config", "dir/**/*.properties")
  - Processes all matching files independently
  - Continues processing even if some files fail
  - Reports summary of successful/failed conversions
  - Cannot be used with --output flag (each file gets auto-generated name)

OUTPUT FILE NAMING:
  If no output file is specified, the tool automatically generates one based on:
  - Input filename (without extension) + output format extension
  - Example: tunnel.config -> tunnel.yaml (default)
  - Example: tunnel.config --out-format ini -> tunnel.conf

EXIT CODES:
  0 : Success
  1 : Error (invalid arguments, conversion failure, validation failure)

For more information, visit: https://github.com/go-i2p/go-i2ptunnel-config`,
		ArgsUsage: "<input-file> [output-file]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "in-format, if",
				Usage: "Override input format detection (properties|ini|yaml)",
			},
			&cli.StringFlag{
				Name:  "out-format, of",
				Usage: "Set output format: properties (Java I2P), ini (i2pd), yaml (go-i2p)",
				Value: "yaml",
			},
			&cli.StringFlag{
				Name:  "output, o",
				Usage: "Specify output file path (auto-generated from input filename if not set)",
			},
			&cli.BoolFlag{
				Name:  "validate",
				Usage: "Validate configuration without performing conversion",
			},
			&cli.BoolFlag{
				Name:  "strict",
				Usage: "Enable strict validation (port ranges, target formats, privileged port warnings)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview conversion output on console without writing files",
			},
			&cli.BoolFlag{
				Name:  "batch",
				Usage: "Process multiple files using glob patterns (e.g., \"*.config\")",
			},
		},
		Action: i2pconv.ConvertCommand,
	}

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
