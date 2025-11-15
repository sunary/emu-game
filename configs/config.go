package configs

import (
	"bytes"
	_ "embed"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all settings
//
//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	Server Server      `yaml:"server" mapstructure:"server"`
	Redis  RedisConfig `yaml:"redis" mapstructure:"redis"`
}

type Server struct {
	Addr string `yaml:"addr" mapstructure:"addr"`
}

// RedisConfig ...
type RedisConfig struct {
	Addr     string `yaml:"addr" mapstructure:"addr"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

func Load() *Config {
	var cfg = &Config{}

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(defaultConfig))
	if err != nil {
		log.Fatalf("failed to read viper config: %v", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return cfg
}
