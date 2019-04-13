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

type Config struct {
	// Server Net Config
	RemoteHosts  []string `mapstructure:"RemoteAddresses"`
	LocalAddress string   `mapstructure:"Address"`

	// General
	Telemetry     TelemetryConfig `mapstructure:"Telemetry"`
	AccessLog     bool            `mapstructure:"EnableAccessLog"`
	AccessLogPath string          `mapstructure:"AccessLogPath"`

	// Rule Config
	ScanAll bool     `mapstructure:"ScanAll"`
	Rules   []string `mapstructure:"ProxyRules"`
}

func DecodeConfig() Config {
	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("failed to parse config file, %s", err)
	}
	return conf
}

func BuildConfig(filePath string) Config {
	log.Infof("Loading config from: %s", filePath)
	fstat, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to access the given config path.")
	}
	if fstat.IsDir() {
		viper.SetConfigName("config")
		viper.SetConfigType("YAML")
		viper.AddConfigPath(filePath)
	} else {
		viper.SetConfigFile(filepath.Base(filePath))
		viper.SetConfigType(strings.Replace(filepath.Ext(filePath), ".", "", 1))
	}

	viper.AddConfigPath(filepath.Dir(filePath))
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
