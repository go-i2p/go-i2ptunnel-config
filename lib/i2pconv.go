package i2pconv

import (
	"fmt"
)

// TunnelConfig represents the configuration for an I2P tunnel.
// It includes various settings such as the tunnel name, type, interface, port, and other options.
//
// Fields:
// - Name: The name of the tunnel (string).
// - Type: The type of the tunnel (string).
// - Interface: The network interface to bind to (string, optional).
// - Port: The port number to bind to (int, optional).
// - PersistentKey: Indicates if the key should be persistent (bool, optional).
// - Description: A description of the tunnel (string, optional).
// - I2CP: A map of I2CP (I2P Control Protocol) options (map[string]interface{}, optional).
// - Tunnel: A map of tunnel-specific options (map[string]interface{}, optional).
// - Inbound: A map of inbound tunnel options (map[string]interface{}, optional).
// - Outbound: A map of outbound tunnel options (map[string]interface{}, optional).
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

// generateOutput generates the output for the given TunnelConfig in the specified format.
//
// Parameters:
//   - config (*TunnelConfig): The tunnel configuration to be converted.
//   - format (string): The desired output format. Supported formats are "properties", "yaml", and "ini".
//
// Returns:
//   - ([]byte): The generated output in the specified format.
//   - (error): An error if the specified format is unsupported or if there is an issue during generation.
//
// Errors:
//   - Returns an error if the specified format is unsupported.
//
// Related:
//   - generateJavaProperties
//   - generateYAML
//   - generateINI
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
}

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
