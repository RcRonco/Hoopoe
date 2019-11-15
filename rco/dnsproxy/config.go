package dnsproxy

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	LBTypeDefaultConfig = "ByOrder"
	AddressDefaultConfig = "127.0.0.1:53"
	TelemetryEnabledDefaultConfig = false
	EnableAccessLogDefaultConfig = true
	AccessLogPathDefaultConfig = "/var/log/hoopoe/access.log"
	ClientMapPathDefaultConfig = ""
	ScanAllDefaultConfig = true
	UpstreamDefaultTimeout = "5s"
)

type Config struct {
	// Server Net Config
	LBType       string           `mapstructure:"LBType"`
	RemoteHosts  []UpstreamServer `mapstructure:"UpstreamServers"`
	LocalAddress string           `mapstructure:"Address"`

	// General
	Telemetry       TelemetryConfig `mapstructure:"Telemetry"`
	AccessLog       bool            `mapstructure:"EnableAccessLog"`
	AccessLogPath   string          `mapstructure:"AccessLogPath"`
	ClientMapFile   string          `mapstructure:"ClientMapFile"`
	UpstreamTimeout string          `mapstructure:"UpstreamTimeout"`

	// Rule Config
	ScanAll bool     `mapstructure:"ScanAll"`
	Rules   []string `mapstructure:"ProxyRules"`
}

func decodeConfig() Config {
	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("failed to parse config file, %s", err)
	}
	conf.Telemetry.Enabled = conf.Telemetry.Address != ""

	return conf
}

func BuildConfig(filePath string) Config {
	log.Infof("Loading config from: %s", filePath)
	fstat, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to access config path.")
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
	viper.SetDefault("LBType", LBTypeDefaultConfig)
	viper.SetDefault("Address", AddressDefaultConfig)
	viper.SetDefault("Telemetry.Address", "")
	viper.SetDefault("EnableAccessLog", EnableAccessLogDefaultConfig)
	viper.SetDefault("AccessLogPath", AccessLogPathDefaultConfig)
	viper.SetDefault("ClientMapFile", ClientMapPathDefaultConfig)
	viper.SetDefault("ScanAll", ScanAllDefaultConfig)
	viper.SetDefault("UpstreamTimeout", UpstreamDefaultTimeout)

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	return decodeConfig()
}
