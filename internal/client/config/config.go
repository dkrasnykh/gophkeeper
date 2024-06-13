// Module config retrieves the application configuration from yaml file.
// The path to yaml configuration file is determined on command line flag ("config") or in environment variable "AUTH_CONFIG_PATH".
// Priority: flag > env > default.
// Default value is empty string.
package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ClientConfig struct {
	StoragePath  string        `yaml:"storage_path" env-required:"true"`
	GRPCAddress  string        `yaml:"grpc_address" env-required:"true"`
	WSURL        string        `yaml:"ws_url" env-required:"true"`
	QueryTimeout time.Duration `yaml:"query_timeout" env-default:"2s"`
	CaCertFile   string        `yaml:"ca_cert_file" env-required:"true"`
}

// MustLoad parses the file into the configuration structure Config.
// Prefix "Must" method name means that the method does not return an error. It executes or throws panic.
func MustLoad() *ClientConfig {
	path := configPath()
	if path == "" {
		log.Print("config path is empty")
		os.Exit(1)
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

// configPath fetches config path from command line flag or environment variable CLIENT_CONFIG_PATH.
// Priority: flag > env > default.
// Default value is empty string.
func configPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CLIENT_CONFIG_PATH")
	}
	return res
}
