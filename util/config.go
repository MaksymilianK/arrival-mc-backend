package util

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type DBConfig struct {
	Host string
	Port int
	Database string
	User string
	Password string
}

type GameConnConfig struct {
	IP string
	Port int
}

type Config struct {
	DB DBConfig
	GameConn GameConnConfig `yaml:"game-connection"`
}

func ReadCfg() (*Config, error) {
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}