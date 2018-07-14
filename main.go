package main

import (
	"dns_rewriter/rco/dnsproxy"
	"github.com/golang/glog"
)

func main() {
	proxy := dnsproxy.DNSProxy{}
	proxy.InitConfig("./config.json")
	glog.Info("Configuration loaded succesfully")
	proxy.ListenAndServe()
}