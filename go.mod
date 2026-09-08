module github.com/go-i2p/go-i2ptunnel-config

go 1.26.0

require (
	github.com/go-i2p/i2pkeys v0.33.92
	github.com/magiconair/properties v1.18.11
	github.com/urfave/cli/v2 v2.27.7
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

retract (
	v0.1.59999
	v0.1.5999
	v0.1.599
)
