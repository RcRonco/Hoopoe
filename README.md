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
  "Port": 8053,
  "Address": "127.0.0.1",
  "RemotePort": 53,
  "RemoteAddress": "8.8.8.8",
  "ProxyRules": [
    {
      "Type": "Rewrite",
      "Pattern": ".com.$",
      "NewPattern": ".co.il."
    }
  ]
}
```
