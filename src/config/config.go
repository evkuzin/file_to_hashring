package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers []string `yaml:"servers"`
}

func NewConfig(f string) (*Config, error) {
	rawConf, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	conf := Config{}
	err = yaml.Unmarshal(rawConf, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
