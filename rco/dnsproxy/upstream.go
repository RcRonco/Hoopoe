package dnsproxy

import (
	"context"
	"fmt"
	"github.com/armon/go-metrics"
	"github.com/miekg/dns"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
	"net"
	"os"
	"sync"
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

	rrLB      *IndexRoundRobin
	regionMap map[string]ServersView

	clientMap []ClientsSubnet
}

type ClientsSubnet struct {
	Network net.IPNet `yaml:"network"`
	Region  string `yaml:"region"`
}

func NewUpstreamsManager(servers []UpstreamServer, lbType string, clientMappingFile string) *UpstreamsManager {
	usm := new(UpstreamsManager)
	usm.Servers = servers
	if lbType == "RoundRobin" {
		usm.LBType = RoundRobinLB
		usm.rrLB = &IndexRoundRobin{
			current: 0,
			max: len(usm.Servers),
		}
	} else {
		usm.LBType = ByOrderLB
	}
	cmInitialized := false
	for _, srv := range usm.Servers {
		if region, ok := srv.Annotations["region"]; ok {
			if !cmInitialized {
				if err := usm.loadSubnetMap(clientMappingFile); err != nil {
					log.Errorf("%s", err)
					return nil
				}
				cmInitialized = true
			}
			usm.regionMap[region] = append(usm.regionMap[region], &srv)

		}
		// Include all Upstreams to all region group
		usm.regionMap[AllGroupName] = append(usm.regionMap[AllGroupName], &srv)
	}

	return usm
}

// Get client map file and parse it
func (usm *UpstreamsManager) loadSubnetMap(path string) error {
	var err error
	data := make([]byte, 4)
	// Read the client map file
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)

	if err != nil {
		goto loadSubnetError
	}
	if _, err = file.Read(data); err != nil {
		goto loadSubnetError
	}

	// Parse the YAML file
	if err = yaml.Unmarshal(data, usm.clientMap); err != nil {
		goto loadSubnetError
	}
	return nil

loadSubnetError:
	return fmt.Errorf("failed to create client map: %s", err)
}

func (usm *UpstreamsManager) GetRegion(ip string) (error, string) {
	// If client map is empty return all servers
	if len(usm.clientMap) < 1 {
		return nil, AllGroupName
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return fmt.Errorf("faild to parse IP: %s", ip), ""
	}
	// find a matching region that the ip fitting in the network
	for _, clientSubnet := range usm.clientMap {
		if clientSubnet.Network.Contains(ipAddr) {
			return nil, clientSubnet.Region
		}
	}

	return nil, AllGroupName
}

// Get Matching Upstream Servers
func (usm *UpstreamsManager) UpstreamSelector(req *dns.Msg, sourceIP string) (error, ServersView) {
	// Get matching region
	if err, region := usm.GetRegion(sourceIP); err == nil {
		return err, nil
	} else {
		// Get regional upstream servers
		if serversList, ok := usm.regionMap[region]; ok {
			return nil, serversList
		} else {
			// Fallback to All server group
			return nil, usm.regionMap[AllGroupName]
		}
	}
}

// Internal function of passing requests to the upstream DNS server
func (usm *UpstreamsManager) forwardRequest(ctx context.Context, req *dns.Msg, sourceIP string) *dns.Msg {
	// Create a DNS client
	client := new(dns.Client)

	// Make a request to the upstream server
	var remoteHost string
	err, servers := usm.UpstreamSelector(req, sourceIP)
	if err != nil {
		return nil
	}

	_, ok := ctx.Deadline()
	for i :=0; ok; i++ {
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
