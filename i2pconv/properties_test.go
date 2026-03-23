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
		{
			name: "full tunnel.0 block with targetDestination, targetPort, and custom field",
			input: `
tunnel.0.name=FullServer
tunnel.0.type=server
tunnel.0.interface=0.0.0.0
tunnel.0.listenPort=9000
tunnel.0.targetDestination=example.b32.i2p
tunnel.0.targetPort=8080
tunnel.0.description=Full server tunnel
tunnel.0.customkey=customvalue
`,
			expected: &TunnelConfig{
				Name:        "FullServer",
				Type:        "server",
				Interface:   "0.0.0.0",
				Port:        9000,
				Target:      "example.b32.i2p",
				Description: "Full server tunnel",
				I2CP:        make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"targetPort": 8080,
					"customkey":  "customvalue",
				},
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "targetHost sets Target when empty",
			input: `
tunnel.0.name=HostTunnel
tunnel.0.type=client
tunnel.0.targetHost=192.168.1.50
`,
			expected: &TunnelConfig{
				Name:     "HostTunnel",
				Type:     "client",
				Target:   "192.168.1.50",
				I2CP:     make(map[string]interface{}),
				Tunnel:   make(map[string]interface{}),
				Inbound:  make(map[string]interface{}),
				Outbound: make(map[string]interface{}),
			},
		},
		{
			name: "alternateName stored when config.Name already set",
			input: `
name=PrimaryName
tunnel.0.type=client
tunnel.1.name=AltName
`,
			expected: &TunnelConfig{
				Name: "PrimaryName",
				Type: "client",
				I2CP: make(map[string]interface{}),
				Tunnel: map[string]interface{}{
					"alternateName": "AltName",
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

// TestParseNumberedTunnelPropertyNilMaps calls parseNumberedTunnelProperty directly
// with nil Tunnel maps to cover the nil-init guard branches that are unreachable
// through parseJavaProperties (which always pre-initialises all maps).
func TestParseNumberedTunnelPropertyNilMaps(t *testing.T) {
	conv := &Converter{}

	nilTunnel := func() *TunnelConfig {
		return &TunnelConfig{
			I2CP:     make(map[string]interface{}),
			Inbound:  make(map[string]interface{}),
			Outbound: make(map[string]interface{}),
			// Tunnel intentionally nil
		}
	}

	t.Run("targetPort with nil Tunnel initialises map", func(t *testing.T) {
		c := nilTunnel()
		conv.parseNumberedTunnelProperty("targetPort", "9090", c)
		if got, ok := c.Tunnel["targetPort"]; !ok || got != 9090 {
			t.Errorf("expected Tunnel[targetPort]=9090, got %v", got)
		}
	})

	t.Run("default property with nil Tunnel initialises map", func(t *testing.T) {
		c := nilTunnel()
		conv.parseNumberedTunnelProperty("customoption", "customvalue", c)
		if got, ok := c.Tunnel["customoption"]; !ok || got != "customvalue" {
			t.Errorf("expected Tunnel[customoption]='customvalue', got %v", got)
		}
	})

	t.Run("name alternateName with nil Tunnel initialises map", func(t *testing.T) {
		c := nilTunnel()
		c.Name = "Primary"
		conv.parseNumberedTunnelProperty("name", "Alternate", c)
		if got, ok := c.Tunnel["alternateName"]; !ok || got != "Alternate" {
			t.Errorf("expected Tunnel[alternateName]='Alternate', got %v", got)
		}
	})
}

// TestParseFlatPropertyKey exercises every branch of the parseFlatPropertyKey helper.
func TestParseFlatPropertyKey(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		check   func(t *testing.T, c *TunnelConfig)
		handled bool
	}{
		{
			key: "name", value: "myTunnel",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "name", c.Name, "myTunnel") },
			handled: true,
		},
		{
			key: "type", value: "httpclient",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "type", c.Type, "httpclient") },
			handled: true,
		},
		{
			key: "interface", value: "127.0.0.1",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "interface", c.Interface, "127.0.0.1") },
			handled: true,
		},
		{
			key: "listenPort", value: "4444",
			check:   func(t *testing.T, c *TunnelConfig) { assertIntEqual(t, "port", c.Port, 4444) },
			handled: true,
		},
		{
			key: "targetDestination", value: "example.i2p",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "target", c.Target, "example.i2p") },
			handled: true,
		},
		{
			key: "targetHost", value: "other.i2p",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "target", c.Target, "other.i2p") },
			handled: true,
		},
		{
			key: "targetPort", value: "8080",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["targetPort"]; !ok || got != 8080 {
					t.Errorf("Tunnel[targetPort]: got %v, want 8080", got)
				}
			},
			handled: true,
		},
		{
			key: "description", value: "a tunnel",
			check:   func(t *testing.T, c *TunnelConfig) { assertEqual(t, "description", c.Description, "a tunnel") },
			handled: true,
		},
		{
			key: "proxyList", value: "exit.stormycloud.i2p",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["proxyList"]; !ok || got != "exit.stormycloud.i2p" {
					t.Errorf("Tunnel[proxyList]: got %v, want exit.stormycloud.i2p", got)
				}
			},
			handled: true,
		},
		{
			key: "sharedClient", value: "true",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["sharedClient"]; !ok || got != true {
					t.Errorf("Tunnel[sharedClient]: got %v, want true", got)
				}
			},
			handled: true,
		},
		{
			key: "startOnLoad", value: "true",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["startOnLoad"]; !ok || got != true {
					t.Errorf("Tunnel[startOnLoad]: got %v, want true", got)
				}
			},
			handled: true,
		},
		{
			key: "accessList", value: "host1.i2p,host2.i2p",
			check: func(t *testing.T, c *TunnelConfig) {
				if _, ok := c.Tunnel["accessList"]; !ok {
					t.Error("Tunnel[accessList] not set")
				}
			},
			handled: true,
		},
		{
			key: "spoofedHost", value: "spoof.i2p",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["spoofedHost"]; !ok || got != "spoof.i2p" {
					t.Errorf("Tunnel[spoofedHost]: got %v, want spoof.i2p", got)
				}
			},
			handled: true,
		},
		{
			key: "i2cpHost", value: "127.0.0.1",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.I2CP["host"]; !ok || got != "127.0.0.1" {
					t.Errorf("I2CP[host]: got %v, want 127.0.0.1", got)
				}
			},
			handled: true,
		},
		{
			key: "i2cpPort", value: "7654",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.I2CP["port"]; !ok || got != 7654 {
					t.Errorf("I2CP[port]: got %v, want 7654", got)
				}
			},
			handled: true,
		},
		{
			key:     "unknownKey",
			value:   "x",
			check:   func(t *testing.T, c *TunnelConfig) {},
			handled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			c := &TunnelConfig{}
			got := parseFlatPropertyKey(tt.key, tt.value, c)
			if got != tt.handled {
				t.Errorf("parseFlatPropertyKey(%q) handled=%v, want %v", tt.key, got, tt.handled)
			}
			tt.check(t, c)
		})
	}
}

