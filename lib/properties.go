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

// parsePropertyKey parses a single Java I2P property key-value pair and updates the TunnelConfig.
// Supports multiple Java I2P property patterns:
//
// Flat properties:
//   - name, type, interface, listenPort, targetDestination, targetHost, description
//   - proxyList, sharedClient, startOnLoad, accessList, targetPort, spoofedHost (stored in Tunnel map)
//   - i2cpHost, i2cpPort (stored in I2CP map)
//
// Numbered tunnel patterns:
//   - tunnel.N.property (e.g., tunnel.0.name, tunnel.1.type, tunnel.2.interface)
//
// Option prefixes:
//   - option.i2cp.* -> stored in I2CP map
//   - option.i2ptunnel.* -> stored in Tunnel map
//   - option.inbound.* -> stored in Inbound map
//   - option.outbound.* -> stored in Outbound map
//   - option.persistentClientKey -> sets PersistentKey field
//
// Comments (#) and configFile properties are ignored.
func (c *Converter) parsePropertyKey(k, s string, config *TunnelConfig) {
	if strings.HasPrefix(k, "#") || strings.HasPrefix(k, "configFile") {
		return // Skip comments and config file path
	}

	// Handle tunnel.N.property patterns for numbered tunnels
	if strings.HasPrefix(k, "tunnel.") && strings.Contains(k, ".") {
		parts := strings.SplitN(k, ".", 3)
		if len(parts) == 3 {
			// Format: tunnel.N.property (e.g., tunnel.0.name, tunnel.1.type)
			property := parts[2]
			c.parseNumberedTunnelProperty(property, s, config)
			return
		}
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
	case "targetPort":
		if port, err := strconv.Atoi(s); err == nil {
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel["targetPort"] = port
		}
	case "description":
		config.Description = s
	case "proxyList":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["proxyList"] = parseValue(s)
	case "sharedClient":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["sharedClient"] = parseValue(s)
	case "startOnLoad":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["startOnLoad"] = parseValue(s)
	case "accessList":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["accessList"] = parseValue(s)
	case "spoofedHost":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["spoofedHost"] = parseValue(s)
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
	} else if strings.HasPrefix(k, "option.i2ptunnel.") {
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		key := strings.TrimPrefix(k, "option.i2ptunnel.")
		config.Tunnel[key] = parseValue(s)
	} else if strings.HasPrefix(k, "option.inbound.") {
		if config.Inbound == nil {
			config.Inbound = make(map[string]interface{})
		}
		key := strings.TrimPrefix(k, "option.inbound.")
		config.Inbound[key] = parseValue(s)
	} else if strings.HasPrefix(k, "option.outbound.") {
		if config.Outbound == nil {
			config.Outbound = make(map[string]interface{})
		}
		key := strings.TrimPrefix(k, "option.outbound.")
		config.Outbound[key] = parseValue(s)
	} else if k == "option.persistentClientKey" {
		config.PersistentKey = parseValue(s).(bool)
	}
}

// parseNumberedTunnelProperty handles properties from tunnel.N.property patterns
// For numbered tunnel configurations, we treat each as a separate tunnel instance
// but for this converter we only support single tunnel configs, so we merge all numbered properties
func (c *Converter) parseNumberedTunnelProperty(property, value string, config *TunnelConfig) {
	switch property {
	case "name":
		// If config.Name is empty, use this as the primary name
		// If config.Name exists, treat as additional tunnel option
		if config.Name == "" {
			config.Name = value
		} else {
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel["alternateName"] = value
		}
	case "type":
		if config.Type == "" {
			config.Type = value
		}
	case "interface":
		if config.Interface == "" {
			config.Interface = value
		}
	case "listenPort":
		if config.Port == 0 {
			if port, err := strconv.Atoi(value); err == nil {
				config.Port = port
			}
		}
	case "targetDestination":
		if config.Target == "" {
			config.Target = value
		}
	case "targetHost":
		if config.Target == "" {
			config.Target = value
		}
	case "targetPort":
		if port, err := strconv.Atoi(value); err == nil {
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel["targetPort"] = port
		}
	case "description":
		if config.Description == "" {
			config.Description = value
		}
	default:
		// Store other numbered tunnel properties in the Tunnel map
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel[property] = parseValue(value)
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
		sb.WriteString(fmt.Sprintf("option.i2cp.%s=%s\n", k, formatPropertyValue(v)))
	}

	for k, v := range config.Tunnel {
		// Handle special flat properties that should not have option.i2ptunnel prefix
		switch k {
		case "proxyList", "sharedClient", "startOnLoad", "accessList", "spoofedHost", "targetPort":
			sb.WriteString(fmt.Sprintf("%s=%s\n", k, formatPropertyValue(v)))
		default:
			// Other tunnel options use the option.i2ptunnel prefix
			sb.WriteString(fmt.Sprintf("option.i2ptunnel.%s=%s\n", k, formatPropertyValue(v)))
		}
	}

	for k, v := range config.Inbound {
		sb.WriteString(fmt.Sprintf("option.inbound.%s=%s\n", k, formatPropertyValue(v)))
	}

	for k, v := range config.Outbound {
		sb.WriteString(fmt.Sprintf("option.outbound.%s=%s\n", k, formatPropertyValue(v)))
	}

	return []byte(sb.String()), nil
}

// formatPropertyValue formats a property value for output
// Arrays/slices are formatted as comma-separated values
func formatPropertyValue(v interface{}) string {
	if slice, ok := v.([]string); ok {
		return strings.Join(slice, ",")
	}
	return fmt.Sprint(v)
}
