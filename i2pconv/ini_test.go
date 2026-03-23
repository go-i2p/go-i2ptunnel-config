package i2pconv

import (
	"reflect"
	"strings"
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

// TestINIMultiSectionParsesFirstOnly documents that a two-section INI file takes
// the name from the first section but merges all subsequent key-value pairs into
// the same config (later values overwrite earlier ones when keys collide).
// This confirms that countINISections correctly detects multi-tunnel input and
// that the warning path is reachable.
func TestINIMultiSectionParsesFirstOnly(t *testing.T) {
	input := `[FirstTunnel]
type = httpclient
host = 127.0.0.1
port = 4444

[SecondTunnel]
type = httpserver
host = 0.0.0.0
port = 80
address = second.i2p
`
	conv := &Converter{}
	config, err := conv.ParseInput([]byte(input), "ini")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// The name is set from the first section header only.
	if config.Name != "FirstTunnel" {
		t.Errorf("expected name 'FirstTunnel', got %q", config.Name)
	}

	// The INI parser merges all sections: later keys overwrite earlier ones.
	// So type/port/address come from SecondTunnel (last section).
	if config.Type != "httpserver" {
		t.Errorf("expected merged type 'httpserver' (from SecondTunnel), got %q", config.Type)
	}
	if config.Port != 80 {
		t.Errorf("expected merged port 80 (from SecondTunnel), got %d", config.Port)
	}

	// countINISections must detect both sections to confirm the warning path is reachable.
	n := countINISections([]byte(input))
	if n != 2 {
		t.Errorf("expected countINISections to return 2, got %d", n)
	}
}

// TestINIKeyValueCoverage covers the individual parseINIKeyValue branches that are
// not reachable through the normal parseINI path (e.g., nil-map guards) as well
// as less commonly exercised cases like "webircpassword" and "name" key override.
func TestINIKeyValueCoverage(t *testing.T) {
	conv := &Converter{}

	// Helper: fresh config with nil maps to trigger nil-init guards in parseINIKeyValue.
	nilConfig := func() *TunnelConfig {
		return &TunnelConfig{
			I2CP:     make(map[string]interface{}),
			Inbound:  make(map[string]interface{}),
			Outbound: make(map[string]interface{}),
			// Tunnel intentionally nil
		}
	}

	t.Run("name key sets name when empty", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("name", "MyTunnel", c)
		if c.Name != "MyTunnel" {
			t.Errorf("expected Name 'MyTunnel', got %q", c.Name)
		}
	})

	t.Run("name key skipped when already set", func(t *testing.T) {
		c := nilConfig()
		c.Name = "SectionName"
		conv.parseINIKeyValue("name", "OverrideName", c)
		if c.Name != "SectionName" {
			t.Errorf("expected Name to remain 'SectionName', got %q", c.Name)
		}
	})

	t.Run("webircpassword is stored in Tunnel map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("webircpassword", "secret", c)
		if got, ok := c.Tunnel["webircpassword"]; !ok || got != "secret" {
			t.Errorf("expected Tunnel[webircpassword]='secret', got %v", got)
		}
	})

	t.Run("gzip with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("gzip", "true", c)
		if got, ok := c.Tunnel["gzip"]; !ok || got != true {
			t.Errorf("expected Tunnel[gzip]=true, got %v", got)
		}
	})

	t.Run("accesslist with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("accesslist", "peer1.i2p,peer2.i2p", c)
		if _, ok := c.Tunnel["accesslist"]; !ok {
			t.Error("expected Tunnel[accesslist] to be set")
		}
	})

	t.Run("signaturetype with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("signaturetype", "7", c)
		if got, ok := c.Tunnel["signaturetype"]; !ok || got != 7 {
			t.Errorf("expected Tunnel[signaturetype]=7, got %v", got)
		}
	})

	t.Run("explicitpeers with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("explicitpeers", "abc.b32.i2p,xyz.b32.i2p", c)
		if _, ok := c.Tunnel["explicitpeers"]; !ok {
			t.Error("expected Tunnel[explicitpeers] to be set")
		}
	})

	t.Run("multicast yes with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("multicast", "yes", c)
		if got, ok := c.Tunnel["multicast"]; !ok || got != true {
			t.Errorf("expected Tunnel[multicast]=true, got %v", got)
		}
	})

	t.Run("maptoloopback with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("maptoloopback", "true", c)
		if got, ok := c.Tunnel["maptoloopback"]; !ok || got != true {
			t.Errorf("expected Tunnel[maptoloopback]=true, got %v", got)
		}
	})

	t.Run("enableuniquelocal false with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("enableuniquelocal", "false", c)
		if got, ok := c.Tunnel["enableuniquelocal"]; !ok || got != false {
			t.Errorf("expected Tunnel[enableuniquelocal]=false, got %v", got)
		}
	})

	t.Run("crypto prefix with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("crypto.tagsToSend", "40", c)
		if got, ok := c.Tunnel["crypto.tagsToSend"]; !ok || got != 40 {
			t.Errorf("expected Tunnel[crypto.tagsToSend]=40, got %v", got)
		}
	})

	t.Run("streamr prefix with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("streamr.rto", "60000", c)
		if got, ok := c.Tunnel["streamr.rto"]; !ok || got != 60000 {
			t.Errorf("expected Tunnel[streamr.rto]=60000, got %v", got)
		}
	})

	t.Run("i2cp prefix with nil I2CP initialises map", func(t *testing.T) {
		c := &TunnelConfig{
			Tunnel:   make(map[string]interface{}),
			Inbound:  make(map[string]interface{}),
			Outbound: make(map[string]interface{}),
			// I2CP intentionally nil
		}
		conv.parseINIKeyValue("i2cp.leaseSetEncType", "4,0", c)
		if _, ok := c.I2CP["leaseSetEncType"]; !ok {
			t.Error("expected I2CP[leaseSetEncType] to be set")
		}
	})

	t.Run("inbound prefix with nil Inbound initialises map", func(t *testing.T) {
		c := &TunnelConfig{
			I2CP:     make(map[string]interface{}),
			Tunnel:   make(map[string]interface{}),
			Outbound: make(map[string]interface{}),
			// Inbound intentionally nil
		}
		conv.parseINIKeyValue("inbound.length", "3", c)
		if got, ok := c.Inbound["length"]; !ok || got != 3 {
			t.Errorf("expected Inbound[length]=3, got %v", got)
		}
	})

	t.Run("outbound prefix with nil Outbound initialises map", func(t *testing.T) {
		c := &TunnelConfig{
			I2CP:    make(map[string]interface{}),
			Tunnel:  make(map[string]interface{}),
			Inbound: make(map[string]interface{}),
			// Outbound intentionally nil
		}
		conv.parseINIKeyValue("outbound.length", "2", c)
		if got, ok := c.Outbound["length"]; !ok || got != 2 {
			t.Errorf("expected Outbound[length]=2, got %v", got)
		}
	})

	t.Run("unknown key with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("customoption", "customvalue", c)
		if got, ok := c.Tunnel["customoption"]; !ok || got != "customvalue" {
			t.Errorf("expected Tunnel[customoption]='customvalue', got %v", got)
		}
	})

	t.Run("keys persistent with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("keys", "mykeyfile.dat", c)
		if !c.PersistentKey {
			t.Error("expected PersistentKey=true for named keyfile")
		}
		if got, ok := c.Tunnel["keyfile"]; !ok || got != "mykeyfile.dat" {
			t.Errorf("expected Tunnel[keyfile]='mykeyfile.dat', got %v", got)
		}
	})

	t.Run("hostoverride with nil Tunnel initialises map", func(t *testing.T) {
		c := nilConfig()
		conv.parseINIKeyValue("hostoverride", "override.example.com", c)
		if got, ok := c.Tunnel["hostoverride"]; !ok || got != "override.example.com" {
			t.Errorf("expected Tunnel[hostoverride]='override.example.com', got %v", got)
		}
	})
}

