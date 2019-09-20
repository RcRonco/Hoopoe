package dnsproxy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/miekg/dns"
	"os"
	"strings"
	"time"
)

// Handle errors that will not cause crash of the system.
func handleError(e error, ln int) {
	if e != nil {
		log.Errorf("%d: %s", ln, e.Error())
	}
}

// Proxy server implementation
type DNSProxy struct {
	config    Config

	accessLog *log.Logger
	telemetry *TelemetryServer
	server    *dns.Server

	engines		   []Engine
	usManager      *UpstreamsManager
	regionMap      RegionMap
}

func NewDNSProxy(configPath string) *DNSProxy {
	proxy := new(DNSProxy)
	proxy.Init(configPath)

	return proxy
}

// Initialize the config of the DNSProxy from json file
func (d *DNSProxy) Init(confPath string) {
	var err error
	// Load the config from json file
	d.config = BuildConfig(confPath)

	// Load Region Map
	if err, d.regionMap = NewRegionMap(d.config.ClientMapFile); err != nil {
		log.Errorf("Failed to open client map file: %s", d.config.AccessLogPath)
		log.Warning("Skipping Client map configuration")
	}

	// Load all engines and managers
	rulesEngine := NewRuleEngine(d.config.Rules)
	rulesEngine.SetScanAll(d.config.ScanAll)
	d.engines = append(d.engines, NewRuleEngine(d.config.Rules))
	d.engines = append(d.engines, NewTemplateEngine())
	d.usManager = NewUpstreamsManager(d.config.RemoteHosts, d.config.LBType, &d.regionMap, d.config.UpstreamTimeout)

	// Enable Telemetry
	if d.config.Telemetry.Address == "" {
		d.telemetry = NewTelemetryServer(&d.config.Telemetry)
		d.telemetry.Init()
	}

	// Enable Access Log
	if d.config.AccessLog {
		d.accessLog = log.New()
		file, err := os.OpenFile(d.config.AccessLogPath, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			d.accessLog.Out = file
		} else {
			log.Errorf("Failed to open log file: %s", d.config.AccessLogPath)
		}
	}
}

// Bind to port and start handle DNS requests
func (d *DNSProxy) ListenAndServe() error {
	// Set handlers
	mux := dns.NewServeMux()
	mux.HandleFunc("arpa.", d.handlePtr)
	mux.HandleFunc(".", d.handleQuery)

	// Build the DNS server
	d.server = &dns.Server{Addr: d.config.LocalAddress, Net: "udp", Handler: mux}
	log.Infof("Starting server, listening on: %s", d.config.LocalAddress)

	// Start telemetry server, will exit immediately if telemetry is disabled
	go d.telemetry.ListenAndServe()

	// Start the DNS server
	return d.server.ListenAndServe()
}

// handle Query requests that are not PTR
func (d *DNSProxy) handleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	// Log the latency of the upstream servers
	if d.config.Telemetry.Address != "" {
		defer metrics.MeasureSince([]string{"hoopoe", "request_latency"}, time.Now())
	}
	// Access Log
	if d.config.AccessLog {
		d.accessLog.Infof("%s Access Record %s", resp.RemoteAddr().String(), req.Question[0].String())
	}

	// Copy the message for applying rulesEngine on it
	upstreamMsg := d.buildUpstreamMsg(resp, req)
	if upstreamMsg == nil {
		return
	}

	// Make a request to the upstream server
	reply := d.forwardRequest(upstreamMsg, resp.RemoteAddr().String())

	respMsg := d.buildResponseMsg(req, reply)
	err := resp.WriteMsg(respMsg)
	handleError(err, 189)
}

// build upstream message by applying Proxy rulesEngine
func (d *DNSProxy) buildUpstreamMsg(resp dns.ResponseWriter, req *dns.Msg) *dns.Msg {
	// Copy the message for applying rulesEngine on it
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)
	upstreamMsg.Question = []dns.Question{}
	metadata := RequestMetadata{
		Region: d.regionMap.GetRegion(strings.Split(resp.RemoteAddr().String(), ":")[0]),
		IPAddress: resp.RemoteAddr().String(),
	}

	// Run on every query entry in the request
	for _, query := range req.Question {
		engineQuery := new(EngineQuery)
		engineQuery.Queries = append(engineQuery.Queries, Query{
			Name: query.Name,
			Type: ARecordType,
		})
		for _, engine := range d.engines {
			engineQuery, err := engine.Apply(engineQuery, metadata)
			if err != nil {
				return nil
			}
			if engineQuery.Result == BLOCKED {
				d.returnBlocked(resp, req)
				// Access Log
				if d.config.AccessLog {
					d.accessLog.Infof(
						"%s: BLOCKED - Record %s",
						engine.Name(),
						resp.RemoteAddr().String(),
						req.Question[0].String(),
					)
				}
				return nil
			}
		}

		// TODO: Add sending to upstream_engine result for final build of upstream
		// Append new Question the the message
		rewrittenQuery := dns.Question{ Name: name, Qtype: query.Qtype, Qclass: query.Qclass}
		upstreamMsg.Question = append(upstreamMsg.Question, rewrittenQuery)
	}

	return upstreamMsg
}

// Internal function of passing requests to the upstream DNS server
func (d *DNSProxy) forwardRequest(req *dns.Msg, sourceIP string) *dns.Msg {
	// Profiling the latency of the upstream servers
	if d.config.Telemetry.Address != "" {
		defer metrics.MeasureSince([]string{"hoopoe", "request_latency"}, time.Now())
	}

	return d.usManager.forwardRequest(req, sourceIP)
}

// Build response message for server message
func (d *DNSProxy) buildResponseMsg(clientRequest *dns.Msg, upstreamReply *dns.Msg) *dns.Msg {
	respMsg := new(dns.Msg)
	respMsg.SetReply(clientRequest)

	if upstreamReply != nil {
		// Set the original name in the response
		for index, q := range clientRequest.Question {
			upstreamReply.Answer[index].Header().Name = q.Name
		}
		respMsg.Answer = upstreamReply.Answer
	}

	return respMsg
}

// Build and send REFUSED response message to client
func (d *DNSProxy) returnBlocked(resp dns.ResponseWriter, req *dns.Msg) {
	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Rcode = dns.RcodeRefused
	err := resp.WriteMsg(respMsg)
	handleError(err, 205)
}

// handle PTR records
// Currently PTR records rulesEngine are not supported
func (d *DNSProxy) handlePtr(resp dns.ResponseWriter, req *dns.Msg) {
	// Build new DNS message
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)

	// Send it to the upstream server
	reply := d.forwardRequest(upstreamMsg, resp.RemoteAddr().String())

	// Build response and send it
	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Answer = reply.Answer
	err := resp.WriteMsg(respMsg)
	handleError(err, 111)
}

// Internal function of passing requests to the upstream DNS server
func (d *DNSProxy) forwardRequestImpl(req *dns.Msg) *dns.Msg {
	// Create a DNS client
	client := new(dns.Client)

	// Make a request to the upstream server
	for _, remoteHost := range d.config.RemoteHosts {
		resp, _, err := client.Exchange(req, remoteHost.Address)
		if err != nil && d.config.Telemetry.Address != ""  {
			metrics.IncrCounterWithLabels([]string{"hoopoe_upstream_DROPS"}, 1, []metrics.Label{
				{
					Name:  "upstream_address",
					Value: remoteHost.Address,
				},
			})
		} else if len(resp.Answer) > 0 {
			return resp
		}
	}

	return nil
}