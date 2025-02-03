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
