package i2pconv

import (
	"fmt"
	"strconv"
	"strings"
)

// parseINIValue converts an INI value string to appropriate Go type
// Similar to parseValue in properties.go but adapted for i2pd conventions
func parseINIValue(s string) interface{} {
	original := strings.TrimSpace(s)
	lower := strings.ToLower(original)

	// Handle i2pd boolean keywords first (not numeric)
	switch lower {
	case "true", "yes", "on", "enabled":
		return true
	case "false", "no", "off", "disabled":
		return false
	}

	// Try integer (including 1/0 which could be boolean or numeric)
	if i, err := strconv.Atoi(original); err == nil {
		return i
	}

	// Try comma-separated list (for accesslist, explicitPeers, etc.)
	if strings.Contains(original, ",") {
		parts := strings.Split(original, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}

	// Default to original string
	return original
}

// parseINIBooleanValue specifically parses boolean values, including 1/0
func parseINIBooleanValue(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "true", "yes", "1", "on", "enabled":
		return true
	case "false", "no", "0", "off", "disabled":
		return false
	default:
		// Default to false for unknown values
		return false
	}
}

// parseINI parses i2pd INI configuration format with full compatibility.
// Supports all i2pd tunnel configuration patterns:
//
// INI Sections:
//   - [tunnel-name] sections for named tunnels
//   - Multiple sections supported (this converter uses first/primary)
//
// Core Properties:
//   - type, host/interface, port, destination/address, keys, description
//   - Automatic server/client destination mapping (address vs destination)
//
// i2pd-Specific Properties:
//   - gzip, multicast, maptoloopback, enableuniquelocal (boolean)
//   - accesslist, explicitpeers (comma-separated lists)
//   - hostoverride, webircpassword (strings)
//   - signaturetype (integer)
//
// Advanced Options:
//   - crypto.* options (e.g., crypto.tagsToSend)
//   - streamr.* options (e.g., streamr.rto)
//   - i2cp.* options with proper type conversion
//   - inbound.* and outbound.* tunnel options
//
// Key Management:
//   - keys=filename -> persistent keys with specified file
//   - keys=transient -> non-persistent keys
//
// Type Conversion:
//   - Automatic boolean detection (true/false, yes/no, on/off, enabled/disabled)
//   - Numeric values preserved as integers
//   - Comma-separated values become string arrays
//   - Context-aware parsing for known boolean properties
func (c *Converter) parseINI(input []byte) (*TunnelConfig, error) {
	config := &TunnelConfig{
		I2CP:     make(map[string]interface{}),
		Tunnel:   make(map[string]interface{}),
		Inbound:  make(map[string]interface{}),
		Outbound: make(map[string]interface{}),
	}

	lines := strings.Split(string(input), "\n")
	currentSection := ""

	for lineNum, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle INI sections [section-name]
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, newParseError(input, lineNum+1, 0, "ini",
					"unclosed section bracket - expected ']'")
			}
			currentSection = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			if currentSection == "" {
				return nil, newParseError(input, lineNum+1, 0, "ini",
					"empty section name - sections must have a name")
			}
			// For single tunnel config, use section name as tunnel name
			if config.Name == "" {
				config.Name = currentSection
			}
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Check if line looks like it should be a key=value pair but is malformed
			if strings.Contains(originalLine, "=") {
				return nil, newParseError(input, lineNum+1, 0, "ini",
					"malformed key=value pair - check for extra '=' characters")
			}
			return nil, newParseError(input, lineNum+1, 0, "ini",
				"expected key=value pair or section header [name]")
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, newParseError(input, lineNum+1, 0, "ini",
				"empty key name - key=value pairs must have a key")
		}

		// Parse key-value pair with i2pd-specific handling
		c.parseINIKeyValue(key, value, config)
	}

	return config, nil
}

