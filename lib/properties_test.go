package i2pconv

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPropertyConversion(t *testing.T) {
	input := `
name=I2P HTTP Proxy
type=httpclient
interface=127.0.0.1
listenPort=4444
description=HTTP proxy for browsing eepsites
option.i2cp.leaseSetEncType=4,0
option.i2cp.reduceIdleTime=900000
option.inbound.length=3
proxyList=exit.stormycloud.i2p
sharedClient=true
`
	conv := &Converter{}
	config, err := conv.parseJavaProperties([]byte(input))
	if err != nil {
		t.Fatalf("Failed to parse properties: %v", err)
	}

	yaml, err := conv.generateYAML(config)
	if err != nil {
		t.Fatalf("Failed to generate YAML: %v", err)
	}

	// Expected YAML structure matching enhanced parsing:
	expected := `tunnels:
  I2P HTTP Proxy:
    name: I2P HTTP Proxy
    type: httpclient
    interface: 127.0.0.1
    port: 4444
    description: HTTP proxy for browsing eepsites
    i2cp:
      leaseSetEncType:
      - "4"
      - "0"
      reduceIdleTime: 900000
    options:
      proxyList: exit.stormycloud.i2p
      sharedClient: true
    inbound:
      length: 3
`
	// Compare YAML output
	if string(yaml) != expected {
		t.Fatalf("Unexpected YAML output:\n%s\nExpected:\n%s", yaml, expected)
	} else {
		t.Logf("YAML output:\n%s", yaml)
	}
}

func TestEnhancedPropertiesParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *TunnelConfig
	}{
		{
			name: "numbered tunnel properties",
			input: `
tunnel.0.name=HTTPProxy
tunnel.0.type=httpclient
tunnel.0.interface=127.0.0.1
tunnel.0.listenPort=4444
tunnel.0.description=HTTP proxy tunnel
`,
			expected: &TunnelConfig{
				Name:        "HTTPProxy",
				Type:        "httpclient",
				Interface:   "127.0.0.1",
				Port:        4444,
				Description: "HTTP proxy tunnel",
				I2CP:        make(map[string]interface{}),
				Tunnel:      make(map[string]interface{}),
				Inbound:     make(map[string]interface{}),
				Outbound:    make(map[string]interface{}),
			},
		},
		{
			name: "option prefixes",
			input: `
name=TestTunnel
type=server
option.i2ptunnel.useCompression=true
option.inbound.length=3
option.outbound.length=2
option.persistentClientKey=true
`,
			expected: &TunnelConfig{
				Name:          "TestTunnel",
				Type:          "server",
				PersistentKey: true,
				I2CP:          make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"useCompression": true,
				},
				Inbound: map[string]interface{}{
					"length": 3,
				},
				Outbound: map[string]interface{}{
					"length": 2,
				},
			},
		},
		{
			name: "additional flat properties",
			input: `
name=WebServer
type=httpserver
proxyList=example.i2p,another.i2p
sharedClient=false
startOnLoad=true
accessList=allow
targetPort=8080
spoofedHost=example.com
`,
			expected: &TunnelConfig{
				Name: "WebServer",
				Type: "httpserver",
				I2CP: make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"proxyList":    []string{"example.i2p", "another.i2p"},
					"sharedClient": false,
					"startOnLoad":  true,
					"accessList":   "allow",
					"targetPort":   8080,
					"spoofedHost":  "example.com",
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "mixed numbered and flat properties",
			input: `
tunnel.0.name=MixedTunnel
tunnel.1.type=client
interface=192.168.1.1
listenPort=7070
proxyList=proxy.i2p
option.i2cp.reduceIdleTime=600000
`,
			expected: &TunnelConfig{
				Name:      "MixedTunnel",
				Type:      "client",
				Interface: "192.168.1.1",
				Port:      7070,
				I2CP: map[string]interface{}{
					"reduceIdleTime": 600000,
				},
				Tunnel: map[string]interface{}{
					"proxyList": "proxy.i2p",
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
	}

	conv := &Converter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := conv.parseJavaProperties([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse properties: %v", err)
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
			if config.Description != tt.expected.Description {
				t.Errorf("Description: expected %q, got %q", tt.expected.Description, config.Description)
			}
			if config.PersistentKey != tt.expected.PersistentKey {
				t.Errorf("PersistentKey: expected %t, got %t", tt.expected.PersistentKey, config.PersistentKey)
			}

			// Check map fields using proper comparison
			for k, v := range tt.expected.I2CP {
				if actualV, ok := config.I2CP[k]; !ok {
					t.Errorf("I2CP[%q]: missing key", k)
				} else if !valuesEqual(v, actualV) {
					t.Errorf("I2CP[%q]: expected %v, got %v", k, v, actualV)
				}
			}
			for k, v := range tt.expected.Tunnel {
				if actualV, ok := config.Tunnel[k]; !ok {
					t.Errorf("Tunnel[%q]: missing key", k)
				} else if !valuesEqual(v, actualV) {
					t.Errorf("Tunnel[%q]: expected %v, got %v", k, v, actualV)
				}
			}
			for k, v := range tt.expected.Inbound {
				if actualV, ok := config.Inbound[k]; !ok {
					t.Errorf("Inbound[%q]: missing key", k)
				} else if !valuesEqual(v, actualV) {
					t.Errorf("Inbound[%q]: expected %v, got %v", k, v, actualV)
				}
			}
			for k, v := range tt.expected.Outbound {
				if actualV, ok := config.Outbound[k]; !ok {
					t.Errorf("Outbound[%q]: missing key", k)
				} else if !valuesEqual(v, actualV) {
					t.Errorf("Outbound[%q]: expected %v, got %v", k, v, actualV)
				}
			}
		})
	}
}

func TestPropertiesRoundTrip(t *testing.T) {
	// Test that parsing and then generating produces consistent results
	input := `name=RoundTripTest
type=httpclient
interface=127.0.0.1
listenPort=8888
targetDestination=example.i2p
description=Test tunnel for round-trip conversion
proxyList=proxy1.i2p,proxy2.i2p
sharedClient=true
startOnLoad=false
option.i2cp.leaseSetEncType=4,0
option.i2cp.reduceIdleTime=900000
option.i2ptunnel.useCompression=true
option.inbound.length=3
option.outbound.length=2
option.persistentClientKey=true`

	conv := &Converter{}

	// Parse properties
	config, err := conv.parseJavaProperties([]byte(input))
	if err != nil {
		t.Fatalf("Failed to parse properties: %v", err)
	}

	// Generate properties back
	output, err := conv.generateJavaProperties(config)
	if err != nil {
		t.Fatalf("Failed to generate properties: %v", err)
	}

	// Parse the generated output again
	config2, err := conv.parseJavaProperties(output)
	if err != nil {
		t.Fatalf("Failed to re-parse generated properties: %v", err)
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
	if config.Description != config2.Description {
		t.Errorf("Round-trip Description: expected %q, got %q", config.Description, config2.Description)
	}
	if config.PersistentKey != config2.PersistentKey {
		t.Errorf("Round-trip PersistentKey: expected %t, got %t", config.PersistentKey, config2.PersistentKey)
	}

	// Check that all options were preserved
	for k, v := range config.I2CP {
		if v2, ok := config2.I2CP[k]; !ok || fmt.Sprint(v) != fmt.Sprint(v2) {
			t.Errorf("Round-trip I2CP[%q]: expected %v, got %v", k, v, v2)
		}
	}

	t.Logf("Round-trip successful. Generated properties:\n%s", string(output))
}

func TestPropertiesEdgeCases(t *testing.T) {
	conv := &Converter{}

	// Test comments and config file paths are ignored
	input := `# This is a comment
configFile=/path/to/file
name=TestTunnel
# Another comment
type=client`

	config, err := conv.ParseInput([]byte(input), "properties")
	if err != nil {
		t.Fatalf("Failed to parse properties: %v", err)
	}

	if config.Name != "TestTunnel" {
		t.Errorf("Expected name 'TestTunnel', got %q", config.Name)
	}
	if config.Type != "client" {
		t.Errorf("Expected type 'client', got %q", config.Type)
	}

	// Test alternate tunnel destination naming
	input2 := `name=AltTunnel
type=server
targetHost=alt.example.i2p
i2cpHost=127.0.0.1
i2cpPort=7654`

	config2, err := conv.ParseInput([]byte(input2), "properties")
	if err != nil {
		t.Fatalf("Failed to parse properties: %v", err)
	}

	if config2.Target != "alt.example.i2p" {
		t.Errorf("Expected target 'alt.example.i2p', got %q", config2.Target)
	}
	if config2.I2CP["host"] != "127.0.0.1" {
		t.Errorf("Expected I2CP host '127.0.0.1', got %v", config2.I2CP["host"])
	}
	if config2.I2CP["port"] != 7654 {
		t.Errorf("Expected I2CP port 7654, got %v", config2.I2CP["port"])
	}

	// Test numbered tunnel conflicts (second values should be ignored when flat values exist)
	input3 := `name=FirstName
tunnel.0.name=SecondName
type=client
tunnel.1.type=server`

	config3, err := conv.ParseInput([]byte(input3), "properties")
	if err != nil {
		t.Fatalf("Failed to parse properties: %v", err)
	}

	// Should keep the first flat value
	if config3.Name != "FirstName" {
		t.Errorf("Expected name 'FirstName', got %q", config3.Name)
	}
	if config3.Type != "client" {
		t.Errorf("Expected type 'client', got %q", config3.Type)
	}
}

// valuesEqual compares two values, handling slices properly
func valuesEqual(expected, actual interface{}) bool {
	return reflect.DeepEqual(expected, actual)
}
