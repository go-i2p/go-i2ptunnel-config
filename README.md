# go-i2ptunnel-config

Command line utility to convert I2P tunnel configurations between Java I2P, i2pd, and go-i2p formats.

## Features

- Converts between .config (Java), .conf (i2pd), and .yaml formats
- Format auto-detection
- Validation checks
- Dry-run mode
- SAM (Simple Anonymous Messaging) integration with I2P key management
- No network connectivity required

## Install

```bash
go install github.com/go-i2p/go-i2ptunnel-config@latest
```

## Usage

Basic conversion with auto-detection:
```bash
go-i2ptunnel-config tunnel.config
```

Specify output format:
```bash
go-i2ptunnel-config --out-format ini tunnel.config
```

Specify custom output file:
```bash
go-i2ptunnel-config -o custom-name.yaml tunnel.config
go-i2ptunnel-config --output /path/to/output.conf tunnel.properties
```

Validate only:
```bash
go-i2ptunnel-config --validate tunnel.config
```

Test conversion (dry-run):
```bash
go-i2ptunnel-config --dry-run tunnel.config
```

Batch process multiple files:
```bash
go-i2ptunnel-config --batch "*.config"
go-i2ptunnel-config --batch --out-format ini "tunnels/*.properties"
```

Override input format detection (useful for non-standard extensions or stdin):
```bash
go-i2ptunnel-config --in-format properties --out-format yaml tunnel.txt
go-i2ptunnel-config --in-format ini --out-format yaml tunnel.txt
```

Read from stdin (pass `-` as the input file; `--in-format` is required):
```bash
cat tunnel.properties | go-i2ptunnel-config --in-format properties --out-format yaml -
go-i2ptunnel-config --in-format yaml --dry-run -  # preview YAML piped on stdin
```

List tunnel names in a multi-tunnel file without converting:
```bash
go-i2ptunnel-config --list-tunnels tunnels.conf
```

Split a multi-tunnel file into one output file per tunnel:
```bash
go-i2ptunnel-config --split tunnels.conf
go-i2ptunnel-config --split --out-format properties tunnels.conf
go-i2ptunnel-config --split --dry-run tunnels.conf   # preview without writing
```

Generate or load SAM I2P keys alongside conversion:
```bash
go-i2ptunnel-config --sam tunnel.yaml
go-i2ptunnel-config --sam --keystore ~/.i2p/keys/ tunnel.yaml
```

## Examples

The `examples/` directory contains ready-to-use configuration templates for common tunnel types in all three formats:

- HTTP client (web proxy)
- HTTP server (eepsite hosting)
- SOCKS proxy
- Generic client tunnel
- Generic server tunnel

Each example includes detailed comments explaining the configuration options. See [examples/README.md](examples/README.md) for more information.

## Contributing

1. Fork repository
2. Create feature branch
3. Run `make fmt`
4. Submit pull request

## Limitations

- **Single-tunnel conversion**: Each invocation converts one tunnel. When an input file contains multiple tunnel definitions (e.g., a multi-section i2pd `tunnels.conf` or a YAML file with several entries under `tunnels:`), only the first tunnel is converted by default. A warning is printed to stderr. Use `--split` to write each tunnel to its own output file, `--list-tunnels` to preview which tunnels are present, or `--batch` to convert a collection of single-tunnel files at once.

## Security

- No network connectivity
- No private key handling
- Configuration files only

## License

MIT License