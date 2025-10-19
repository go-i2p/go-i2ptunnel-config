package i2pconv

import (
	"reflect"
	"testing"
)

func TestEnhancedINIParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *TunnelConfig
	}{
		{
			name: "i2pd INI section format",
			input: `[HttpProxy]
type = httpclient
host = 127.0.0.1
port = 4444
destination = example.i2p
keys = proxy.dat
`,
			expected: &TunnelConfig{
				Name:          "HttpProxy",
				Type:          "httpclient",
				Interface:     "127.0.0.1",
				Port:          4444,
				Target:        "example.i2p",
				PersistentKey: true,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"keyfile": "proxy.dat",
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "i2pd server tunnel with address",
			input: `[WebServer]
type = httpserver
host = 0.0.0.0
port = 80
address = webserver.i2p
keys = webserver.dat
gzip = true
`,
			expected: &TunnelConfig{
				Name:          "WebServer",
				Type:          "httpserver",
				Interface:     "0.0.0.0",
				Port:          80,
				Target:        "webserver.i2p",
				PersistentKey: true,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"keyfile": "webserver.dat",
					"gzip":    true,
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "i2pd boolean and numeric conversion",
			input: `[TestTunnel]
type = client
gzip = yes
multicast = no
signaturetype = 7
crypto.tagsToSend = 40
streamr.rto = 120
`,
			expected: &TunnelConfig{
				Name:          "TestTunnel",
				Type:          "client",
				PersistentKey: false,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"gzip":              true,
					"multicast":         false,
					"signaturetype":     7,
					"crypto.tagsToSend": 40,
					"streamr.rto":       120,
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "i2pd access list and explicit peers",
			input: `[RestrictedTunnel]
type = httpclient
accesslist = peer1.i2p, peer2.i2p, peer3.i2p
explicitpeers = explicit1.i2p,explicit2.i2p
hostoverride = override.example.com
`,
			expected: &TunnelConfig{
				Name:          "RestrictedTunnel",
				Type:          "httpclient",
				PersistentKey: false,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"accesslist":    []string{"peer1.i2p", "peer2.i2p", "peer3.i2p"},
					"explicitpeers": []string{"explicit1.i2p", "explicit2.i2p"},
					"hostoverride":  "override.example.com",
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "transient keys",
			input: `[TransientTunnel]
type = client
keys = transient
maptoloopback = 1
enableuniquelocal = 0
`,
			expected: &TunnelConfig{
				Name:          "TransientTunnel",
				Type:          "client",
				PersistentKey: false,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"maptoloopback":     true,
					"enableuniquelocal": false,
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "i2cp and inbound/outbound options",
			input: `[AdvancedTunnel]
type = client
i2cp.leaseSetEncType = 4,0
i2cp.reduceIdleTime = 900000
inbound.length = 3
inbound.lengthVariance = 1
outbound.length = 2
outbound.lengthVariance = 0
`,
			expected: &TunnelConfig{
				Name:          "AdvancedTunnel",
				Type:          "client",
				PersistentKey: false,
				I2CP: map[string]interface{}{
					"leaseSetEncType": []string{"4", "0"},
					"reduceIdleTime":  900000,
				},
				Tunnel: make(map[string]interface{}),
				Inbound: map[string]interface{}{
					"length":         3,
					"lengthVariance": 1,
				},
				Outbound: map[string]interface{}{
					"length":         2,
					"lengthVariance": 0,
				},
			},
		},
	}

	conv := &Converter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := conv.ParseInput([]byte(tt.input), "ini")
			if err != nil {
				t.Fatalf("Failed to parse INI: %v", err)
			}

			// Check basic fields
			if config.Name != tt.expected.Name {
				t.Errorf("Name: expected %q, got %q", tt.expected.Name, config.Name)
			}
			if config.Type != tt.expected.Type {
				t.Errorf("Type: expected %q, got %q", tt.expected.Type, config.Type)
			}
			if config.Interface != tt.expected.Interface {
				t.Errorf("Interface: expected %q, got %q", tt.expected.Interface, config.Interface)
			}
			if config.Port != tt.expected.Port {
				t.Errorf("Port: expected %d, got %d", tt.expected.Port, config.Port)
			}
			if config.Target != tt.expected.Target {
				t.Errorf("Target: expected %q, got %q", tt.expected.Target, config.Target)
			}
			if config.PersistentKey != tt.expected.PersistentKey {
				t.Errorf("PersistentKey: expected %t, got %t", tt.expected.PersistentKey, config.PersistentKey)
			}

			// Check map fields using proper comparison
			for k, v := range tt.expected.I2CP {
				if actualV, ok := config.I2CP[k]; !ok {
					t.Errorf("I2CP[%q]: missing key", k)
				} else if !iniValuesEqual(v, actualV) {
					t.Errorf("I2CP[%q]: expected %v (%T), got %v (%T)", k, v, v, actualV, actualV)
				}
			}
			for k, v := range tt.expected.Tunnel {
				if actualV, ok := config.Tunnel[k]; !ok {
					t.Errorf("Tunnel[%q]: missing key", k)
				} else if !iniValuesEqual(v, actualV) {
					t.Errorf("Tunnel[%q]: expected %v (%T), got %v (%T)", k, v, v, actualV, actualV)
				}
			}
			for k, v := range tt.expected.Inbound {
				if actualV, ok := config.Inbound[k]; !ok {
					t.Errorf("Inbound[%q]: missing key", k)
				} else if !iniValuesEqual(v, actualV) {
					t.Errorf("Inbound[%q]: expected %v (%T), got %v (%T)", k, v, v, actualV, actualV)
				}
			}
			for k, v := range tt.expected.Outbound {
				if actualV, ok := config.Outbound[k]; !ok {
					t.Errorf("Outbound[%q]: missing key", k)
				} else if !iniValuesEqual(v, actualV) {
					t.Errorf("Outbound[%q]: expected %v (%T), got %v (%T)", k, v, v, actualV, actualV)
				}
			}
		})
	}
}

func TestINIRoundTrip(t *testing.T) {
	// Test that parsing and then generating produces consistent results
	input := `[RoundTripTest]
type = httpclient
host = 192.168.1.100
port = 8080
destination = example.i2p
keys = test.dat
description = Test tunnel for round-trip conversion
gzip = true
accesslist = peer1.i2p, peer2.i2p
multicast = false
signaturetype = 7
i2cp.leaseSetEncType = 4,0
i2cp.reduceIdleTime = 900000
crypto.tagsToSend = 40
inbound.length = 3
outbound.length = 2`

	conv := &Converter{}

	// Parse INI
	config, err := conv.ParseInput([]byte(input), "ini")
	if err != nil {
		t.Fatalf("Failed to parse INI: %v", err)
	}

	// Generate INI back
	output, err := conv.generateINI(config)
	if err != nil {
		t.Fatalf("Failed to generate INI: %v", err)
	}

	// Parse the generated output again
	config2, err := conv.ParseInput(output, "ini")
	if err != nil {
		t.Fatalf("Failed to re-parse generated INI: %v", err)
	}

	// Compare the two configs
	if config.Name != config2.Name {
		t.Errorf("Round-trip Name: expected %q, got %q", config.Name, config2.Name)
	}
	if config.Type != config2.Type {
		t.Errorf("Round-trip Type: expected %q, got %q", config.Type, config2.Type)
	}
	if config.Interface != config2.Interface {
		t.Errorf("Round-trip Interface: expected %q, got %q", config.Interface, config2.Interface)
	}
	if config.Port != config2.Port {
		t.Errorf("Round-trip Port: expected %d, got %d", config.Port, config2.Port)
	}
	if config.Target != config2.Target {
		t.Errorf("Round-trip Target: expected %q, got %q", config.Target, config2.Target)
	}
	if config.PersistentKey != config2.PersistentKey {
		t.Errorf("Round-trip PersistentKey: expected %t, got %t", config.PersistentKey, config2.PersistentKey)
	}

	// Check that all options were preserved
	for k, v := range config.I2CP {
		if v2, ok := config2.I2CP[k]; !ok {
			t.Errorf("Round-trip I2CP[%q]: missing key", k)
		} else if !iniValuesEqual(v, v2) {
			t.Errorf("Round-trip I2CP[%q]: expected %v, got %v", k, v, v2)
		}
	}

	t.Logf("Round-trip successful. Generated INI:\n%s", string(output))
}

func TestINIEdgeCases(t *testing.T) {
	conv := &Converter{}

	// Test comments and empty lines are ignored
	input := `# This is a comment
; This is also a comment

[TestTunnel]
type = client
# Another comment
host = 127.0.0.1

; More comments
port = 1234`

	config, err := conv.ParseInput([]byte(input), "ini")
	if err != nil {
		t.Fatalf("Failed to parse INI with comments: %v", err)
	}

	if config.Name != "TestTunnel" {
		t.Errorf("Expected name 'TestTunnel', got %q", config.Name)
	}
	if config.Type != "client" {
		t.Errorf("Expected type 'client', got %q", config.Type)
	}
	if config.Interface != "127.0.0.1" {
		t.Errorf("Expected interface '127.0.0.1', got %q", config.Interface)
	}
	if config.Port != 1234 {
		t.Errorf("Expected port 1234, got %d", config.Port)
	}

	// Test boolean variations
	boolInput := `[BoolTest]
type = client
option1 = true
option2 = false
option3 = yes
option4 = no
option5 = 1
option6 = 0
option7 = on
option8 = off
option9 = enabled
option10 = disabled`

	config2, err := conv.ParseInput([]byte(boolInput), "ini")
	if err != nil {
		t.Fatalf("Failed to parse boolean INI: %v", err)
	}

	expectedValues := map[string]interface{}{
		"option1":  true,  // "true"
		"option2":  false, // "false"
		"option3":  true,  // "yes"
		"option4":  false, // "no"
		"option5":  1,     // "1" - parsed as integer
		"option6":  0,     // "0" - parsed as integer
		"option7":  true,  // "on"
		"option8":  false, // "off"
		"option9":  true,  // "enabled"
		"option10": false, // "disabled"
	}

	for k, expected := range expectedValues {
		if actual, ok := config2.Tunnel[k]; !ok {
			t.Errorf("Option %q missing", k)
		} else if actual != expected {
			t.Errorf("Option %q: expected %v (%T), got %v (%T)", k, expected, expected, actual, actual)
		}
	}
}

// iniValuesEqual compares two values, handling slices and type differences properly
func iniValuesEqual(expected, actual interface{}) bool {
	return reflect.DeepEqual(expected, actual)
}
