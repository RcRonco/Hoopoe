package dnsproxy

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"github.com/spf13/viper"
	log "github.com/Sirupsen/logrus"
)

type ProxyRuleConfig struct {
	Rule       string `mapstructure:"Type"`
	Pattern    string `mapstructure:"Pattern"`
	NewPattern string `mapstructure:"NewPattern"`
}

type Config struct {
	// Server Net Config
	RemoteHost   string `mapstructure:"RemoteAddress"`
	RemotePort   uint16 `mapstructure:"RemotePort"`
	LocalAddress string `mapstructure:"Address"`
	LocalPort    uint16 `mapstructure:"Port"`

	// General
	StatisticsOn bool `mapstructure:"EnableStats"`

	// Rule Config
	ScanAll bool              `mapstructure:"ScanAll"`
	Rules   []ProxyRuleConfig `mapstructure:"ProxyRules"`
}

func DecodeConfig() Config {
	var conf Config
	viper.Unmarshal(&conf)
	return conf
}

func BuildConfig(file_path string) Config {
	log.Infof("Loading config from: %s", file_path)
	fstat, err := os.Stat(file_path)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to access the given config path.")
	}
	if fstat.IsDir() {
		viper.SetConfigName("config")
		viper.SetConfigType("YAML")
		viper.AddConfigPath(file_path)
	} else {
		viper.SetConfigFile(filepath.Base(file_path))
		viper.SetConfigType(strings.Replace(filepath.Ext(file_path), ".", "", 1))
	}

	viper.AddConfigPath(filepath.Dir(file_path))
	viper.SetDefault("RemotePort", 53)
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("LocalPort", 53)
	viper.SetDefault("EnableStats", false)
	viper.SetDefault("ScanAll", true)

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	return DecodeConfig()
}
