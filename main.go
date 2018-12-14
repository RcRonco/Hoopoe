package main

import (
	flag  "github.com/spf13/pflag"
	log "github.com/Sirupsen/logrus"
	"github.com/RcRonco/dns_proxy/rco/dnsproxy"
)

// TODO: Fix EnableStats option ignored

func main() {
	configPath := flag.String("config-path", "config.yml", "Configuration file path")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors: true,
	})

	proxy := dnsproxy.DNSProxy{}
	proxy.InitConfig(*configPath)
	log.Info("Configuration loaded succesfully")
	proxy.ListenAndServe()
}