package dnsproxy

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type TelemetryConfig struct {
	Enabled bool   `mapstructure:"Enabled"`
	Address string `mapstructure:"Address"`
}

type ProxyRuleConfig struct {
	Rule       string `mapstructure:"Type"`
	Pattern    string `mapstructure:"Pattern"`
	NewPattern string `mapstructure:"NewPattern"`
}

type Config struct {
	// Server Net Config
	RemoteHosts  []string `mapstructure:"RemoteAddresses"`
	LocalAddress string   `mapstructure:"Address"`

	// General
	Telemetry  		TelemetryConfig	`mapstructure:"Telemetry"`
	AccessLog     	bool			`mapstructure:"EnableAccessLog"`
	AccessLogPath	string			`mapstructure:"AccessLogPath"`

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
	viper.SetDefault("Address", "127.0.0.1:53")
	viper.SetDefault("Telemetry.Enabled", false)
	viper.SetDefault("Telemetry.Address", "127.0.0.1:8080")
	viper.SetDefault("EnableAccessLog", true)
	viper.SetDefault("AccessLogPath", "/var/log/hopoe/access.log")
	viper.SetDefault("ScanAll", true)

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	return DecodeConfig()
}
