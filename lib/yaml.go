package i2pconv

import "gopkg.in/yaml.v2"

// Example YAML parser
func (c *Converter) parseYAML(input []byte) (*TunnelConfig, error) {
	config := &TunnelConfig{}
	err := yaml.Unmarshal(input, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Converter) generateYAML(config *TunnelConfig) ([]byte, error) {
	type wrapper struct {
		Tunnels map[string]*TunnelConfig `yaml:"tunnels"`
	}

	out := wrapper{
		Tunnels: map[string]*TunnelConfig{
			config.Name: config,
		},
	}

	return yaml.Marshal(out)
}