// TestINIClientAddressBecomesInterface verifies that for client tunnel types,
// the i2pd "address" key (local bind address) is mapped to config.Interface
// rather than config.Target.
func TestINIClientAddressBecomesInterface(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantInterface string
		wantTarget    string
	}{
		{
			name: "http alias client address",
			input: `[MyHTTPProxy]
type = http
address = 127.0.0.1
port = 4444
keys = httpclient-keys.dat
`,
			wantInterface: "127.0.0.1",
			wantTarget:    "",
		},
		{
			name: "httpclient canonical client address",
			input: `[HttpClient]
type = httpclient
address = 0.0.0.0
port = 4444
`,
			wantInterface: "0.0.0.0",
			wantTarget:    "",
		},
		{
			name: "ircclient address",
			input: `[IRCProxy]
type = ircclient
address = 127.0.0.1
port = 6668
destination = irc.example.i2p
`,
			wantInterface: "127.0.0.1",
			wantTarget:    "irc.example.i2p",
		},
		{
			name: "server tunnel address stays as Target",
			input: `[WebServer]
type = httpserver
address = webserver.i2p
port = 80
host = 127.0.0.1
`,
			wantInterface: "127.0.0.1",
			wantTarget:    "webserver.i2p",
		},
	}

	conv := &Converter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := conv.ParseInput([]byte(tt.input), "ini")
			if err != nil {
				t.Fatalf("ParseInput: %v", err)
			}
			if config.Interface != tt.wantInterface {
				t.Errorf("Interface: want %q, got %q", tt.wantInterface, config.Interface)
			}
			if config.Target != tt.wantTarget {
				t.Errorf("Target: want %q, got %q", tt.wantTarget, config.Target)
			}
		})
	}
}

