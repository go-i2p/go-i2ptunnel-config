package i2pconv

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *Converter) parseINI(input []byte) (*TunnelConfig, error) {
	config := &TunnelConfig{
		I2CP:     make(map[string]interface{}),
		Tunnel:   make(map[string]interface{}),
		Inbound:  make(map[string]interface{}),
		Outbound: make(map[string]interface{}),
	}

	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				config.Name = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			} else {
				continue
			}
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "keys":
			config.PersistentKey = true
		case "type":
			config.Type = value
		case "host":
			config.Interface = value
		case "port":
			port, err := strconv.Atoi(value)
			if err == nil {
				config.Port = port
			}
		case "destination":
			config.Target = value
		default:
			if strings.HasPrefix(key, "i2cp.") {
				config.I2CP[strings.TrimPrefix(key, "i2cp.")] = value
			} else if strings.HasPrefix(key, "inbound") {
				config.Inbound[strings.TrimPrefix(key, "inbound.")] = value
			} else if strings.HasPrefix(key, "outbound") {
				config.Outbound[strings.TrimPrefix(key, "outbound.")] = value
			} else {
				config.Tunnel[key] = value
			}
		}
	}

	return config, nil
}

func (c *Converter) generateINI(config *TunnelConfig) ([]byte, error) {
	var sb strings.Builder

	if config.Name != "" {
		sb.WriteString(fmt.Sprintf("name = %s\n", config.Name))
	}
	if config.Type != "" {
		sb.WriteString(fmt.Sprintf("type = %s\n", config.Type))
	}
	if config.Interface != "" {
		sb.WriteString(fmt.Sprintf("interface = %s\n", config.Interface))
	}
	if config.Port != 0 {
		sb.WriteString(fmt.Sprintf("port = %d\n", config.Port))
	}

	if config.PersistentKey {
		sb.WriteString(fmt.Sprintf("keys = %s\n", strings.TrimSpace(strings.ReplaceAll(config.Name, " ", "_"))))
	}
	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("description = %s\n", config.Description))
	}

	for k, v := range config.I2CP {
		sb.WriteString(fmt.Sprintf("i2cp.%s = %v\n", k, v))
	}

	for k, v := range config.Tunnel {
		sb.WriteString(fmt.Sprintf("%s = %v\n", k, v))
	}

	for k, v := range config.Inbound {
		sb.WriteString(fmt.Sprintf("inbound.%s = %v\n", k, v))
	}

	for k, v := range config.Outbound {
		sb.WriteString(fmt.Sprintf("outbound.%s = %v\n", k, v))
	}

	return []byte(sb.String()), nil
}
