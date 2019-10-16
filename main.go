package main

import (
	"github.com/RcRonco/Hoopoe/rco/dnsproxy"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func main() {
	configPath := flag.String("config-path", "config.yml", "Configuration file path")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors: true,
	})

	proxy := dnsproxy.NewDNSProxy(*configPath)
	log.Info("Configuration loaded successfully")

	if err := proxy.ListenAndServe(); err != nil {
		log.Errorf("%d: %s", 25, err.Error())
	}
}
