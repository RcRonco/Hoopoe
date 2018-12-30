# Hoopoe - DNS Proxy
## Description
Simple DNS Proxy for rewrites, written in Go

### Install and Run
###### Build

```shell
go get github.com/RcRonco/Hoopoe
cd $GOPATH/src/github.com/RcRonco/Hoopoe
go build -o hoopoe main.go
```

###### Install

```shell
cp ./hoopoe /usr/local/bin/hoopoe
mkdir /etc/hoopoe.d
cp $GOPATH/src/github.com/RcRonco/Hoopoe/config.yml.example /etc/hoopoe.d/config.yml
```
* Edit ```config.yml``` for your need.

###### Run
```shell
hoopoe --config-path=/etc/hoopoe.d/config.yml
```

### Flags
| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| --config-path | Configuration folder | No | ```./config.yml``` | POSIX-PATH format | --config-path=/etc/hoopoe.d/config.yml |
  
## Configuration
Configuration Options for the YAML configuration file.

| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| Address | Listening IP Address | No | ```127.0.0.1``` | IP Address | 192.168.1.5 |
| Port | Listening UDP Port | No | ```53``` | 1-65535 | 12021 |
| RemoteAddress | Remote DNS Server IP Address | No | ```127.0.0.1``` | IP Address | 8.8.8.8 |
| RemotePort | Remote DNS Server Port | No | ```53``` | 1-65535 | 8600 |
| Telemetry.Enabled | Enable Performance stats, **can cause performance degradation** | No | ```false``` | ```true/false```| ``` true``` |
| Telemetry.Address | Stats HTTP address, **can cause performance degradation** | No | ```127.0.0.1:8080``` | IP Address and Port | ``` 0.0.0.0:80``` |
| ScanAll | Enable ScallAll mode, which will apply all rewrite rules on query instead of the first one to match **can cause performance degration** | No | ```true``` | ```true/false```| ``` false``` | 
| ProxyRules | Rules that will Rewrite/Deny/Allow/Pass the query  | Yes | - | ```[]{Type, Pattern, NewPattern}``` | Check the example below |

#### Proxy Rules
Currently the are 4 types of rules supported.
* ```Pass``` - A rule is set for every query that the pattern matching to, will passed without any other rule type.
  Parameters:    
    * Type: ```Pass/p```   
    * Pattern: ```Regexp```     
* ```Allow``` - A Whitelist rule, any request that not match any ```Allow``` rule will be **DROPPED**.    
  Parameters:   
    * Type: ```Allow/a```    
    * Pattern: ```Regexp```   
* ```Deny``` - A Blacklist rule, any request that match one of the ```Deny``` rule will be **DROPPED**.   
    When ```Allow``` rule is also defined the Deny rule is used to block specific query inside the Whitelist query space.    
  Parameters:
    * Type: ```Deny/d```    
    * Pattern: ```Regexp```   
* ```Rewrite``` - This rule used to edit the query before it arriving the Remote DNS Server.    
  Parameters:   
    * Type: ```Rewrite```/```rw```    
    * Pattern: ```Regexp```   
    * NewPattern: ```Regexp```   
      
## Example
Start DNS Proxy listen on ```127.0.0.1:53``` and send to upstream server in ```8.8.8.8:53```.

###### Rules
1. Rewrite every *.com into *.co.il
2. Accepts every request for domain name that end with ```.com```
3. Blocks every request for domain and subdomain of mywebsite.com

```yaml
---
Address: "127.0.0.1:8601"
RemoteAddresses:
  - "127.0.0.1:8600"
  - "8.8.8.8:53"
Telemetry:
  Enabled: true
  Address: "0.0.0.0:80"
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

## TODO:
* [x] - Add Logging and Tracing Capabilities  
* [x] - Benchmark Latency
* [x] - Adding Support for Allow and Deny (Black/White List) Rules.
* [ ] - Add Caching for Rewrite
* [ ] - Add Access log
* [ ] - Add support for cobra cmd
* [ ] - Rebrand the project :/
