package dnsproxy

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/golang/glog"
)

func handleError(e error) {
	if e != nil {
		glog.Error(e)
	}
}

type DNSProxy struct {
	config Config
	server *dns.Server
	rules []ProxyRule
}

func (d *DNSProxy) InitConfig(conf_path string) {
	d.config = BuildConfig(conf_path)
	d.rules = CompileRules(d.config.Rules)
}

func (d *DNSProxy) GetSocketAddress() string {
	return fmt.Sprintf("%s:%d",d.config.Local_addr, d.config.Local_port)
}

func (d *DNSProxy) ListenAndServe() error {
	mux := dns.NewServeMux()
	mux.HandleFunc("arpa.", d.handlePtr)
	mux.HandleFunc(".", d.handleQuery)

	d.server = &dns.Server{Addr: d.GetSocketAddress(), Net: "udp", Handler: mux}
	glog.Infof("Starting server, listening on: %s", d.GetSocketAddress())

	return d.server.ListenAndServe()
}

func (d *DNSProxy) forwardRequest(req *dns.Msg) *dns.Msg {
	dns_srv := fmt.Sprintf("%s:%d",d.config.Remote_host, d.config.Remote_port)
	client := new(dns.Client)
	resp, _, err := client.Exchange(req, dns_srv)
	handleError(err)
	return resp
}

func (d *DNSProxy) handlePtr(resp dns.ResponseWriter, req *dns.Msg) {
	upstream_msg := new(dns.Msg)
	req.CopyTo(upstream_msg)

	reply := d.forwardRequest(upstream_msg)

	resp_msg := new(dns.Msg)
	resp_msg.SetReply(req)
	resp_msg.Answer = reply.Answer
	err := resp.WriteMsg(resp_msg)
	handleError(err)
}

func (d *DNSProxy) handleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	upstream_msg := new(dns.Msg)
	req.CopyTo(upstream_msg)
	upstream_msg.Question = []dns.Question{}

	for _, query := range req.Question {
		name := query.Name
		for _, rule := range d.rules {
			if rule.CheckPattern(query.Name) {
				name = rule.Apply(query.Name)
				break
			}
		}

		rewrited_query := dns.Question{ name, query.Qtype, query.Qclass}
		upstream_msg.Question = append(upstream_msg.Question, rewrited_query)
	}

	reply := d.forwardRequest(upstream_msg)
	reply.Answer[0].Header().Name = req.Question[0].Name

	resp_msg := new(dns.Msg)
	resp_msg.SetReply(req)
	resp_msg.Answer = reply.Answer
	err := resp.WriteMsg(resp_msg)
	handleError(err)
}
