package dnsproxy

import "github.com/miekg/dns"

const (
	ALLOWED int8 = 1 << iota
	BLOCKED int8 = 1 << iota
	ERROR   int8 = 1 << iota
)

type Query struct {
	Name string
	Type uint16
}

type RequestMetadata struct {
	Region    string
	IPAddress string
}

type EngineQuery struct {
	Queries []Query
	Result  int8
	dnsMsg  *dns.Msg
}

type Engine interface {
	Apply(*EngineQuery, RequestMetadata) (*EngineQuery, error)
	Name() string
}
