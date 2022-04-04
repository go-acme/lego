package localserver

import (
	"errors"
	"log"
	"net"
	"net/netip"
	"os"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "LOCALSERVER_"

	EnvListen = envNamespace + "LISTEN"
)

type DNSProvider struct {
	addr   string
	server *dns.Server
}

func NewDNSProvider() (*DNSProvider, error) {
	addr := os.Getenv(EnvListen)
	if addr == "" {
		return nil, errors.New("localserver: listen addr is nil")
	}
	_, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, err
	}
	return &DNSProvider{addr: addr}, nil
}

func (s *DNSProvider) Present(domain, token, keyAuth string) error {
	addr, err := netip.ParseAddrPort(s.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(addr))
	if err != nil {
		return err
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	handler := &dnsHandler{acmeResponse: value, fqdn: fqdn}
	s.server = &dns.Server{Handler: handler, PacketConn: conn}

	startCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	s.server.NotifyStartedFunc = func() {
		startCh <- struct{}{}
	}
	go func() {
		err := s.server.ActivateAndServe()
		if err != nil {
			errCh <- err
			return
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-startCh:
	}

	return nil
}

func (s *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	if s.server == nil {
		return errors.New("dns server is not running")
	}
	err := s.server.Shutdown()
	s.server = nil
	return err
}

type dnsHandler struct {
	acmeResponse string
	fqdn         string
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false
	question := r.Question[0]
	if question.Name == h.fqdn && question.Qtype == dns.TypeTXT {
		m.Answer = append(m.Answer, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
			},
			Txt: []string{h.acmeResponse},
		})
		log.Printf("localserver dns response: found, request: %#v\n", question)
	} else {
		m.Rcode = dns.RcodeNameError
		log.Printf("localserver dns response: not found, request: %#v\n", question)
	}
	if err := w.WriteMsg(m); err != nil {
		log.Println(err.Error())
	}
}
