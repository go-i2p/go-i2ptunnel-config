package i2pconv

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// TunnelType represents the different types of I2P tunnels
type TunnelType string

const (
	// Client tunnel types
	TunnelTypeHTTPClient   TunnelType = "httpclient"
	TunnelTypeSOCKS        TunnelType = "sockstunnel"
	TunnelTypeSOCKSServer  TunnelType = "socksserver"
	TunnelTypeIRCClient    TunnelType = "ircclient"
	TunnelTypeClient       TunnelType = "client"
	TunnelTypeStreamClient TunnelType = "streamrclient"

	// Server tunnel types
	TunnelTypeHTTPServer   TunnelType = "httpserver"
	TunnelTypeServer       TunnelType = "server"
	TunnelTypeIRCServer    TunnelType = "ircserver"
	TunnelTypeStreamServer TunnelType = "streamrserver"

	// Special types
	TunnelTypeHTTPBidir TunnelType = "httpbidirserver"
	TunnelTypeSOCKSIRC  TunnelType = "socksirc"
)

// ValidationRule defines a validation constraint
type ValidationRule struct {
	Field       string
	Required    bool
	Validator   func(config *TunnelConfig) error
	Description string
}

// TunnelTypeSpec defines validation rules for a specific tunnel type
type TunnelTypeSpec struct {
	Name        TunnelType
	Description string
	Rules       []ValidationRule
}

// ValidationContext holds validation configuration
type ValidationContext struct {
	Strict      bool
	Format      string
	TunnelSpecs map[TunnelType]TunnelTypeSpec
}

// NewValidationContext creates a new validation context with predefined tunnel type specifications
func NewValidationContext(strict bool, format string) *ValidationContext {
	ctx := &ValidationContext{
		Strict:      strict,
		Format:      format,
		TunnelSpecs: make(map[TunnelType]TunnelTypeSpec),
	}
	ctx.initializeTunnelSpecs()
	return ctx
}

