package dnsproxy

import "github.com/miekg/dns"

const (
	PTRRecordType       uint8 = 0
	ARecordType         uint8 = 1
	AFallbackRecordType uint8 = 2
)

const (
	ALLOWED int8 = 1 << iota
	BLOCKED int8 = 1 << iota
	ERROR   int8 = 1 << iota
)

type Query struct {
	Name string
	Type uint8
}

type RequestMetadata struct {
	Region    string
	IPAddress string
}

type EngineQuery struct {
	Queries    []Query
	Result     int8
	originRequest *dns.Msg
}

type Engine interface {
	Apply(*EngineQuery, RequestMetadata) (*EngineQuery, error)
	Name() string
}
