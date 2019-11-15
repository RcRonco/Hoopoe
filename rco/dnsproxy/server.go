package dnsproxy

import (
	"github.com/armon/go-metrics"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
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

var globalConfig = Config{}

// Proxy server implementation
type DNSProxy struct {
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
	globalConfig = BuildConfig(confPath)

	// Load Region Map
	if err, d.regionMap = NewRegionMap(globalConfig.ClientMapFile); err != nil {
		log.Errorf("Failed to open client map file: %s, message: %s", globalConfig.ClientMapFile, err)
		log.Warning("Skipping Client map configuration")
	}

	// Load all engines and managers
	rulesEngine := NewRuleEngine(globalConfig.Rules)
	rulesEngine.SetScanAll(globalConfig.ScanAll)
	d.engines = append(d.engines, NewRuleEngine(globalConfig.Rules))
	d.engines = append(d.engines, NewTemplateEngine())
	d.usManager = NewUpstreamsManager(
		globalConfig.RemoteHosts,
		globalConfig.LBType,
		&d.regionMap,
		globalConfig.UpstreamTimeout,
	)

	// Init Telemetry
	d.telemetry = NewTelemetryServer(&globalConfig.Telemetry)

	// Enable Access Log
	if globalConfig.AccessLog {
		d.accessLog = log.New()
		file, err := os.OpenFile(globalConfig.AccessLogPath, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			d.accessLog.Out = file
		} else {
			log.Errorf("Failed to open log file: %s", globalConfig.AccessLogPath)
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
	d.server = &dns.Server{Addr: globalConfig.LocalAddress, Net: "udp", Handler: mux}

	// Start telemetry server, will exit immediately if telemetry is disabled
	go d.telemetry.ListenAndServe()
	log.Infof("Starting Telemetry, listening on: %s", globalConfig.Telemetry.Address)

	log.Infof("Starting server, listening on: %s", globalConfig.LocalAddress)

	// Start the DNS server
	return d.server.ListenAndServe()
}

// handle Query requests that are not PTR
func (d *DNSProxy) handleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	// Log the latency of the upstream servers
	if globalConfig.Telemetry.Enabled {
		defer metrics.MeasureSince([]string{"hoopoe", "request_latency"}, time.Now())
	}

	// Access Log
	if globalConfig.AccessLog {
		d.accessLog.Infof("%s Access Record %s", resp.RemoteAddr().String(), req.Question[0].String())
	}

	// Copy the message for applying rulesEngine on it
	reply, err := d.processMsg(resp, req)
	handleError(err, 107)
	if reply == nil {
		d.returnBlocked(resp, req)
	} else {
		respMsg := d.buildResponseMsg(req, reply.dnsMsg)
		err = resp.WriteMsg(respMsg)
		handleError(err, 114)
	}
}

// process message by applying Engines
func (d *DNSProxy) processMsg(resp dns.ResponseWriter, req *dns.Msg) (*EngineQuery, error) {
	// Copy the message for applying rulesEngine on it
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)
	upstreamMsg.Question = []dns.Question{}
	metadata := RequestMetadata{
		Region: d.regionMap.GetRegion(strings.Split(resp.RemoteAddr().String(), ":")[0]),
		IPAddress: resp.RemoteAddr().String(),
	}

	// Only one question supported
	query := req.Question[0]
	engineQuery := new(EngineQuery)
	engineQuery.dnsMsg = req
	engineQuery.Queries = append(
		engineQuery.Queries, Query{
			Name: query.Name,
			Type: query.Qtype,
		})
	// Run on each registered Engine
	for _, engine := range d.engines {
		// Process query with current Engine
		engineQuery, err := engine.Apply(engineQuery, metadata)
		if err != nil {
			return nil, err
		}
		// Check if engine return that this query need to be blocked
		if engineQuery.Result == BLOCKED {
			// Access Log
			if globalConfig.AccessLog {
				d.accessLog.Infof(
					"%s: BLOCKED - Record %s",
					engine.Name(),
					resp.RemoteAddr().String(),
					req.Question[0].String(),
				)
			}
			return nil, nil
		}
	}

	engineQuery, err := d.usManager.Apply(engineQuery, metadata)
	if err != nil {
		return nil, err
	}

	return engineQuery, nil
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

	metadata := RequestMetadata{
		Region: d.regionMap.GetRegion(strings.Split(resp.RemoteAddr().String(), ":")[0]),
		IPAddress: resp.RemoteAddr().String(),
	}

	// Send it to the upstream server
	reply := d.usManager.forwardRequest(upstreamMsg, metadata)

	// Build response and send it
	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Answer = reply.Answer
	err := resp.WriteMsg(respMsg)
	handleError(err, 111)
}