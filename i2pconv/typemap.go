package i2pconv

import "strings"

// i2pdTypeAliases maps i2pd-style tunnel type names to their canonical
// Java I2P / go-i2p equivalents. i2pd uses shorter names (e.g., "http")
// while Java I2P uses descriptive names (e.g., "httpclient"). The canonical
// name is what all internal TunnelConfig representations and output generators
// use, ensuring lossless round-trips between formats.
var i2pdTypeAliases = map[string]string{
	"http":      "httpclient",
	"socks":     "sockstunnel",
	"udptunnel": "client",
}

// NormalizeTypeName converts an i2pd alias type name to its canonical equivalent.
// If the name is already canonical (or unknown), it is returned unchanged.
func NormalizeTypeName(t string) string {
	lower := strings.ToLower(t)
	if canonical, ok := i2pdTypeAliases[lower]; ok {
		return canonical
	}
	return lower
}
