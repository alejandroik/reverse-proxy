package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   Server    `mapstructure:"server"`
	Limiters []Limiter `mapstructure:"limiters"`
}

type Server struct {
	Port       string `mapstructure:"port"`
	RemoteHost string `mapstructure:"remote_host"`
}

type Limiter struct {
	Endpoint   string     `mapstructure:"endpoint"`
	RateConfig RateConfig `mapstructure:"rate_config"`
}

type RateConfig struct {
	RateLimit       int `mapstructure:"rate_limit"`
	ClientRateLimit int `mapstructure:"client_rate_limit"`
	CleanInterval   int `mapstructure:"clean_interval"`
}

// GetConfig returns the configuration
func GetConfig(path string) *Config {
	config, err := loadConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// loadConfig loads the configuration file from the given path
func loadConfig(path string) (cfg *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	viper.SetDefault("Server.Port", "8080")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}