// parseINIKeyValue handles individual key-value pairs with i2pd-specific mappings
func (c *Converter) parseINIKeyValue(key, value string, config *TunnelConfig) {
	switch key {
	// Core tunnel properties
	case "name":
		if config.Name == "" { // Don't override section name
			config.Name = value
		}
	case "type":
		config.Type = value
	case "host", "interface":
		config.Interface = value
	case "port":
		if port, err := strconv.Atoi(value); err == nil {
			config.Port = port
		}
	case "destination":
		config.Target = value
	case "address": // i2pd server tunnel address
		config.Target = value
	case "hostoverride": // i2pd client tunnel host override
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["hostoverride"] = value
	case "description":
		config.Description = value

	// Key management
	case "keys":
		// In i2pd, keys can be a filename or "transient"
		if strings.ToLower(value) == "transient" {
			config.PersistentKey = false
		} else {
			config.PersistentKey = true
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel["keyfile"] = value
		}

	// i2pd-specific properties
	case "gzip":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["gzip"] = parseINIBooleanValue(value)
	case "accesslist":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["accesslist"] = parseINIValue(value)
	case "signaturetype":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["signaturetype"] = parseINIValue(value)
	case "explicitpeers":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["explicitpeers"] = parseINIValue(value)
	case "multicast":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["multicast"] = parseINIBooleanValue(value)
	case "webircpassword":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["webircpassword"] = value
	case "maptoloopback":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["maptoloopback"] = parseINIBooleanValue(value)
	case "enableuniquelocal":
		if config.Tunnel == nil {
			config.Tunnel = make(map[string]interface{})
		}
		config.Tunnel["enableuniquelocal"] = parseINIBooleanValue(value)

	default:
		// Handle prefixed keys
		if strings.HasPrefix(key, "i2cp.") {
			if config.I2CP == nil {
				config.I2CP = make(map[string]interface{})
			}
			optionKey := strings.TrimPrefix(key, "i2cp.")
			config.I2CP[optionKey] = parseINIValue(value)
		} else if strings.HasPrefix(key, "crypto.") {
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel[key] = parseINIValue(value)
		} else if strings.HasPrefix(key, "streamr.") {
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel[key] = parseINIValue(value)
		} else if strings.HasPrefix(key, "inbound.") {
			if config.Inbound == nil {
				config.Inbound = make(map[string]interface{})
			}
			optionKey := strings.TrimPrefix(key, "inbound.")
			config.Inbound[optionKey] = parseINIValue(value)
		} else if strings.HasPrefix(key, "outbound.") {
			if config.Outbound == nil {
				config.Outbound = make(map[string]interface{})
			}
			optionKey := strings.TrimPrefix(key, "outbound.")
			config.Outbound[optionKey] = parseINIValue(value)
		} else {
			// Store other options in Tunnel map
			if config.Tunnel == nil {
				config.Tunnel = make(map[string]interface{})
			}
			config.Tunnel[key] = parseINIValue(value)
		}
	}
}

// generateINI generates i2pd-compatible INI configuration
// Outputs proper INI sections and i2pd-specific properties
func (c *Converter) generateINI(config *TunnelConfig) ([]byte, error) {
	var sb strings.Builder

	// Write INI section header if name is provided
	if config.Name != "" {
		sb.WriteString(fmt.Sprintf("[%s]\n", config.Name))
	}

	// Core tunnel properties
	if config.Type != "" {
		sb.WriteString(fmt.Sprintf("type = %s\n", config.Type))
	}
	if config.Interface != "" {
		sb.WriteString(fmt.Sprintf("host = %s\n", config.Interface))
	}
	if config.Port != 0 {
		sb.WriteString(fmt.Sprintf("port = %d\n", config.Port))
	}

	// Handle target based on tunnel type (destination for client, address for server)
	if config.Target != "" {
		if config.Type == "server" || config.Type == "httpserver" || config.Type == "ircserver" {
			sb.WriteString(fmt.Sprintf("address = %s\n", config.Target))
		} else {
			sb.WriteString(fmt.Sprintf("destination = %s\n", config.Target))
		}
	}

	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("description = %s\n", config.Description))
	}

	// Key management
	if config.PersistentKey {
		// Check if keyfile is specified in Tunnel options
		if keyfile, ok := config.Tunnel["keyfile"]; ok {
			sb.WriteString(fmt.Sprintf("keys = %s\n", keyfile))
		} else {
			// Generate default keyfile name
			keyName := strings.ReplaceAll(config.Name, " ", "_")
			if keyName == "" {
				keyName = "tunnel"
			}
			sb.WriteString(fmt.Sprintf("keys = %s.dat\n", keyName))
		}
	} else {
		sb.WriteString("keys = transient\n")
	}

	// I2CP options
	for k, v := range config.I2CP {
		sb.WriteString(fmt.Sprintf("i2cp.%s = %s\n", k, formatINIValue(v)))
	}

	// Tunnel options with i2pd-specific handling
	for k, v := range config.Tunnel {
		// Skip keyfile as it's handled above
		if k == "keyfile" {
			continue
		}

		// Handle special i2pd properties
		switch k {
		case "hostoverride", "gzip", "accesslist", "signaturetype", "explicitpeers",
			"multicast", "webircpassword", "maptoloopback", "enableuniquelocal":
			sb.WriteString(fmt.Sprintf("%s = %s\n", k, formatINIValue(v)))
		default:
			// Include crypto.* and streamr.* options directly
			if strings.HasPrefix(k, "crypto.") || strings.HasPrefix(k, "streamr.") {
				sb.WriteString(fmt.Sprintf("%s = %s\n", k, formatINIValue(v)))
			} else {
				// Other tunnel options
				sb.WriteString(fmt.Sprintf("%s = %s\n", k, formatINIValue(v)))
			}
		}
	}

	// Inbound/Outbound options
	for k, v := range config.Inbound {
		sb.WriteString(fmt.Sprintf("inbound.%s = %s\n", k, formatINIValue(v)))
	}

	for k, v := range config.Outbound {
		sb.WriteString(fmt.Sprintf("outbound.%s = %s\n", k, formatINIValue(v)))
	}

	return []byte(sb.String()), nil
}

// formatINIValue formats a value for INI output
// Handles arrays as comma-separated values and booleans as true/false
func formatINIValue(v interface{}) string {
	switch val := v.(type) {
	case []string:
		return strings.Join(val, ", ")
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(val)
	}
}
