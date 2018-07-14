package rco

import (
	"github.com/spf13/viper"
)

type ProxyRuleConfig struct {
	Rule string `mapstructure:"Type"`
	Pattern string `mapstructure:"Pattern"`
	NewPattern string `mapstructure:"NewPattern"`
}

type Config struct {
	Remote_host string `mapstructure:"RemoteAddress"`
	Remote_port uint16 `mapstructure:"RemotePort"`
	Local_addr string `mapstructure:"Address"`
	Local_port uint16 `mapstructure:"Port"`
	Rules []ProxyRuleConfig `mapstructure:"ProxyRules"`
}


func DecodeConfig() Config {
	var conf Config
	viper.Unmarshal(&conf)
	return conf
}

func BuildConfig(file_path string) Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	viper.SetDefault("RemotePort", 53)
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("LocalPort", 53)
	viper.ReadInConfig()
	return DecodeConfig()
}