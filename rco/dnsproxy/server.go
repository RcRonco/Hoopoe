package dnsproxy

import (
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"fmt"
)

// Handle errors that will not cause crash of the system.
func handleError(e error) {
	if e != nil {
		glog.Error(e)
	}
}

// Proxy server implementation
type DNSProxy struct {
	config Config
	server *dns.Server
	rules RuleEngine
}


// Initialize the config of the DNSProxy from json file
func (d *DNSProxy) InitConfig(conf_path string) {
	// Load the config from json file
	d.config = BuildConfig(conf_path)

	// Compile all the rules from the config
	err := d.rules.CompileRules(d.config.Rules)
	handleError(err)
}

// Return the address and port of the server
func (d *DNSProxy) GetSocketAddress() string {
	return fmt.Sprintf("%s:%d",d.config.Local_addr, d.config.Local_port)
}

// Bind to port and start handle DNS requests
func (d *DNSProxy) ListenAndServe() error {
	// Set handlers
	mux := dns.NewServeMux()
	mux.HandleFunc("arpa.", d.handlePtr)
	mux.HandleFunc(".", d.handleQuery)

	// Build the DNS server
	d.server = &dns.Server{Addr: d.GetSocketAddress(), Net: "udp", Handler: mux}
	glog.Infof("Starting server, listening on: %s", d.GetSocketAddress())

	// Start the DNS server
	return d.server.ListenAndServe()
}

// Internal function of passing requests to the upstream DNS server
func (d *DNSProxy) forwardRequest(req *dns.Msg) *dns.Msg {
	// Create a DNS client
	client := new(dns.Client)

	// Make a request to the upstream server
	resp, _, err := client.Exchange(req, fmt.Sprintf("%s:%d",d.config.Remote_host, d.config.Remote_port))
	handleError(err)
	return resp
}

// handle PTR records
// Currently PTR records rules are not supported
func (d *DNSProxy) handlePtr(resp dns.ResponseWriter, req *dns.Msg) {
	// Build new DNS message
	upstream_msg := new(dns.Msg)
	req.CopyTo(upstream_msg)

	// Send it to the upstream server
	reply := d.forwardRequest(upstream_msg)

	// Build response and send it
	resp_msg := new(dns.Msg)
	resp_msg.SetReply(req)
	resp_msg.Answer = reply.Answer
	err := resp.WriteMsg(resp_msg)
	handleError(err)
}

// handle Query requests that are not PTR
func (d *DNSProxy) handleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	// Copy the message for applying rules on it
	upstream_msg := new(dns.Msg)
	req.CopyTo(upstream_msg)
	upstream_msg.Question = []dns.Question{}

	// Run on every query entry in the request
	for _, query := range req.Question {
		name := query.Name
		res, name := d.rules.Apply(name)

		// Check if the query has been blocked by white/black list rule
		if (res.Code == BLOCKED) {
			d.returnBlocked(resp, req)
			return
		}

		// Append new Question the the message
		rewrited_query := dns.Question{ name, query.Qtype, query.Qclass}
		upstream_msg.Question = append(upstream_msg.Question, rewrited_query)
	}

	// Make a request to the upstream server
	reply := d.forwardRequest(upstream_msg)

	// Set the original name in the response
	for index, q := range req.Question {
		reply.Answer[index].Header().Name = q.Name
	}

	resp_msg := new(dns.Msg)
	resp_msg.SetReply(req)
	resp_msg.Answer = reply.Answer
	err := resp.WriteMsg(resp_msg)
	handleError(err)
}

// Build and send REFUSED response message to client
func (d *DNSProxy) returnBlocked(resp dns.ResponseWriter, req *dns.Msg) {
	resp_msg := new(dns.Msg)
	resp_msg.SetReply(req)
	resp_msg.Rcode = dns.RcodeRefused
	err := resp.WriteMsg(resp_msg)
	handleError(err)
}