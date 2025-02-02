# go-i2ptunnel-config

Command line utility to convert I2P tunnel configurations between Java I2P, i2pd, and go-i2p formats.

## Features

- Converts between .config (Java), .conf (i2pd), and .yaml formats
- Format auto-detection
- Validation checks
- Dry-run mode
- No network connectivity required

## Install

```bash
go install github.com/go-i2p/go-i2ptunnel-config@latest
```

## Usage

Basic conversion:
```bash
go-i2ptunnel-config -in tunnel.config -out-format yaml
```

Validate only:
```bash
go-i2ptunnel-config -in tunnel.config -validate
```

Test conversion:
```bash
go-i2ptunnel-config -in tunnel.config -output-format yaml -dry-run
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