# DNS Proxy
##### Description
Simple DNS Proxy for rewrites, written in Go

##### TODO:
1. <del>Add Logging and Tracing Capabilities<del> 
2. Benchmark Latency
3. <del>Adding Support for Allow and Deny (Black/White List) Rules.<del>

##### Configuration
Configuration example for DNSProxy that accepts every domain name that end with ```.com``` (Rule number 2), Blocks every domain and subdomain of mywebsite.com and in if the request didn't get blocked also return result for xxxx.co.il instead of xxxx.com

```json
{
  "Port": 53,
  "Address": "127.0.0.1",
  "RemotePort": 53,
  "RemoteAddress": "8.8.8.8",
  "ProxyRules": [
    // "Change every query from xxx.com into xxx.co.il"
    {
      "Type": "Rewrite",
      "Pattern": ".com.$",
      "NewPattern": ".co.il."
    },
    // "Allow request that only end with .com"
    {
      "Type": "Allow",
      "Pattern": ".com.$"
    },
    // "Block mywebsite.com"
    {
      "Type": "Deny",
      "Pattern": "mywebsite.com.$"
    }
  ]
}


```
