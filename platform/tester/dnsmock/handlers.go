package dnsmock

import (
	"fmt"

	"github.com/miekg/dns"
)

func DumpRequest() dns.HandlerFunc {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		fmt.Println(req)

		Noop(w, req)
	}
}

func SOA(name string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		if name == "" {
			name = req.Question[0].Name
		}

		// Handle TLD
		base := name
		if dns.CountLabel(req.Question[0].Name) == 1 {
			base = "nic." + req.Question[0].Name
		}

		answer := &dns.SOA{
			Hdr:     dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 120},
			Ns:      "ns1." + base,
			Mbox:    "admin." + base,
			Serial:  2016022801,
			Refresh: 28800,
			Retry:   7200,
			Expire:  2419200,
			Minttl:  1200,
		}

		Answer(answer)(w, req)
	}
}

func CNAME(target string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		answer := &dns.CNAME{
			Hdr:    dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 1},
			Target: dns.Fqdn(target),
		}

		Answer(answer)(w, req)
	}
}

func Noop(w dns.ResponseWriter, req *dns.Msg) {
	_ = w.WriteMsg(new(dns.Msg).SetReply(req))
}

func Error(rcode int) dns.HandlerFunc {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		_ = w.WriteMsg(new(dns.Msg).SetRcode(req, rcode))
	}
}

func Answer(answer ...dns.RR) func(w dns.ResponseWriter, req *dns.Msg) {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg).SetReply(req)

		m.Answer = answer

		err := w.WriteMsg(m)
		if err != nil {
			panic(err.Error())
		}
	}
}
