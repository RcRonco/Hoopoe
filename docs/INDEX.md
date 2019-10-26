# Hoopoe - DNS Proxy
![Hoopoe](data/hoopoe-small.png)  
Hoopoe (pronounced Hu-pu as the [bird](https://en.wikipedia.org/wiki/Hoopoe)) is a simple DNS Proxy for rewrites, written in Go

## Content
* [Getting Started](#install-and-run)
* [Configuration](CONFIG.md)
* [Rules](RULES.md)
* [client_mapping](CLIENT_MAPPING.md)
* [Internals](INTERNAL.md)

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

## TODO:
* [ ] - Refactor metrics
* [ ] - Rewrite docs
* [ ] - Restructure the project
