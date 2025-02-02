package i2pconv

import (
	"fmt"
)

// TunnelConfig represents a normalized tunnel configuration
type TunnelConfig struct {
	Name          string                 `yaml:"name"`
	Type          string                 `yaml:"type"`
	Interface     string                 `yaml:"interface,omitempty"`
	Port          int                    `yaml:"port,omitempty"`
	PersistentKey bool                   `yaml:"persistentKey,omitempty"`
	Description   string                 `yaml:"description,omitempty"`
	I2CP          map[string]interface{} `yaml:"i2cp,omitempty"`
	Tunnel        map[string]interface{} `yaml:"options,omitempty"`
	Inbound       map[string]interface{} `yaml:"inbound,omitempty"`
	Outbound      map[string]interface{} `yaml:"outbound,omitempty"`
}

// generateOutput generates the output based on the format
func (c *Converter) generateOutput(config *TunnelConfig, format string) ([]byte, error) {
	switch format {
	case "properties":
		return c.generateJavaProperties(config)
	case "yaml":
		return c.generateYAML(config)
	case "ini":
		return c.generateINI(config)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// validate checks the configuration for any errors
func (c *Converter) validate(config *TunnelConfig) error {
	if config.Name == "" {
		return fmt.Errorf("name is required")
	}
	if config.Type == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}

// Converter handles configuration format conversions
type Converter struct {
	strict bool
	// logger Logger
}

// Convert transforms between configuration formats
func (c *Converter) Convert(input []byte, inFormat, outFormat string) ([]byte, error) {
	config, err := c.parseInput(input, inFormat)
	if err != nil {
		return nil, &ConversionError{Op: "parse", Err: err}
	}

	if err := c.validate(config); err != nil {
		return nil, &ValidationError{Config: config, Err: err}
	}

	return c.generateOutput(config, outFormat)
}

// parseInput parses the input based on the format
func (c *Converter) parseInput(input []byte, format string) (*TunnelConfig, error) {
	switch format {
	case "properties":
		return c.parseJavaProperties(input)
	case "yaml":
		return c.parseYAML(input)
	case "ini":
		return c.parseINI(input)
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}
