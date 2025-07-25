package config

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config map[string][]string

func LoadConfig(path string) (Config, error) {
	// reads Config file and returns a Config map
	// path => YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
