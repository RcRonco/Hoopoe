package dnsproxy

import (
	"github.com/spf13/viper"
	"github.com/golang/glog"
	"os"
	"fmt"
	"path/filepath"
	"strings"
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
	glog.Infof("Loading config from: %s", file_path)
	fstat, err := os.Stat(file_path)
	if err != nil {
		fmt.Println(err)
		glog.Exitf("Failed to access the given config path.")
	}
	if fstat.IsDir() {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(file_path)
	} else {
		viper.SetConfigFile(filepath.Base(file_path))
		viper.SetConfigType(strings.Replace(filepath.Ext(file_path), ".", "", 1))
	}

	viper.AddConfigPath(filepath.Dir(file_path))
	viper.SetDefault("RemotePort", 53)
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("LocalPort", 53)

	err = viper.ReadInConfig()
	if err != nil {
		glog.Exitln(err)
	}
	return DecodeConfig()
}