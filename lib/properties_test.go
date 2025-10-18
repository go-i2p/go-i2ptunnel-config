package i2pconv

import "testing"

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

	// Expected YAML structure matching actual format:
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
`
	// Compare YAML output
	if string(yaml) != expected {
		t.Fatalf("Unexpected YAML output:\n%s\nExpected:\n%s", yaml, expected)
	} else {
		t.Logf("YAML output:\n%s", yaml)
	}
}
