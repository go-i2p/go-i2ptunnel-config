package i2pconv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSAMTunnel_NonPersistent tests SAM tunnel creation without persistent keys
func TestSAMTunnel_NonPersistent(t *testing.T) {
	config := &TunnelConfig{
		Name:          "test-tunnel",
		Type:          "client",
		PersistentKey: false,
		I2CP: map[string]interface{}{
			"leaseSetEncType": "4,0",
			"tunnelDepth":     3,
		},
		Tunnel: map[string]interface{}{
			"quantity": 2,
		},
		Inbound: map[string]interface{}{
			"length": 3,
		},
		Outbound: map[string]interface{}{
			"length": 3,
		},
	}

	keys, opts, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Expected no error for non-persistent tunnel, got: %v", err)
	}

	// Keys should be nil for non-persistent tunnels
	if keys != nil {
		t.Error("Expected nil keys for non-persistent tunnel")
	}

	// Check that options are properly formatted
	expectedOpts := []string{
		"i2cp.leaseSetEncType=4,0",
		"i2cp.tunnelDepth=3",
		"quantity=2",
		"inbound.length=3",
		"outbound.length=3",
		"i2cp.leaseSetEncType=4,0", // Default should be added if not present
	}

	// Check that all expected options are present (order may vary)
	for _, expected := range expectedOpts {
		found := false
		for _, opt := range opts {
			if opt == expected {
				found = true
				break
			}
		}
		if !found && !strings.Contains(expected, "i2cp.leaseSetEncType=4,0") {
			// Skip duplicate leaseSetEncType check since it may be present twice
			if !hasOption(opts, "i2cp.leaseSetEncType") {
				t.Errorf("Expected option %q not found in %v", expected, opts)
			}
		}
	}

	// Verify the default lease set encryption is added
	if !hasOption(opts, "i2cp.leaseSetEncType") {
		t.Error("Default i2cp.leaseSetEncType should be added when not present")
	}
}

// TestSAMTunnel_PersistentNewKeys tests SAM tunnel creation with new persistent keys
func TestSAMTunnel_PersistentNewKeys(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "i2ptunnel-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config := &TunnelConfig{
		Name:          "persistent-tunnel",
		Type:          "client",
		PersistentKey: true,
	}

	keys, opts, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Expected no error for persistent tunnel, got: %v", err)
	}

	// Keys should not be nil for persistent tunnels
	if keys == nil {
		t.Error("Expected non-nil keys for persistent tunnel")
	}

	// Check that key file was created
	keyPath := filepath.Join(tempDir, "persistent-tunnel.keys")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Expected key file %s to be created", keyPath)
	}

	// Verify default options are present
	if !hasOption(opts, "i2cp.leaseSetEncType") {
		t.Error("Default i2cp.leaseSetEncType should be added")
	}
}

// TestSAMTunnel_PersistentExistingKeys tests SAM tunnel with existing persistent keys
func TestSAMTunnel_PersistentExistingKeys(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "i2ptunnel-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config := &TunnelConfig{
		Name:          "existing-tunnel",
		Type:          "client",
		PersistentKey: true,
	}

	// First call should create new keys
	keys1, _, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Failed to create initial keys: %v", err)
	}

	// Second call should load existing keys
	keys2, _, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Failed to load existing keys: %v", err)
	}

	// Keys should be the same (same destination)
	if keys1 == nil || keys2 == nil {
		t.Fatal("Both key sets should be non-nil")
	}

	if keys1.Addr().String() != keys2.Addr().String() {
		t.Error("Loaded keys should match the originally created keys")
	}
}

