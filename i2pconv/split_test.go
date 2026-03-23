package i2pconv

import (
	"testing"
)

// TestSplitTunnels_INI verifies that a 3-section INI file produces 3 TunnelConfigs.
func TestSplitTunnels_INI(t *testing.T) {
	input := `[TunnelA]
type = httpclient
host = 127.0.0.1
port = 4444

[TunnelB]
type = server
host = 0.0.0.0
port = 8080
address = mysite.i2p

[TunnelC]
type = client
host = 127.0.0.1
port = 6668
`
	conv := &Converter{}
	configs, err := conv.SplitTunnels([]byte(input), "ini")
	if err != nil {
		t.Fatalf("SplitTunnels returned error: %v", err)
	}
	if len(configs) != 3 {
		t.Fatalf("expected 3 tunnels, got %d", len(configs))
	}
	names := map[string]bool{}
	for _, cfg := range configs {
		names[cfg.Name] = true
	}
	for _, want := range []string{"TunnelA", "TunnelB", "TunnelC"} {
		if !names[want] {
			t.Errorf("expected tunnel %q in results, got names: %v", want, names)
		}
	}
}

// TestSplitTunnels_YAML verifies that a multi-tunnel YAML file is split correctly.
func TestSplitTunnels_YAML(t *testing.T) {
	input := `tunnels:
  alpha:
    type: httpclient
    interface: 127.0.0.1
    port: 4444
  beta:
    type: server
    target: mysite.i2p
  gamma:
    type: client
    interface: 127.0.0.1
    port: 6668
`
	conv := &Converter{}
	configs, err := conv.SplitTunnels([]byte(input), "yaml")
	if err != nil {
		t.Fatalf("SplitTunnels returned error: %v", err)
	}
	if len(configs) != 3 {
		t.Fatalf("expected 3 tunnels, got %d", len(configs))
	}
	names := map[string]bool{}
	for _, cfg := range configs {
		names[cfg.Name] = true
	}
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !names[want] {
			t.Errorf("expected tunnel %q in results, got names: %v", want, names)
		}
	}
}

// TestSplitTunnels_Properties verifies that numbered tunnel.N.* properties are split.
func TestSplitTunnels_Properties(t *testing.T) {
	input := `tunnel.0.name=MyHTTPClient
tunnel.0.type=httpclient
tunnel.0.interface=127.0.0.1
tunnel.0.listenPort=4444
tunnel.1.name=MyServer
tunnel.1.type=server
tunnel.1.targetDestination=mysite.i2p
`
	conv := &Converter{}
	configs, err := conv.SplitTunnels([]byte(input), "properties")
	if err != nil {
		t.Fatalf("SplitTunnels returned error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 tunnels, got %d", len(configs))
	}
}

// TestSplitTunnels_UnsupportedFormat verifies that an error is returned for unknown formats.
func TestSplitTunnels_UnsupportedFormat(t *testing.T) {
	conv := &Converter{}
	_, err := conv.SplitTunnels([]byte("data"), "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}

// TestSplitTunnels_SingleSection verifies that a single-section INI is returned as-is.
func TestSplitTunnels_SingleSection(t *testing.T) {
	input := `[OnlyTunnel]
type = httpclient
host = 127.0.0.1
port = 4444
`
	conv := &Converter{}
	configs, err := conv.SplitTunnels([]byte(input), "ini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(configs))
	}
	if configs[0].Name != "OnlyTunnel" {
		t.Errorf("expected name 'OnlyTunnel', got %q", configs[0].Name)
	}
}
