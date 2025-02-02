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
	if strings.HasPrefix(k, "#") {
		return
	}
	kv := strings.Split(k, "=")
	if len(kv) != 2 {
		return
	}
	parts := strings.Split(kv[0], ".")

	switch parts[1] {
	case "name":
		config.Name = s
	case "type":
		config.Type = s
	case "interface":
		config.Interface = s
	case "listenPort":
		port, err := strconv.Atoi(s)
		if err == nil {
			config.Port = port
		}
	case "option.persistentClientKey":
		config.PersistentKey = true
	case "description":
		config.Description = s
	default:
		if strings.HasPrefix(parts[1], "option.i2cp") {
			config.I2CP[parts[2]] = s
		} else if strings.HasPrefix(parts[1], "option.i2ptunnel") {
			config.Tunnel[parts[2]] = s
		} else if strings.HasPrefix(parts[1], "option.inbound") {
			config.Inbound[parts[2]] = s
		} else if strings.HasPrefix(parts[1], "option.outbound") {
			config.Outbound[parts[2]] = s
		}
	}
}

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
