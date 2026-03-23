package i2pconv

import (
	"testing"
)

func TestValidationContext_BasicValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "valid basic config",
			config: &TunnelConfig{
				Name: "test",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "missing name",
			config: &TunnelConfig{
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			expectedErr: true,
			errorText:   "tunnel name is required",
		},
		{
			name: "missing type",
			config: &TunnelConfig{
				Name: "test",
				Port: 8080,
			},
			strict:      false,
			expectedErr: true,
			errorText:   "tunnel type is required",
		},
		{
			name: "invalid name characters",
			config: &TunnelConfig{
				Name: "test tunnel",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			expectedErr: true,
			errorText:   "contains invalid characters",
		},
		{
			name: "name with equals sign",
			config: &TunnelConfig{
				Name: "test=tunnel",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			expectedErr: true,
			errorText:   "contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, "")
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_HTTPClientValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "valid httpclient",
			config: &TunnelConfig{
				Name: "webclient",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "httpclient without port",
			config: &TunnelConfig{
				Name: "webclient",
				Type: "httpclient",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "Local port is required",
		},
		{
			name: "httpclient with invalid port",
			config: &TunnelConfig{
				Name: "webclient",
				Type: "httpclient",
				Port: 70000,
			},
			strict:      false,
			expectedErr: true,
			errorText:   "port 70000 is out of valid range",
		},
		{
			name: "httpclient with privileged port (non-strict)",
			config: &TunnelConfig{
				Name: "webclient",
				Type: "httpclient",
				Port: 80,
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "httpclient with privileged port (strict)",
			config: &TunnelConfig{
				Name: "webclient",
				Type: "httpclient",
				Port: 80,
			},
			strict:      true,
			expectedErr: true,
			errorText:   "privileged range",
		},
		{
			name: "httpclient with valid target",
			config: &TunnelConfig{
				Name:   "webclient",
				Type:   "httpclient",
				Port:   8080,
				Target: "localhost:3000",
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "httpclient with invalid target",
			config: &TunnelConfig{
				Name:   "webclient",
				Type:   "httpclient",
				Port:   8080,
				Target: "localhost:99999",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "target port 99999 is out of valid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, "")
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_HTTPServerValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "valid httpserver",
			config: &TunnelConfig{
				Name:   "webserver",
				Type:   "httpserver",
				Target: "127.0.0.1:8080",
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "httpserver without target",
			config: &TunnelConfig{
				Name: "webserver",
				Type: "httpserver",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "Target is required for HTTP server",
		},
		{
			name: "httpserver with invalid target format",
			config: &TunnelConfig{
				Name:   "webserver",
				Type:   "httpserver",
				Target: "host:port:extra",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "invalid format",
		},
		{
			name: "httpserver with empty host",
			config: &TunnelConfig{
				Name:   "webserver",
				Type:   "httpserver",
				Target: ":8080",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "empty host part",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, "")
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_InterfaceValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "valid IPv4 interface",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "127.0.0.1",
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "valid IPv6 interface",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "::1",
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "localhost interface",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "localhost",
			},
			strict:      false,
			expectedErr: false,
		},
		{
			name: "interface with spaces",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "127.0.0.1 bad",
			},
			strict:      false,
			expectedErr: true,
			errorText:   "whitespace characters",
		},
		{
			name: "invalid interface (strict)",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "eth0",
			},
			strict:      true,
			expectedErr: true,
			errorText:   "should be a valid IP address",
		},
		{
			name: "invalid interface (non-strict)",
			config: &TunnelConfig{
				Name:      "test",
				Type:      "httpclient",
				Port:      8080,
				Interface: "eth0",
			},
			strict:      false,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, "")
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_FormatSpecificValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		format      string
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "properties format with dot in name (strict)",
			config: &TunnelConfig{
				Name: "test.tunnel",
				Type: "httpclient",
				Port: 8080,
			},
			format:      "properties",
			strict:      true,
			expectedErr: true,
			errorText:   "may cause issues in properties format",
		},
		{
			name: "properties format with dot in name (non-strict)",
			config: &TunnelConfig{
				Name: "test.tunnel",
				Type: "httpclient",
				Port: 8080,
			},
			format:      "properties",
			strict:      false,
			expectedErr: false,
		},
		{
			name: "ini format with brackets in name (strict)",
			config: &TunnelConfig{
				Name: "test[tunnel]",
				Type: "httpclient",
				Port: 8080,
			},
			format:      "ini",
			strict:      true,
			expectedErr: true,
			errorText:   "contains invalid characters", // Basic validation catches this first
		},
		{
			name: "yaml format with leading spaces (strict)",
			config: &TunnelConfig{
				Name: " testtunnel",
				Type: "httpclient",
				Port: 8080,
			},
			format:      "yaml",
			strict:      true,
			expectedErr: true,
			errorText:   "contains invalid characters", // Basic validation catches this first
		},
		{
			name: "yaml format normal name",
			config: &TunnelConfig{
				Name: "testtunnel",
				Type: "httpclient",
				Port: 8080,
			},
			format:      "yaml",
			strict:      true,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, tt.format)
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_UnknownTunnelType(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		expectedErr bool
		errorText   string
	}{
		{
			name: "unknown tunnel type (strict)",
			config: &TunnelConfig{
				Name: "test",
				Type: "unknowntype",
				Port: 8080,
			},
			strict:      true,
			expectedErr: true,
			errorText:   "unknown tunnel type: unknowntype",
		},
		{
			name: "unknown tunnel type (non-strict)",
			config: &TunnelConfig{
				Name: "test",
				Type: "unknowntype",
				Port: 8080,
			},
			strict:      false,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewValidationContext(tt.strict, "")
			err := ctx.Validate(tt.config)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

func TestValidationContext_TunnelTypeSpecs(t *testing.T) {
	ctx := NewValidationContext(false, "")

	// Test that all expected tunnel types are defined
	expectedTypes := []TunnelType{
		TunnelTypeHTTPClient,
		TunnelTypeHTTPServer,
		TunnelTypeSOCKS,
		TunnelTypeSOCKSServer,
		TunnelTypeClient,
		TunnelTypeServer,
		TunnelTypeIRCClient,
		TunnelTypeIRCServer,
		TunnelTypeStreamClient,
		TunnelTypeStreamServer,
		TunnelTypeHTTPBidir,
		TunnelTypeSOCKSIRC,
	}

	for _, expectedType := range expectedTypes {
		t.Run(string(expectedType), func(t *testing.T) {
			spec, exists := ctx.TunnelSpecs[expectedType]
			if !exists {
				t.Errorf("tunnel type %s not found in specs", expectedType)
				return
			}

			if spec.Name != expectedType {
				t.Errorf("expected name %s, got %s", expectedType, spec.Name)
			}

			if spec.Description == "" {
				t.Errorf("description is empty for tunnel type %s", expectedType)
			}

			if len(spec.Rules) == 0 {
				t.Errorf("no validation rules defined for tunnel type %s", expectedType)
			}
		})
	}

	// Test GetSupportedTunnelTypes
	supportedTypes := ctx.GetSupportedTunnelTypes()
	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}

	// Test GetTunnelTypeDescription
	desc := ctx.GetTunnelTypeDescription(TunnelTypeHTTPClient)
	if desc == "Unknown tunnel type" {
		t.Errorf("expected valid description for httpclient, got unknown")
	}

	unknownDesc := ctx.GetTunnelTypeDescription("nonexistent")
	if unknownDesc != "Unknown tunnel type" {
		t.Errorf("expected 'Unknown tunnel type' for nonexistent type, got '%s'", unknownDesc)
	}
}

func TestConverter_ValidationIntegration(t *testing.T) {
	tests := []struct {
		name        string
		config      *TunnelConfig
		strict      bool
		format      string
		expectedErr bool
		errorText   string
	}{
		{
			name: "converter basic validation",
			config: &TunnelConfig{
				Name: "test",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			format:      "",
			expectedErr: false,
		},
		{
			name: "converter validation with format",
			config: &TunnelConfig{
				Name: "test",
				Type: "httpclient",
				Port: 8080,
			},
			strict:      false,
			format:      "yaml",
			expectedErr: false,
		},
		{
			name: "converter validation failure",
			config: &TunnelConfig{
				Name: "test",
				Type: "httpclient",
				// Missing required port
			},
			strict:      false,
			format:      "",
			expectedErr: true,
			errorText:   "Local port is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &Converter{strict: tt.strict}

			var err error
			if tt.format == "" {
				err = converter.validate(tt.config)
			} else {
				err = converter.validateWithFormat(tt.config, tt.format)
			}

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err != nil && tt.errorText != "" {
				if !contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) <= len(s) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}
