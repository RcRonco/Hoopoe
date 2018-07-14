package main

import (
	"dns_rewriter/rco"
)


func main() {
	proxy := rco.DNSProxy{}
	proxy.InitConfig()
	proxy.ListenAndServe()
}