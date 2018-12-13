# DNS Proxy
##### Description
Simple DNS Proxy for rewrites, written in Go

##### TODO:
1. <del>Add Logging and Tracing Capabilities<del> 
2. Benchmark Latency
3. <del>Adding Support for Allow and Deny (Black/White List) Rules.<del>

##### Configuration
## Configuration example for DNSProxy.
Start DNS Proxy listen on ```127.0.0.1:53``` and send to upstream server in ```8.8.8.8:53```.

#### Rules
1. Rewrite every *.com into *.co.il
2. Accepts every request for domain name that end with ```.com```
3. Blocks every request for domain and subdomain of mywebsite.com

```yaml
---
Address: "127.0.0.1"
Port: 8601
RemoteAddress: "127.0.0.1"
RemotePort: 8600
EnableStats: true
ScanAll: true
ProxyRules:
  - Type: "Rewrite"
    Pattern: ".ronco.$"
    NewPattern: ".service.consul."
  - Type: "Rewrite"
    Pattern: ".meme.$"
    NewPattern: ".query.consul."
  - Type: "Allow"
    Pattern: ".ronco.$"
  - Type: "Deny"
    Pattern: "mywebsite.com.$"
```
