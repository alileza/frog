package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DSN     string   `yaml:"datasource"`
	Targets []string `yaml:"targets"`
}

func Retrieve(configFile string) (*Config, error) {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, errors.Wrapf(err, "unmarshal yaml : %s", string(b))
	}

	return &cfg, nil
}