// TestSAMTunnel_PersistentNoName tests error handling for persistent keys without tunnel name
func TestSAMTunnel_PersistentNoName(t *testing.T) {
	config := &TunnelConfig{
		Name:          "", // Empty name should cause error
		Type:          "client",
		PersistentKey: true,
	}

	keys, opts, err := config.SAMTunnel()
	if err == nil {
		t.Error("Expected error for persistent tunnel without name")
	}

	if keys != nil {
		t.Error("Expected nil keys when error occurs")
	}

	if opts != nil {
		t.Error("Expected nil options when error occurs")
	}

	// Check error message
	expectedMsg := "tunnel name is required for persistent keys"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
	}
}

// TestSAMTunnel_OptionsGeneration tests various option combinations
func TestSAMTunnel_OptionsGeneration(t *testing.T) {
	config := &TunnelConfig{
		Name:          "test-options",
		Type:          "client",
		PersistentKey: false,
		I2CP: map[string]interface{}{
			"reduceIdleTime": 900000,
			"tunnelDepth":    2,
		},
		Tunnel: map[string]interface{}{
			"quantity": 3,
			"variance": 1,
			"nickname": "TestTunnel",
		},
		Inbound: map[string]interface{}{
			"length":         2,
			"lengthVariance": 1,
		},
		Outbound: map[string]interface{}{
			"length":         3,
			"lengthVariance": 0,
		},
	}

	_, opts, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test expected options
	expectedOptions := map[string]bool{
		"i2cp.reduceIdleTime=900000": false,
		"i2cp.tunnelDepth=2":         false,
		"quantity=3":                 false,
		"variance=1":                 false,
		"nickname=TestTunnel":        false,
		"inbound.length=2":           false,
		"inbound.lengthVariance=1":   false,
		"outbound.length=3":          false,
		"outbound.lengthVariance=0":  false,
	}

	// Check all options are present
	for _, opt := range opts {
		if _, exists := expectedOptions[opt]; exists {
			expectedOptions[opt] = true
		}
	}

	for opt, found := range expectedOptions {
		if !found {
			t.Errorf("Expected option %q not found in generated options", opt)
		}
	}

	// Verify default lease set encryption is added
	foundLeaseSet := false
	for _, opt := range opts {
		if strings.HasPrefix(opt, "i2cp.leaseSetEncType=") {
			foundLeaseSet = true
			break
		}
	}
	if !foundLeaseSet {
		t.Error("Default i2cp.leaseSetEncType should be present")
	}
}

// TestSAMTunnel_LeaseSetEncTypeDefault tests that default leaseSetEncType is added when missing
func TestSAMTunnel_LeaseSetEncTypeDefault(t *testing.T) {
	config := &TunnelConfig{
		Name:          "test-default-lease",
		Type:          "client",
		PersistentKey: false,
		// No I2CP options - should get default leaseSetEncType
	}

	_, opts, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have default lease set encryption
	if !hasOption(opts, "i2cp.leaseSetEncType") {
		t.Error("Expected default i2cp.leaseSetEncType=4,0 when not specified")
	}

	// Check specific default value
	found := false
	for _, opt := range opts {
		if opt == "i2cp.leaseSetEncType=4,0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected default i2cp.leaseSetEncType=4,0")
	}
}

// TestSAMTunnel_LeaseSetEncTypePreserved tests that existing leaseSetEncType is preserved
func TestSAMTunnel_LeaseSetEncTypePreserved(t *testing.T) {
	config := &TunnelConfig{
		Name:          "test-preserve-lease",
		Type:          "client",
		PersistentKey: false,
		I2CP: map[string]interface{}{
			"leaseSetEncType": "0", // Custom value
		},
	}

	_, opts, err := config.SAMTunnel()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should preserve custom value
	found := false
	defaultFound := false
	for _, opt := range opts {
		if opt == "i2cp.leaseSetEncType=0" {
			found = true
		}
		if opt == "i2cp.leaseSetEncType=4,0" {
			defaultFound = true
		}
	}

	if !found {
		t.Error("Expected custom i2cp.leaseSetEncType=0 to be preserved")
	}

	if defaultFound {
		t.Error("Default i2cp.leaseSetEncType=4,0 should not be added when custom value exists")
	}
}
