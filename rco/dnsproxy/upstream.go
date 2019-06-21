package dnsproxy

import (
	"github.com/armon/go-metrics"
	"github.com/miekg/dns"
	"github.com/prometheus/common/log"
	"sync"
	"time"
)

const (
	ByOrderLB uint8 = iota
	RoundRobinLB uint8 = iota
	AllGroupName = "all"
)

type UpstreamServer struct {
	Address     string            `mapstructure:"Address"`
	Annotations map[string]string `mapstructure:"Annotations"`
}

type ServersView []*UpstreamServer

type UpstreamsManager struct {
	Servers []UpstreamServer
	LBType  uint8

	rrLB             *IndexRoundRobin
	regionMap        *RegionMap
	serversRegionMap map[string]ServersView

	Timeout time.Duration
}

func NewUpstreamsManager(servers []UpstreamServer, lbType string, regionMap *RegionMap, timeout string) *UpstreamsManager {
	usm := new(UpstreamsManager)
	usm.Servers = servers
	var err error
	usm.Timeout, err = time.ParseDuration(timeout)
	if err != nil {
		log.Fatal("Failed to parse Timeout")
	}
	if lbType == "RoundRobin" {
		usm.LBType = RoundRobinLB
		usm.rrLB = &IndexRoundRobin{
			current: 0,
			max: len(usm.Servers),
		}
	} else {
		usm.LBType = ByOrderLB
	}
	usm.regionMap = regionMap

	for _, srv := range usm.Servers {
		if region, ok := srv.Annotations["region"]; ok {
			usm.serversRegionMap[region] = append(usm.serversRegionMap[region], &srv)
		}
		// Include all Upstreams to "all" region group
		usm.serversRegionMap[AllGroupName] = append(usm.serversRegionMap[AllGroupName], &srv)
	}

	return usm
}

// Get Matching Upstream Servers
func (usm *UpstreamsManager) UpstreamSelector(req *dns.Msg, sourceIP string) (error, ServersView) {
	var region string

	// Skip region checking if region map do not exists
	if usm.regionMap == nil {
		goto allServers
	}

	// Get matching region
	region = usm.regionMap.GetRegion(sourceIP)

	// Get regional upstream servers
	if serversList, ok := usm.serversRegionMap[region]; ok {
		return nil, serversList
	} else {
		// Fallback to All server group
		goto allServers
	}

	allServers:
		return nil, usm.serversRegionMap[AllGroupName]
}

// Internal function of passing requests to the upstream DNS server
func (usm *UpstreamsManager) forwardRequest(req *dns.Msg, sourceIP string) *dns.Msg {
	startTime := time.Now()
	// Create a DNS client
	client := new(dns.Client)

	// Make a request to the upstream server
	var remoteHost string
	err, servers := usm.UpstreamSelector(req, sourceIP)
	if err != nil {
		return nil
	}

	currentTime := time.Now()
	for i :=0; currentTime.Before(startTime.Add(usm.Timeout)); i++ {
		if usm.LBType == RoundRobinLB {
			remoteHost = servers[usm.rrLB.LimitedGet(len(servers) - 1)].Address
		} else {
			remoteHost = servers[i].Address
		}
		resp, _, err := client.Exchange(req, remoteHost)
		if err != nil {
			metrics.SetGauge([]string{remoteHost, "DROPS"}, 1)
		} else if len(resp.Answer) > 0 {
			return resp
		}
		currentTime = time.Now()
	}

	return nil
}

type IndexRoundRobin struct {
	sync.Mutex

	current int
	max     int
}

func (r *IndexRoundRobin) Get() int {
	r.Lock()
	defer r.Unlock()

	if r.current >= r.max {
		r.current = r.current % r.max
	}

	result := r.current
	r.current++
	return result
}

func (r *IndexRoundRobin) LimitedGet(max int) int {
	r.Lock()
	defer r.Unlock()

	if r.current >= max {
		r.current = r.current % max
	}

	result := r.current
	r.current++
	return result
}
