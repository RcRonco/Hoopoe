# DNS Proxy
##### Description
Simple DNS Proxy for rewrites, written in Go

##### TODO:
1. Add Logging and Tracing Capabilities 
2. Benchmark Latency
3. Adding Support for Allow and Deny (Black/White List) Rules.

##### Configuration
```json
{
  "Port": 53,
  "Address": "127.0.0.1",
  "RemotePort": 53,
  "RemoteAddress": "8.8.8.8",
  "ProxyRules": [
    // Change every query from xxx.com into xxx.co.il
    {
      "Type": "Rewrite",
      "Pattern": ".com.$",
      "NewPattern": ".co.il."
    },
    // Allow request that only end with .com
    {
      "Type": "Allow",
      "Pattern": ".com.$"
    },
    // Block mywebsite.com
    {
      "Type": "Deny",
      "Pattern": "mywebsite.com.$"
    }
  ]
}


```
