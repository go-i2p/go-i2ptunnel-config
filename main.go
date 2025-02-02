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
		Name:  "i2pconv",
		Usage: "Convert I2P tunnel configurations between formats",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "input-format, if",
				Usage:    "Input format (properties|ini|yaml)",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "output-format, of",
				Usage:    "Output format (properties|ini|yaml)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "strict",
				Usage: "Enable strict validation",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate without writing output",
			},
		},
		Action: i2pconv.ConvertCommand,
	}

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
