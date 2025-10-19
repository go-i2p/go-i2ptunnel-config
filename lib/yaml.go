package i2pconv

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

func (c *Converter) parseYAML(input []byte) (*TunnelConfig, error) {
	config := &TunnelConfig{}
	err := yaml.Unmarshal(input, config)
	if err != nil {
		return nil, c.enhanceYAMLError(input, err)
	}
	return config, nil
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
