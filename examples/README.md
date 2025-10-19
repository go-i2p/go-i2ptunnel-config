# I2P Tunnel Configuration Examples

This directory contains example tunnel configuration files for all three supported formats:

- **Java I2P format** (`.properties` files)
- **i2pd format** (`.conf` files)
- **go-i2p format** (`.yaml` files)

## Available Examples

### HTTP Client Tunnel (HTTP Proxy)

Creates a local HTTP proxy to access I2P websites (eepsites).

- `httpclient.properties` - Java I2P format
- `httpclient.conf` - i2pd format
- `httpclient.yaml` - go-i2p format

**Use case**: Browse eepsites through your web browser configured to use the proxy.

**Default port**: 4444

### HTTP Server Tunnel (Eepsite)

Publishes a local web server as an I2P hidden service.

- `httpserver.properties` - Java I2P format
- `httpserver.conf` - i2pd format
- `httpserver.yaml` - go-i2p format

**Use case**: Host your own website accessible only through I2P.

**Target**: Local web server (default: 127.0.0.1:8080)

### SOCKS Tunnel (SOCKS Proxy)

Creates a SOCKS5 proxy for general I2P network access.

- `socks.properties` - Java I2P format
- `socks.conf` - i2pd format
- `socks.yaml` - go-i2p format

**Use case**: Route any SOCKS-compatible application through I2P (IRC, SSH, etc.).

**Default port**: 9050

### Generic Client Tunnel

Creates a TCP client tunnel for any protocol.

- `client.properties` - Java I2P format
- `client.conf` - i2pd format
- `client.yaml` - go-i2p format

**Use case**: Connect to any I2P service using a custom protocol.

**Default port**: 7000

### Generic Server Tunnel

Publishes any local TCP service as an I2P hidden service.

- `server.properties` - Java I2P format
- `server.conf` - i2pd format
- `server.yaml` - go-i2p format

**Use case**: Make any TCP service accessible through I2P (SSH server, game server, etc.).

**Target**: Local service (default: 127.0.0.1:9000)

## Using the Examples

### Converting Between Formats

Convert any example to a different format using the tool:

```bash
# Convert Java I2P properties to YAML
go-i2ptunnel-config httpclient.properties

# Convert i2pd conf to properties
go-i2ptunnel-config --out-format properties httpserver.conf

# Convert YAML to i2pd conf
go-i2ptunnel-config --out-format ini socks.yaml
```

### Validating Configuration

Before using a configuration file, validate it:

```bash
go-i2ptunnel-config --validate httpclient.properties
go-i2ptunnel-config --validate --strict server.yaml
```

### Customizing Examples

All examples contain sensible defaults. Modify these fields for your use case:

#### Required Fields

- **name**: Unique identifier for your tunnel
- **type**: Tunnel type (don't change unless you know what you're doing)

#### Common Fields to Customize

**For Client Tunnels (HTTP Client, SOCKS, Generic Client)**:

- **interface**: Network interface to bind (default: 127.0.0.1)
- **port**: Local port to listen on
- **target/destination**: I2P destination to connect to (for generic client)

**For Server Tunnels (HTTP Server, Generic Server)**:

- **target**: Local service address (format: `host:port`)
- **spoofedHost**: Custom hostname for your eepsite (optional)

#### Advanced Options

**I2CP Options** (i2cp.*):

- `leaseSetEncType`: Encryption type for the lease set (recommended: "4,0")
- `closeIdleTime`: Time in milliseconds before closing idle connections
- `newDestOnResume`: Whether to create a new destination on restart

**Tunnel Options** (inbound/outbound):

- `length`: Number of hops in the tunnel (higher = more anonymous, slower)
- `quantity`: Number of parallel tunnels (higher = more reliable, more resources)

**Other Options**:

- `persistentKey`: Keep the same I2P address across restarts (true/false)
- `gzip`: Enable compression (true/false)

## Format Differences

### Java I2P Properties Format (`.properties`)

- Flat key-value pairs
- Uses prefixes like `option.i2cp.*`, `option.inbound.*`
- Standard in Java I2P router
- Example: `option.inbound.length=3`

### i2pd INI Format (`.conf`)

- Section-based structure `[TunnelName]`
- Simpler syntax, more readable
- Native to i2pd router
- Example: `inbound.length = 3`

### go-i2p YAML Format (`.yaml`)

- Hierarchical nested structure with `tunnels` map
- Most readable and maintainable
- Native to go-i2p implementation
- Supports multiple tunnels in one file
- Example:

```yaml
tunnels:
  MyTunnel:
    type: client
    port: 7000
    inbound:
      length: 3
```

## Tunnel Types Reference

| Type | Description | Requires Port | Requires Target |
|------|-------------|---------------|-----------------|
| `httpclient` | HTTP proxy client | Yes | No |
| `httpserver` | HTTP server (eepsite) | No | Yes |
| `sockstunnel` | SOCKS5 proxy | Yes | No |
| `client` | Generic client tunnel | Yes | Optional |
| `server` | Generic server tunnel | No | Yes |
| `ircclient` | IRC client tunnel | Yes | No |
| `ircserver` | IRC server tunnel | No | Yes |
| `streamrclient` | Streaming client | Yes | No |
| `streamrserver` | Streaming server | No | Yes |

## Security Considerations

- **Persistent Keys**: Set `persistentKey: true` to keep the same I2P address across restarts. This is important for servers so users can find you again.
- **Tunnel Length**: Higher values (3-7) provide better anonymity but slower performance.
- **Tunnel Quantity**: More tunnels provide better reliability and performance but use more resources.
- **Local Interface**: Binding to `127.0.0.1` ensures only local applications can use your tunnel. Never bind to `0.0.0.0` unless you understand the security implications.

## Testing Configurations

Use dry-run mode to test conversions without creating files:

```bash
go-i2ptunnel-config --dry-run httpclient.properties
go-i2ptunnel-config --dry-run --out-format yaml httpserver.conf
```

## Common Use Cases

### Setting up an eepsite

1. Use `httpserver.properties` (or `.conf`/`.yaml`)
2. Modify `target` to point to your web server (e.g., `127.0.0.1:8080`)
3. Set `persistentKey: true` so your address stays the same
4. Optionally set `spoofedHost` to a memorable `.i2p` hostname

### Browsing eepsites

1. Use `httpclient.properties` (or `.conf`/`.yaml`)
2. Keep default port `4444` or choose your own
3. Configure your browser to use `127.0.0.1:4444` as HTTP proxy
4. Visit `.i2p` addresses in your browser

### General I2P Network Access

1. Use `socks.properties` (or `.conf`/`.yaml`)
2. Keep default port `9050` or choose your own
3. Configure applications to use `127.0.0.1:9050` as SOCKS5 proxy
4. Works with SSH, IRC, and other SOCKS-compatible applications

## Further Reading

- [I2P Documentation](https://geti2p.net/en/docs)
- [i2pd Documentation](https://i2pd.readthedocs.io/)
- [go-i2p GitHub](https://github.com/go-i2p)

## Contributing

Found an issue with the examples or want to add more? Please submit a pull request or open an issue on GitHub.
