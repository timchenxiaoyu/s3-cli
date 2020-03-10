package main

import (
	"github.com/go-ini/ini"
	"github.com/urfave/cli"
	"os"
	"path"
)

type Config struct {
	AccessKey string `ini:"access_key"`
	SecretKey string `ini:"secret_key"`

	HostBase   string `ini:"host_base"`
	HostBucket string `ini:"host_bucket"`

	Region   string `ini:"region"`
	UseHttps bool   `ini:"use_https"`
}

func NewConfig(c *cli.Context) (*Config, error) {
	var cfgPath string

	if value := os.Getenv("HOME"); len(value) > 0 {
		cfgPath = path.Join(value, ".s3cfg")
	} else {
		cfgPath = ".s3cfg"
	}

	config, err := loadConfigFile(cfgPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func loadConfigFile(path string) (*Config, error) {
	config := &Config{}

	cfg, err := ini.Load(path)
	if err != nil {
		return config, err
	}
	if err := cfg.Section("default").MapTo(config); err != nil {
		return nil, err
	}
	return config, nil
}
