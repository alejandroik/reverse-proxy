package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
    Server      Server      `mapstructure:"server"`
    Endpoints   []Endpoint  `mapstructure:"endpoints"`
}

type Server struct {
    Port        string      `mapstructure:"port"`
    RemoteHost  string      `mapstructure:"remote_host"`
}

type Endpoint struct {
    Endpoint    string       `mapstructure:"endpoint"`
    RateConfig  RateConfig   `mapstructure:"rate_config"`
}

type RateConfig struct {
    Enabled         bool    `mapstructure:"enabled"`
    RateLimit       int     `mapstructure:"rate_limit"`
    ClientRateLimit int     `mapstructure:"client_rate_limit"`
    CleanInterval   int     `mapstructure:"clean_interval"`
}

// GetConfig returns the configuration
func GetConfig() *Config {
	config, err := loadConfig(".")
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
    viper.SetDefault("RateConfig.CleanInterval", 10)

    err = viper.ReadInConfig()
    if err != nil {
        return
    }

    err = viper.Unmarshal(&cfg)
    return
}