package util

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
)

type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type WSConfig struct {
	MCIP string
	MCPort int
}

type Config struct {
	DB DBConfig `yaml:"database"`
	WS WSConfig `yaml:"websocket"`
}

const file = "./config.yaml"

func ReadCfg() (*Config) {
	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			if err := ioutil.WriteFile(file, []byte{}, fs.ModePerm); err != nil {
				panic(err)
			}
		}
		panic(err)
	}

	if info.IsDir() {
		panic(fmt.Sprintf("Config file '%s' is a directory", file))
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return &cfg
}
