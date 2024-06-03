package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type ClientConfig struct {
	StoragePath  string        `yaml:"storage_path" env-required:"true"`
	GRPCAddress  string        `yaml:"grpc_address" env-required:"true"`
	WSURL        string        `yaml:"ws_url" env-required:"true"`
	QueryTimeout time.Duration `yaml:"query_timeout" env-default:"2s"`
}

func MustLoad() *ClientConfig {
	path := configPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg ClientConfig

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
		res = os.Getenv("CLIENT_CONFIG_PATH")
	}
	return res
}
