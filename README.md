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

Validate only:
```bash
go-i2ptunnel-config --validate tunnel.config
```

Test conversion (dry-run):
```bash
go-i2ptunnel-config --dry-run tunnel.config
```

## Contributing

1. Fork repository
2. Create feature branch
3. Run `make fmt`
4. Submit pull request

## Security

- No network connectivity
- No private key handling
- Configuration files only

## License

MIT License