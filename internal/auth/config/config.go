// Module config retrieves the application configuration from yaml file.
// The path to yaml configuration file is determined on command line flag ("config") or in environment variable "AUTH_CONFIG_PATH".
// Priority: flag > env > default.
// Default value is empty string.
package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatabaseURL    string        `yaml:"database_url" env-required:"true"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
	ConnectTimeout time.Duration `yaml:"connect_timeout" env-default:"2s"`
	CertFile       string        `yaml:"cert_file" env-required:"true"`
	KeyFile        string        `yaml:"key_file" env-required:"true"`
	GRPC           GRPCConfig    `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// MustLoad parses the file into the configuration structure Config.
// Prefix "Must" method name means that the method does not return an error. It executes or throws panic.
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

// configPath fetches config path from command line flag or environment variable AUTH_CONFIG_PATH.
// Priority: flag > env > default.
// Default value is empty string.
func configPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("AUTH_CONFIG_PATH")
	}
	return res
}
