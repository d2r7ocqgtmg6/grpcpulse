package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Target represents a single gRPC service to health-check.
type Target struct {
	Name     string        `yaml:"name"`
	Address  string        `yaml:"address"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	TLS      bool          `yaml:"tls"`
}

// Config holds the full daemon configuration.
type Config struct {
	ListenAddr string        `yaml:"listen_addr"`
	Targets    []Target      `yaml:"targets"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: ":9090",
	}
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := DefaultConfig()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	for i, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("config: target[%d]: name is required", i)
		}
		if t.Address == "" {
			return fmt.Errorf("config: target %q: address is required", t.Name)
		}
		if t.Interval <= 0 {
			c.Targets[i].Interval = 15 * time.Second
		}
		if t.Timeout <= 0 {
			c.Targets[i].Timeout = 5 * time.Second
		}
	}
	return nil
}