// TestLooksLikeI2PDestination exercises the helper directly.
func TestLooksLikeI2PDestination(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"example.i2p", true},
		{"something.b32.i2p", true},
		{"127.0.0.1", false},
		{"0.0.0.0", false},
		{"localhost", false},
		{strings.Repeat("A", 516), true},
		{strings.Repeat("A", 515), false},
	}
	for _, c := range cases {
		got := looksLikeI2PDestination(c.input)
		if got != c.want {
			t.Errorf("looksLikeI2PDestination(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

// TestNormalizeTypeName verifies that i2pd short-form type aliases are mapped to
// their canonical Java-I2P counterparts and that unknown / already-canonical
// names pass through unchanged (lowercased).
func TestNormalizeTypeName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"http", "httpclient"},
		{"socks", "sockstunnel"},
		{"udptunnel", "client"},
		{"httpclient", "httpclient"}, // already canonical
		{"server", "server"},         // unknown alias — pass through
		{"HTTP", "httpclient"},       // case-insensitive
		{"SOCKS", "sockstunnel"},     // case-insensitive
	}
	for _, tc := range cases {
		got := NormalizeTypeName(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeTypeName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestINITypeNormalizationRoundTrip parses an i2pd-style conf that uses the
// short alias "http" and confirms the parsed TunnelConfig holds the canonical
// name "httpclient", and that converting to .properties format preserves it.
func TestINITypeNormalizationRoundTrip(t *testing.T) {
	conv := &Converter{}
	ini := "[MyProxy]\ntype = http\nport = 4444\nkeys = transient\n"
	config, err := conv.ParseInput([]byte(ini), "ini")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if config.Type != "httpclient" {
		t.Errorf("Type after INI parse: got %q, want \"httpclient\"", config.Type)
	}
	props, err := conv.generateOutput(config, "properties")
	if err != nil {
		t.Fatalf("generate properties: %v", err)
	}
	if !strings.Contains(string(props), "type=httpclient") {
		t.Errorf("properties output missing type=httpclient:\n%s", props)
	}
}
