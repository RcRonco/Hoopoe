package dnsproxy

import (
	"fmt"
	"time"
	"github.com/miekg/dns"
	log "github.com/Sirupsen/logrus"
)

// Handle errors that will not cause crash of the system.
func handleError(e error, ln int) {
	if e != nil {
		log.Errorf("%d: %s", ln, e.Error())
	}
}

type Stats struct {
	TotalRequests int64
	TotalLatency time.Duration
	UpstreamLatency time.Duration
	StartTime time.Time
}

// Proxy server implementation
type DNSProxy struct {
	config Config
	server *dns.Server
	rules  RuleEngine
	stats  Stats
}

func (s *Stats) logStats() {
	totalAvgLatency := float64(s.TotalLatency.Nanoseconds()) / float64((s.TotalRequests * 1000))
	upstreamAvgLatency := float64(s.UpstreamLatency.Nanoseconds()) / float64((s.TotalRequests * 1000))
	log.Infof("DNSProxy Server Statistics:\n" +
					   "\tTotal Requests: %d\n" +
					   "\tLatency: (Average)\n" +
					   "\t  Total: %fus\tProxy: %fus\tDNS Backend: %fus\n" +
					   "\tStart time: %s\n" +
					   "\tUp time: %s", s.TotalRequests, totalAvgLatency,
					   	totalAvgLatency-upstreamAvgLatency, upstreamAvgLatency,
					   	s.StartTime.String(), time.Since(s.StartTime).String())
}

// Initialize the config of the DNSProxy from json file
func (d *DNSProxy) InitConfig(conf_path string) {
	// Load the config from json file
	d.config = BuildConfig(conf_path)

	// Compile all the rules from the config
	err := d.rules.CompileRules(d.config.Rules)
	d.rules.SetScanAll(d.config.ScanAll)
	handleError(err, 52)
}

// Return the address and port of the server
func (d *DNSProxy) GetSocketAddress() string {
	return fmt.Sprintf("%s:%d",d.config.LocalAddress, d.config.LocalPort)
}

// Bind to port and start handle DNS requests
func (d *DNSProxy) ListenAndServe() error {
	// Set handlers
	mux := dns.NewServeMux()
	mux.HandleFunc("arpa.", d.handlePtr)


	if !d.config.StatisticsOn {
		log.Info("Statistics: enabled.")
		d.stats = Stats{0,0,
					   0, time.Now()}
		mux.HandleFunc(".", d.handleQueryStats)
	} else {
		log.Info("Statistics: disabled.")
		mux.HandleFunc(".", d.handleQuery)
	}

	// Build the DNS server
	d.server = &dns.Server{Addr: d.GetSocketAddress(), Net: "udp", Handler: mux}
	log.Infof("Starting server, listening on: %s", d.GetSocketAddress())

	// Start the DNS server
	return d.server.ListenAndServe()
}

// Internal function of passing requests to the upstream DNS server
func (d *DNSProxy) forwardRequest(req *dns.Msg) *dns.Msg {
	// Create a DNS client
	client := new(dns.Client)

	// Make a request to the upstream server
	resp, _, err := client.Exchange(req, fmt.Sprintf("%s:%d",d.config.RemoteHost, d.config.RemotePort))
	handleError(err, 92)
	return resp
}

// handle PTR records
// Currently PTR records rules are not supported
func (d *DNSProxy) handlePtr(resp dns.ResponseWriter, req *dns.Msg) {
	// Build new DNS message
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)

	// Send it to the upstream server
	reply := d.forwardRequest(upstreamMsg)

	// Build response and send it
	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Answer = reply.Answer
	err := resp.WriteMsg(respMsg)
	handleError(err, 111)
}

// handle Query requests that are not PTR
func (d *DNSProxy) handleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	// Copy the message for applying rules on it
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)
	upstreamMsg.Question = []dns.Question{}

	// Run on every query entry in the request
	for _, query := range req.Question {
		name := query.Name
		res, name := d.rules.Apply(name)

		// Check if the query has been blocked by white/black list rule
		if res.Code == BLOCKED {
			d.returnBlocked(resp, req)
			return
		}

		// Append new Question the the message
		rewrittenQuery := dns.Question{ Name: name, Qtype: query.Qtype, Qclass: query.Qclass}
		upstreamMsg.Question = append(upstreamMsg.Question, rewrittenQuery)
	}

	// Make a request to the upstream server
	reply := d.forwardRequest(upstreamMsg)

	// Set the original name in the response
	for index, q := range req.Question {
		reply.Answer[index].Header().Name = q.Name
	}

	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Answer = reply.Answer
	err := resp.WriteMsg(respMsg)
	handleError(err, 149)
}

// handle Query requests that are not PTR
func (d *DNSProxy) handleQueryStats(resp dns.ResponseWriter, req *dns.Msg) {
	StartProcessingTime := time.Now()
	// Copy the message for applying rules on it
	upstreamMsg := new(dns.Msg)
	req.CopyTo(upstreamMsg)
	upstreamMsg.Question = []dns.Question{}

	// Run on every query entry in the request
	for _, query := range req.Question {
		name := query.Name
		res, name := d.rules.Apply(name)

		// Check if the query has been blocked by white/black list rule
		if res.Code == BLOCKED {
			d.returnBlocked(resp, req)
			return
		}

		// Append new Question the the message
		rewrittenQuery := dns.Question{ Name: name, Qtype: query.Qtype, Qclass: query.Qclass}
		upstreamMsg.Question = append(upstreamMsg.Question, rewrittenQuery)
	}

	UpstreamStartTime := time.Now()
	// Make a request to the upstream server
	reply := d.forwardRequest(upstreamMsg)
	UpstreamDuration := time.Since(UpstreamStartTime)
	// Set the original name in the response
	for index, q := range req.Question {
		reply.Answer[index].Header().Name = q.Name
	}

	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Answer = reply.Answer
	err := resp.WriteMsg(respMsg)
	handleError(err, 189)

	d.stats.TotalLatency += time.Since(StartProcessingTime)
	d.stats.UpstreamLatency += UpstreamDuration
	d.stats.TotalRequests++
	if d.stats.TotalRequests % 50 == 0 {
		d.stats.logStats()
	}
}

// Build and send REFUSED response message to client
func (d *DNSProxy) returnBlocked(resp dns.ResponseWriter, req *dns.Msg) {
	respMsg := new(dns.Msg)
	respMsg.SetReply(req)
	respMsg.Rcode = dns.RcodeRefused
	err := resp.WriteMsg(respMsg)
	handleError(err, 205)
}