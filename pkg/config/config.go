package config

import (
	"buding-kube/pkg/logs"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Port      int    `mapstructure:"port"`
	JwtSecret string `mapstructure:"jwtSecret"`
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
		logs.Info("👉未能读取到有效的config,将会以默认模式启动")
		return
	}

	if err := viper.Unmarshal(&globalConfig); err != nil {
		log.Fatal("转换配置失败，将会以默认模式启动 😏")
	}
}

func GetConfig() *Config {
	return &globalConfig
}
