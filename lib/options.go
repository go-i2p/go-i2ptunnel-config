package i2pconv

import (
	"fmt"
	"strconv"
)

// Return the optins as a map of key-value pairs.
func (t *TunnelConfig) Options() map[string]string {
	options := make(map[string]string)
	options["Name"] = t.Name
	options["Type"] = t.Type
	if t.Interface != "" {
		options["Interface"] = t.Interface
	}
	if t.Port != 0 {
		options["Port"] = strconv.Itoa(t.Port)
	}
	if t.Target != "" {
		options["Target"] = t.Target
	}
	if t.PersistentKey {
		options["PersistentKey"] = "true"
	}
	if t.Description != "" {
		options["Description"] = t.Description
	}
	for k, v := range t.I2CP {
		options["I2CP."+k] = fmt.Sprintf("%v", v)
	}
	for k, v := range t.Tunnel {
		options["Tunnel."+k] = fmt.Sprintf("%v", v)
	}
	for k, v := range t.Inbound {
		options["Inbound."+k] = fmt.Sprintf("%v", v)
	}
	for k, v := range t.Outbound {
		options["Outbound."+k] = fmt.Sprintf("%v", v)
	}
	return options
}

// SetOptions sets the options from a map of key-value pairs.
func (t *TunnelConfig) SetOptions(options map[string]string) error {
	for k, v := range options {
		switch k {
		case "Name":
			t.Name = v
		case "Type":
			t.Type = v
		case "Interface":
			t.Interface = v
		case "Port":
			port, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("invalid port: %v", err)
			}
			t.Port = port
		case "Target":
			t.Target = v
		case "PersistentKey":
			persistentKey, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid persistent key: %v", err)
			}
			t.PersistentKey = persistentKey
		case "Description":
			t.Description = v
		default:
			if len(k) > 5 && k[:5] == "I2CP." {
				if t.I2CP == nil {
					t.I2CP = make(map[string]interface{})
				}
				t.I2CP[k[5:]] = v
				continue
			}
			if len(k) > 7 && k[:7] == "Tunnel." {
				if t.Tunnel == nil {
					t.Tunnel = make(map[string]interface{})
				}
				t.Tunnel[k[7:]] = v
				continue
			}
			if len(k) > 8 && k[:8] == "Inbound." {
				if t.Inbound == nil {
					t.Inbound = make(map[string]interface{})
				}
				t.Inbound[k[8:]] = v
				continue
			}
			if len(k) > 9 && k[:9] == "Outbound." {
				if t.Outbound == nil {
					t.Outbound = make(map[string]interface{})
				}
				t.Outbound[k[9:]] = v
				continue
			}
		}
	}
	return nil
}
