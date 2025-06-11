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

var GlobalConfig Config

func init() {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		log.Fatal(err)
	}
}

func GetConfig() *Config {
	return &GlobalConfig
}
