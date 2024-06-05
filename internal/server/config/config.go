package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatabaseURL  string        `yaml:"database_url" env-required:"true"`
	QueryTimeout time.Duration `yaml:"query_timeout" env-default:"2s"`
	CertFile     string        `yaml:"cert_file" env-required:"true"`
	KeyFile      string        `yaml:"key_file" env-required:"true"`
	Key          string        `yaml:"key" env-required:"true"`
	WS           WSConfig      `yaml:"ws"`
}

type WSConfig struct {
	Address string `yaml:"address"`
}

func MustLoad() *Config {
	path := configPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func configPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("SERVER_CONFIG_PATH")
	}
	return res
}
