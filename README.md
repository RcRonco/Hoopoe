# Hoopoe - DNS Proxy
![test image size](docs/data/hoopoe-small.png)  
Hoopoe (pronounced Hu-pu as the [bird](https://en.wikipedia.org/wiki/Hoopoe)) is a simple DNS Proxy for rewrites, written in Go

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
| Address | Listening IP Address and Port | No | ```127.0.0.1:53``` | IP Address | 192.168.1.5:53 |
| RemoteAddresses | Remote DNS Servers | No | ```[127.0.0.1:8600, 1.1.1.1:53]``` | IP Address | 8.8.8.8:53 |
| Telemetry.Enabled | Enable Performance stats, **can cause performance degradation** | No | ```false``` | ```true/false```| ``` true``` |
| Telemetry.Address | Stats HTTP address, **can cause performance degradation** | No | ```127.0.0.1:8080``` | IP Address and Port | ``` 0.0.0.0:80``` |
| EnableAccessLog | Access log enabled  | No | ```True``` | ```bool``` | ```True``` |
| AccessLogPath | Access log file path **can cause performance degradation** | No | ```/var/log/hoopoe/access.log``` | POSIX file path | ```/tmp/access.log``` |
| ScanAll | Enable ScallAll mode, which will apply all rewrite rules on query instead of the first one to match **can cause performance degration** | No | ```true``` | ```true/false```| ``` false``` | 
| ProxyRules | Rules that will Rewrite/Deny/Allow/Pass the query  | Yes | - | ```[]string``` | Check the example below |

#### Telemetry
Hoopoe support exporting performance telemetry to prometheus with HTTP endpoint.    
Hoopoe metrics access by the ```/metrics``` endpoint.

#### Proxy Rules
Rules defined in configuration file, their job is to Act in predefined action for specified DNS query.  
Every rule must be in the format: ```TYPE ACTION PATTERN OPTIONS```
Currently the are 4 types of rules supported.  
* ```Pass``` - A rule is set for every query that the pattern matching to, will passed without any other rule type.
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None     
* ```Allow``` - A Whitelist rule, any request that not match any ```Allow``` rule will be **DROPPED**.    
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None     
* ```Deny``` - A Blacklist rule, any request that match one of the ```Deny``` rule will be **DROPPED**.   
    When ```Allow``` rule is also defined the Deny rule is used to block specific query inside the Whitelist query space.    
  **Parameters**:
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: None     
* ```Rewrite``` - This rule used to edit the query before it arriving the Remote DNS Server.    
  **Parameters**:   
    * **Action**: ```All string matching actions```   
    * **Pattern**: ```string```
    * **Options**: 
        * Replacement: ```string``` - string to replace pattern with.
        
Currently the type of actions are ```String Matching```:
* **PREFIX**: Matching the prefix of string with ```Pattern```.
* **SUFFIX**: Matching the suffix of string with ```Pattern```.
* **SUBSTRING**: Will match if string contains ```Pattern```.
* **REGEXP**: Will match if string matches regexp ```Pattern```.
      
## Example
Start DNS Proxy listen on ```127.0.0.1:8601``` and send to upstream server in ```8.8.8.8:53```.

```yaml
---
---
Address: "127.0.0.1:8601"
RemoteAddresses:
  - "8.8.8.8:53"
Telemetry:
  Enabled: true
  Address: "0.0.0.0:80"
AccessLogPath: access.log
ScanAll: true
ProxyRules:
  # Will Rewrite every query starting with mail to start with www
  - Rewrite PREFIX mail www
  # Will allow only queries ends with .com and .co.il
  - Allow SUFFIX .com
  - Allow SUFFIX .co.il
  # Will deny every query end with youtube.com
  - Deny SUFFIX youtube.com
  # Will pass with out any processing every query equal to www.youtube.com
  - Pass REGEXP www.youtube.com
  # Will rewrite every query ends with .co.il to .com and myorg.com to service.consul
  - Rewrite SUFFIX co.il com
  - Rewrite SUFFIX myorg.com service.consul
```

## TODO:
* [x] - Add Logging and Tracing Capabilities  
* [x] - Benchmark Latency
* [x] - Adding Support for Allow and Deny (Black/White List) Rules.
* [x] - Add Access log
* [ ] - Rebrand the project :/
