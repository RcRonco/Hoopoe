# Client Mapping
Client mapping is used to bind DNS clients to specific Upstream servers.  
Client mapping is using the ```region``` annotation of the upstream server with addition client mapping configuration file.

### Config
| Name    | Description    | Required    | Default    | Values | Examples |
|:--|:--|:-:|:-:|:-:|:--|
| networks | IP Addresses or Subnets of the clients | Yes | - | []IP Address | [example](#example) |
| region | Region of the provided networks | Yes | - | ```string``` | us-east |


### Example
```yaml
---
regions:
  - region: il
    networks:
    - "192.168.1.0/24"
  - region: us
    networks:
    - "192.168.2.0/24"
  - region: ru
    networks:
    - "192.168.3.0/24"
  - region: all
    networks:
      - "0.0.0.0/0"
```