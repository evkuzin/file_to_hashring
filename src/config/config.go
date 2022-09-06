package config

import (
	"go.uber.org/zap"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers []string `yaml:"servers"`
	State   string   `yaml:"state"`
	Logger  *zap.Config
}

func NewConfig(f string) (*Config, error) {
	rawConf, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(rawConf, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
