package i2pconv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/magiconair/properties"
)

func (c *Converter) parseJavaProperties(input []byte) (*TunnelConfig, error) {
	p := properties.MustLoadString(string(input))

	config := &TunnelConfig{
		I2CP:     make(map[string]interface{}),
		Tunnel:   make(map[string]interface{}),
		Inbound:  make(map[string]interface{}),
		Outbound: make(map[string]interface{}),
	}

	// Parse tunnel.*.name pattern
	for _, k := range p.Keys() {
		c.parsePropertyKey(k, p.GetString(k, ""), config)
	}

	return config, nil
}

func (c *Converter) parsePropertyKey(k, s string, config *TunnelConfig) {
	if strings.HasPrefix(k, "#") || strings.HasPrefix(k, "configFile") {
		return // Skip comments and config file path
	}

	// Handle flat keys
	switch k {
	case "name":
		config.Name = s
	case "type":
		config.Type = s
	case "interface":
		config.Interface = s
	case "listenPort":
		if port, err := strconv.Atoi(s); err == nil {
			config.Port = port
		}
	case "targetDestination":
		config.Target = s
	case "targetHost":
		config.Target = s // Alternative naming
	case "description":
		config.Description = s
	case "i2cpHost":
		if config.I2CP == nil {
			config.I2CP = make(map[string]interface{})
		}
		config.I2CP["host"] = s
	case "i2cpPort":
		if config.I2CP == nil {
			config.I2CP = make(map[string]interface{})
		}
		if port, err := strconv.Atoi(s); err == nil {
			config.I2CP["port"] = port
		}
	}

	// Handle prefixed keys
	if strings.HasPrefix(k, "option.i2cp.") {
		if config.I2CP == nil {
			config.I2CP = make(map[string]interface{})
		}
		key := strings.TrimPrefix(k, "option.i2cp.")
		config.I2CP[key] = parseValue(s)
	}
}

// Helper to parse property values with type conversion
func parseValue(s string) interface{} {
	// Try boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	// Try integer
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	// Try comma-separated list
	if strings.Contains(s, ",") {
		return strings.Split(s, ",")
	}

	// Default to string
	return s
}

// generateJavaProperties generates a Java properties file content based on the provided TunnelConfig.
// It constructs the properties as a byte slice.
//
// Parameters:
//   - config (*TunnelConfig): The configuration for the tunnel. It includes various fields such as Name, Type, Interface, Port, PersistentKey, Description, and maps for I2CP, Tunnel, Inbound, and Outbound options.
//
// Returns:
//   - ([]byte): A byte slice containing the generated properties file content.
//   - (error): An error if any occurs during the generation process.
//
// Notable Errors/Edge Cases:
//   - The function does not handle any specific errors internally but returns any error encountered during the string building process.
//
// Related Code Entities:
//   - TunnelConfig: The structure that holds the configuration for the tunnel.
func (c *Converter) generateJavaProperties(config *TunnelConfig) ([]byte, error) {
	var sb strings.Builder

	if config.Name != "" {
		sb.WriteString(fmt.Sprintf("name=%s\n", config.Name))
	}
	if config.Type != "" {
		sb.WriteString(fmt.Sprintf("type=%s\n", config.Type))
	}
	if config.Interface != "" {
		sb.WriteString(fmt.Sprintf("interface=%s\n", config.Interface))
	}
	if config.Port != 0 {
		sb.WriteString(fmt.Sprintf("listenPort=%d\n", config.Port))
	}
	if config.Target != "" {
		sb.WriteString(fmt.Sprintf("targetDestination=%s\n", config.Target))
	}
	if config.PersistentKey {
		sb.WriteString("option.persistentClientKey=true\n")
	}
	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("description=%s\n", config.Description))
	}

	for k, v := range config.I2CP {
		sb.WriteString(fmt.Sprintf("option.i2cp.%s=%v\n", k, v))
	}

	for k, v := range config.Tunnel {
		sb.WriteString(fmt.Sprintf("option.i2ptunnel.%s=%v\n", k, v))
	}

	for k, v := range config.Inbound {
		sb.WriteString(fmt.Sprintf("option.inbound.%s=%v\n", k, v))
	}

	for k, v := range config.Outbound {
		sb.WriteString(fmt.Sprintf("option.outbound.%s=%v\n", k, v))
	}

	return []byte(sb.String()), nil
}
