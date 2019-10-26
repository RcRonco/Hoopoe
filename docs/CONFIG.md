```yaml
layout: configuration
title: "Configuration"
permalink: /CONFIG/
```
## Configuration file

Configuration Options for the YAML configuration file.

#### Config
| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| Address | Listening IP Address and Port | No | ```127.0.0.1:53``` | IP Address | 192.168.1.5:53 |
| UpstreamServers | Remote DNS Servers | Yes | - | [[]UpstreamServer](#upstreamserver) | [example](#example) |
| Telemetry | Telemetry configuration | Yes | - | [Telemtry](#telemetry) | [example](#example) |
| EnableAccessLog | Access log enabled  | No | ```True``` | ```bool``` | ```True``` |
| AccessLogPath | Access log file path **can cause performance degradation** | No | ```/var/log/hoopoe/access.log``` | POSIX file path | ```/tmp/access.log``` |
| ClientMapFile | file path to ClientMapping | No | - | POSIX file path | ```/tmp/clientmap.yml``` |
| ScanAll | Enable ScallAll mode, which will apply all rewrite rules on query instead of the first one to match **can cause performance degration** | No | ```true``` | ```true/false```| ``` false``` | 
| ProxyRules | Rules that will Rewrite/Deny/Allow/Pass the query  | Yes | - | ```[]string``` | Check the example below |

#### UpstreamServer
| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| Address | IP Address and Port of the Upstream Server | Yes | - | IP Address | 192.168.1.5:53 |
| Annotations | map of metadata about the Upstream Server | No | - | ```map[string]string``` | [example](#example) |

###### Annotations:
```region``` - Will be mapped to region of client mapping feature    
```domain``` - Will be mapped to domain mapping **(Not impelemented yet)**

#### Telemetry
| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| Enabled | Enable Performance stats, **can cause performance degradation** | No | ```false``` | ```true/false```| ``` true``` |
| Address | Stats HTTP address, **can cause performance degradation** | No | ```127.0.0.1:8080``` | IP Address and Port | ``` 0.0.0.0:80``` |

## Example
```yaml
---
Address: "0.0.0.0:8601"
LBType: RoundRobin
UpstreamServers:
  - Address: "192.1.1.1:53"
    Annotations:
      region: "il"
      domain: "co.il"
  - Address: "8.8.4.4:53"
    Annotations:
      region: "us"
      domain: "com"
ClientMapFile: clientMap.yml
Telemetry:
  Enabled: true
  Address: "0.0.0.0:8080"
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
