package i2pconv

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// parseYAML parses YAML using the standard nested structure with "tunnels" map.
// This is the go-i2p format where tunnels are defined in a "tunnels" map.
// The parser extracts the first tunnel for single-tunnel conversion workflows.
// The tunnel name is set from the map key.
func (c *Converter) parseYAML(input []byte) (*TunnelConfig, error) {
	type wrapper struct {
		Tunnels map[string]*TunnelConfig `yaml:"tunnels"`
	}

	var w wrapper
	err := yaml.Unmarshal(input, &w)
	if err != nil {
		return nil, c.enhanceYAMLError(input, err)
	}

	// Extract the first tunnel (for single-tunnel conversion)
	if len(w.Tunnels) == 0 {
		return nil, fmt.Errorf("yaml: no tunnels found in tunnels map")
	}

	// Return the first tunnel with its name set from the map key
	for name, config := range w.Tunnels {
		config.Name = name
		return config, nil
	}

	return nil, fmt.Errorf("yaml: failed to extract tunnel from nested structure")
}

// enhanceYAMLError wraps YAML parsing errors with line context.
// The yaml.v2 library provides line numbers in its error messages.
func (c *Converter) enhanceYAMLError(input []byte, err error) error {
	// yaml.v2 errors typically include "line N" in the message
	errMsg := err.Error()
	var lineNum int
	var matched bool

	// Try multiple patterns for extracting line numbers from YAML errors
	// Pattern 1: "yaml: line N: error message"
	if _, scanErr := fmt.Sscanf(errMsg, "yaml: line %d:", &lineNum); scanErr == nil {
		matched = true
	}

	// Pattern 2: "line N: error message" (without yaml: prefix)
	if !matched {
		if _, scanErr := fmt.Sscanf(errMsg, "line %d:", &lineNum); scanErr == nil {
			matched = true
		}
	}

	// Pattern 3: Look for "line N" anywhere in the message
	if !matched {
		if idx := strings.Index(errMsg, "line "); idx != -1 {
			remainder := errMsg[idx+5:]
			if _, scanErr := fmt.Sscanf(remainder, "%d", &lineNum); scanErr == nil {
				matched = true
			}
		}
	}

	if matched && lineNum > 0 {
		// Extract the actual error message (remove yaml: and line info prefix)
		message := errMsg
		if idx := strings.Index(message, ": "); idx != -1 {
			message = message[idx+2:]
		}
		return newParseError(input, lineNum, 0, "yaml", message)
	}

	// If we can't extract line number, return original error
	return fmt.Errorf("yaml parse error: %w", err)
}

// generateYAML creates YAML output in the standard nested structure format.
// This is the go-i2p format where tunnels are defined in a "tunnels" map.
// The tunnel is keyed by its name in the map, allowing for future multi-tunnel support.
func (c *Converter) generateYAML(config *TunnelConfig) ([]byte, error) {
	type wrapper struct {
		Tunnels map[string]*TunnelConfig `yaml:"tunnels"`
	}

	out := wrapper{
		Tunnels: map[string]*TunnelConfig{
			config.Name: config,
		},
	}

	return yaml.Marshal(out)
}