// TestParsePrefixedPropertyKey exercises every branch of the parsePrefixedPropertyKey helper.
func TestParsePrefixedPropertyKey(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		check   func(t *testing.T, c *TunnelConfig)
		handled bool
	}{
		{
			key: "option.i2cp.leaseSetEncType", value: "4",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.I2CP["leaseSetEncType"]; !ok || got != 4 {
					t.Errorf("I2CP[leaseSetEncType]: got %v, want 4", got)
				}
			},
			handled: true,
		},
		{
			key: "option.i2ptunnel.maxParticipants", value: "10",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Tunnel["maxParticipants"]; !ok || got != 10 {
					t.Errorf("Tunnel[maxParticipants]: got %v, want 10", got)
				}
			},
			handled: true,
		},
		{
			key: "option.inbound.length", value: "3",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Inbound["length"]; !ok || got != 3 {
					t.Errorf("Inbound[length]: got %v, want 3", got)
				}
			},
			handled: true,
		},
		{
			key: "option.outbound.length", value: "2",
			check: func(t *testing.T, c *TunnelConfig) {
				if got, ok := c.Outbound["length"]; !ok || got != 2 {
					t.Errorf("Outbound[length]: got %v, want 2", got)
				}
			},
			handled: true,
		},
		{
			key: "option.persistentClientKey", value: "true",
			check: func(t *testing.T, c *TunnelConfig) {
				if !c.PersistentKey {
					t.Error("PersistentKey should be true")
				}
			},
			handled: true,
		},
		{
			key:     "unknownPrefix",
			value:   "x",
			check:   func(t *testing.T, c *TunnelConfig) {},
			handled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			c := &TunnelConfig{}
			got := parsePrefixedPropertyKey(tt.key, tt.value, c)
			if got != tt.handled {
				t.Errorf("parsePrefixedPropertyKey(%q) handled=%v, want %v", tt.key, got, tt.handled)
			}
			tt.check(t, c)
		})
	}
}

// assertEqual is a tiny local helper for string comparison used by TestParseFlatPropertyKey.
func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

// assertIntEqual is a tiny local helper for int comparison.
func assertIntEqual(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", field, got, want)
	}
}

// TestCountPropertiesTunnels verifies the new countPropertiesTunnels helper.
func TestCountPropertiesTunnels(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "single tunnel via numbered keys",
			input: "tunnel.0.name=A\ntunnel.0.type=httpclient\n",
			want:  1,
		},
		{
			name:  "two tunnels via numbered keys",
			input: "tunnel.0.name=A\ntunnel.0.listenPort=4444\ntunnel.1.name=B\ntunnel.1.listenPort=5555\n",
			want:  2,
		},
		{
			name:  "three tunnels via numbered keys",
			input: "tunnel.0.name=A\ntunnel.1.name=B\ntunnel.2.name=C\n",
			want:  3,
		},
		{
			name:  "flat keys only treated as single tunnel",
			input: "name=myTunnel\ntype=httpclient\n",
			want:  1,
		},
		{
			name:  "empty input treated as single tunnel",
			input: "",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countPropertiesTunnels([]byte(tt.input))
			if got != tt.want {
				t.Errorf("countPropertiesTunnels: got %d, want %d", got, tt.want)
			}
		})
	}
}