// initializeTunnelSpecs defines validation rules for each tunnel type
func (v *ValidationContext) initializeTunnelSpecs() {
	// HTTP Client tunnel - requires local port, may have target
	v.TunnelSpecs[TunnelTypeHTTPClient] = TunnelTypeSpec{
		Name:        TunnelTypeHTTPClient,
		Description: "HTTP proxy client tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for HTTP client"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
			{Field: "target", Required: false, Validator: v.validateTarget, Description: "Target must be valid if specified"},
		},
	}

	// HTTP Server tunnel - requires target, may have port
	v.TunnelSpecs[TunnelTypeHTTPServer] = TunnelTypeSpec{
		Name:        TunnelTypeHTTPServer,
		Description: "HTTP server tunnel",
		Rules: []ValidationRule{
			{Field: "target", Required: true, Validator: v.validateTarget, Description: "Target is required for HTTP server"},
			{Field: "port", Required: false, Validator: v.validatePort, Description: "Port must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// SOCKS tunnel - requires port
	v.TunnelSpecs[TunnelTypeSOCKS] = TunnelTypeSpec{
		Name:        TunnelTypeSOCKS,
		Description: "SOCKS proxy tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for SOCKS tunnel"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// SOCKS Server - similar to SOCKS
	socksServerSpec := v.TunnelSpecs[TunnelTypeSOCKS]
	socksServerSpec.Name = TunnelTypeSOCKSServer
	socksServerSpec.Description = "SOCKS server tunnel"
	v.TunnelSpecs[TunnelTypeSOCKSServer] = socksServerSpec

	// Generic client tunnel - requires port
	v.TunnelSpecs[TunnelTypeClient] = TunnelTypeSpec{
		Name:        TunnelTypeClient,
		Description: "Generic client tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for client tunnel"},
			{Field: "target", Required: false, Validator: v.validateTarget, Description: "Target must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// Generic server tunnel - requires target
	v.TunnelSpecs[TunnelTypeServer] = TunnelTypeSpec{
		Name:        TunnelTypeServer,
		Description: "Generic server tunnel",
		Rules: []ValidationRule{
			{Field: "target", Required: true, Validator: v.validateTarget, Description: "Target is required for server tunnel"},
			{Field: "port", Required: false, Validator: v.validatePort, Description: "Port must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// IRC Client - requires port
	v.TunnelSpecs[TunnelTypeIRCClient] = TunnelTypeSpec{
		Name:        TunnelTypeIRCClient,
		Description: "IRC client tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for IRC client"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// IRC Server - requires target
	v.TunnelSpecs[TunnelTypeIRCServer] = TunnelTypeSpec{
		Name:        TunnelTypeIRCServer,
		Description: "IRC server tunnel",
		Rules: []ValidationRule{
			{Field: "target", Required: true, Validator: v.validateTarget, Description: "Target is required for IRC server"},
			{Field: "port", Required: false, Validator: v.validatePort, Description: "Port must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// Streaming client - requires port
	v.TunnelSpecs[TunnelTypeStreamClient] = TunnelTypeSpec{
		Name:        TunnelTypeStreamClient,
		Description: "Streaming client tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for streaming client"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// Streaming server - requires target
	v.TunnelSpecs[TunnelTypeStreamServer] = TunnelTypeSpec{
		Name:        TunnelTypeStreamServer,
		Description: "Streaming server tunnel",
		Rules: []ValidationRule{
			{Field: "target", Required: true, Validator: v.validateTarget, Description: "Target is required for streaming server"},
			{Field: "port", Required: false, Validator: v.validatePort, Description: "Port must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// HTTP Bidirectional server - special case
	v.TunnelSpecs[TunnelTypeHTTPBidir] = TunnelTypeSpec{
		Name:        TunnelTypeHTTPBidir,
		Description: "HTTP bidirectional server tunnel",
		Rules: []ValidationRule{
			{Field: "target", Required: true, Validator: v.validateTarget, Description: "Target is required for bidirectional server"},
			{Field: "port", Required: false, Validator: v.validatePort, Description: "Port must be valid if specified"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}

	// SOCKS IRC - hybrid type
	v.TunnelSpecs[TunnelTypeSOCKSIRC] = TunnelTypeSpec{
		Name:        TunnelTypeSOCKSIRC,
		Description: "SOCKS IRC proxy tunnel",
		Rules: []ValidationRule{
			{Field: "port", Required: true, Validator: v.validatePort, Description: "Local port is required for SOCKS IRC"},
			{Field: "interface", Required: false, Validator: v.validateInterface, Description: "Interface must be valid if specified"},
		},
	}
}

// Validate validates a tunnel configuration according to its type and format
func (v *ValidationContext) Validate(config *TunnelConfig) error {
	// Basic validation - name and type are always required
	if err := v.validateBasicFields(config); err != nil {
		return err
	}

	// Type-specific validation
	tunnelType := TunnelType(strings.ToLower(config.Type))
	spec, exists := v.TunnelSpecs[tunnelType]
	if !exists {
		if v.Strict {
			return fmt.Errorf("unknown tunnel type: %s", config.Type)
		}
		// In non-strict mode, only do basic validation for unknown types
		return nil
	}

	// Apply type-specific rules
	for _, rule := range spec.Rules {
		if err := v.applyRule(config, rule); err != nil {
			return err
		}
	}

	// Format-specific validation
	if err := v.validateFormatSpecific(config); err != nil {
		return err
	}

	return nil
}

// validateBasicFields performs basic validation that applies to all tunnel types
func (v *ValidationContext) validateBasicFields(config *TunnelConfig) error {
	if config.Name == "" {
		return fmt.Errorf("tunnel name is required")
	}

	if config.Type == "" {
		return fmt.Errorf("tunnel type is required")
	}

	// Validate name doesn't contain problematic characters
	if strings.ContainsAny(config.Name, " \t\n\r=[]") {
		return fmt.Errorf("tunnel name '%s' contains invalid characters (spaces, tabs, newlines, equals, or brackets)", config.Name)
	}

	return nil
}

// applyRule applies a specific validation rule to the configuration
func (v *ValidationContext) applyRule(config *TunnelConfig, rule ValidationRule) error {
	// Check if field is required but missing
	if rule.Required {
		switch rule.Field {
		case "port":
			if config.Port <= 0 {
				return fmt.Errorf("%s: %s", rule.Description, "port must be specified and greater than 0")
			}
		case "target":
			if config.Target == "" {
				return fmt.Errorf("%s: %s", rule.Description, "target must be specified")
			}
		case "interface":
			if config.Interface == "" {
				return fmt.Errorf("%s: %s", rule.Description, "interface must be specified")
			}
		}
	}

	// Apply field-specific validation if field has a value
	if rule.Validator != nil {
		if err := rule.Validator(config); err != nil {
			return fmt.Errorf("%s: %w", rule.Description, err)
		}
	}

	return nil
}

// validatePort validates that the port is in a valid range
func (v *ValidationContext) validatePort(config *TunnelConfig) error {
	if config.Port <= 0 {
		return nil // No port specified, which is valid for some tunnel types
	}

	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port %d is out of valid range (1-65535)", config.Port)
	}

	// In strict mode, warn about privileged ports
	if v.Strict && config.Port < 1024 {
		return fmt.Errorf("port %d is in privileged range (1-1023), may require root privileges", config.Port)
	}

	return nil
}

// validateInterface validates that the interface specification is valid
func (v *ValidationContext) validateInterface(config *TunnelConfig) error {
	if config.Interface == "" {
		return nil // No interface specified
	}

	// Check if it's a valid IP address
	if ip := net.ParseIP(config.Interface); ip != nil {
		return nil // Valid IP address
	}

	// Check common interface names
	validInterfaces := []string{"localhost", "127.0.0.1", "0.0.0.0", "::", "::1"}
	for _, valid := range validInterfaces {
		if config.Interface == valid {
			return nil
		}
	}

	// In strict mode, be more restrictive
	if v.Strict {
		return fmt.Errorf("interface '%s' should be a valid IP address or localhost", config.Interface)
	}

	// In non-strict mode, allow any reasonable interface name
	if strings.ContainsAny(config.Interface, " \t\n\r") {
		return fmt.Errorf("interface '%s' contains whitespace characters", config.Interface)
	}

	return nil
}

// validateTarget validates that the target specification is valid
func (v *ValidationContext) validateTarget(config *TunnelConfig) error {
	if config.Target == "" {
		return nil // No target specified
	}

	// Target can be in format "host:port" or just "host"
	parts := strings.Split(config.Target, ":")

	if len(parts) > 2 {
		return fmt.Errorf("target '%s' has invalid format, should be 'host' or 'host:port'", config.Target)
	}

	host := parts[0]
	if host == "" {
		return fmt.Errorf("target '%s' has empty host part", config.Target)
	}

	// Validate host part - can be IP, hostname, or localhost
	if ip := net.ParseIP(host); ip == nil {
		// Not an IP, check if it's a reasonable hostname
		if strings.ContainsAny(host, " \t\n\r") {
			return fmt.Errorf("target host '%s' contains whitespace characters", host)
		}
	}

	// If port is specified, validate it
	if len(parts) == 2 {
		portStr := parts[1]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("target port '%s' is not a valid number", portStr)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("target port %d is out of valid range (1-65535)", port)
		}
	}

	return nil
}

// validateFormatSpecific performs validation specific to the configuration format
func (v *ValidationContext) validateFormatSpecific(config *TunnelConfig) error {
	switch v.Format {
	case "properties":
		return v.validatePropertiesFormat(config)
	case "ini":
		return v.validateINIFormat(config)
	case "yaml":
		return v.validateYAMLFormat(config)
	default:
		return nil // Unknown format, skip format-specific validation
	}
}

// validatePropertiesFormat validates properties-format-specific constraints
func (v *ValidationContext) validatePropertiesFormat(config *TunnelConfig) error {
	// Properties format has specific constraints for certain fields
	if v.Strict {
		// In strict mode, check for properties format compatibility
		if strings.ContainsAny(config.Name, ".=") {
			return fmt.Errorf("tunnel name '%s' contains characters that may cause issues in properties format", config.Name)
		}
	}
	return nil
}

// validateINIFormat validates INI-format-specific constraints
func (v *ValidationContext) validateINIFormat(config *TunnelConfig) error {
	// INI format has section name constraints
	if v.Strict {
		if strings.ContainsAny(config.Name, "[]") {
			return fmt.Errorf("tunnel name '%s' contains characters that may cause issues in INI format", config.Name)
		}
	}
	return nil
}

// validateYAMLFormat validates YAML-format-specific constraints
func (v *ValidationContext) validateYAMLFormat(config *TunnelConfig) error {
	// YAML format is generally more flexible, fewer constraints
	if v.Strict {
		if strings.HasPrefix(config.Name, " ") || strings.HasSuffix(config.Name, " ") {
			return fmt.Errorf("tunnel name '%s' has leading or trailing spaces that may cause issues in YAML format", config.Name)
		}
	}
	return nil
}

// GetSupportedTunnelTypes returns a list of all supported tunnel types
func (v *ValidationContext) GetSupportedTunnelTypes() []TunnelType {
	types := make([]TunnelType, 0, len(v.TunnelSpecs))
	for tunnelType := range v.TunnelSpecs {
		types = append(types, tunnelType)
	}
	return types
}

// GetTunnelTypeDescription returns a description for a specific tunnel type
func (v *ValidationContext) GetTunnelTypeDescription(tunnelType TunnelType) string {
	if spec, exists := v.TunnelSpecs[tunnelType]; exists {
		return spec.Description
	}
	return "Unknown tunnel type"
}
