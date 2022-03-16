package config

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
	SERVER_PORT         string  `mapstructure:"SERVER_PORT"`
	REMOTE_URL          string  `mapstructure:"REMOTE_URL"`

    IP_RATE_ENABLED     bool    `mapstructure:"IP_RATE_ENABLED"`
	IP_RATE_LIMIT       int     `mapstructure:"IP_RATE_LIMIT"`
    IP_BURST_LIMIT      int     `mapstructure:"IP_BURST_LIMIT"`
    IP_CLEAN_INTERVAL   int     `mapstructure:"IP_CLEAN_INTERVAL"`

    PATH_RATE_ENABLED   bool    `mapstructure:"PATH_RATE_ENABLED"`
    PATH_RATE_LIMIT     int     `mapstructure:"PATH_RATE_LIMIT"`
    PATH_BURST_LIMIT    int     `mapstructure:"PATH_BURST_LIMIT"`
    PATH_CLEAN_INTERVAL int     `mapstructure:"PATH_CLEAN_INTERVAL"`
}

func GetConfig() Configuration {
	config, err := loadConfig("./config")
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func loadConfig(path string) (config Configuration, err error) {
    viper.AddConfigPath(path)
    viper.SetConfigName("config")

    viper.AutomaticEnv()

    err = viper.ReadInConfig()
    if err != nil {
        return
    }

    err = viper.Unmarshal(&config)
    return
}