package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type KubeConfig struct {
	Ns int `mapstructure:"ns"`
}

var globalConfig Config

func init() {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	if err := viper.Unmarshal(&globalConfig); err != nil {
		log.Fatal(err)
	}
}

func GetConfig() *Config {
	return &globalConfig
}
