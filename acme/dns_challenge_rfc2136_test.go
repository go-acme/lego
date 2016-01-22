package acme

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

var (
	rfc2136TestDomain     = "123456789.www.example.com"
	rfc2136TestKeyAuth    = "123d=="
	rfc2136TestValue      = "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo"
	rfc2136TestFqdn       = "_acme-challenge.123456789.www.example.com."
	rfc2136TestZone       = "example.com."
	rfc2136TestTTL        = 120
	rfc2136TestTsigKey    = "example.com."
	rfc2136TestTsigSecret = "IwBTJx9wrDp4Y1RyC3H0gA=="
)

var reqChan = make(chan *dns.Msg, 10)

func TestRFC2136CanaryLocalTestServer(t *testing.T) {
	dns.HandleFunc("example.com.", serverHandlerHello)
	defer dns.HandleRemove("example.com.")

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeTXT)
	r, _, err := c.Exchange(m, addrstr)
	if err != nil || len(r.Extra) == 0 {
		t.Fatalf("Failed to communicate with test server: %v", err)
	}
	txt := r.Extra[0].(*dns.TXT).Txt[0]
	if txt != "Hello world" {
		t.Error("Expected test server to return 'Hello world' but got: ", txt)
	}
}

func TestRFC2136ServerSuccess(t *testing.T) {
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnSuccess)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderRFC2136(addrstr, rfc2136TestZone, "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderRFC2136() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}
}

func TestRFC2136ServerError(t *testing.T) {
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnErr)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderRFC2136(addrstr, rfc2136TestZone, "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderRFC2136() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err == nil {
		t.Errorf("Expected Present() to return an error but it did not.")
	} else if !strings.Contains(err.Error(), "NOTZONE") {
		t.Errorf("Expected Present() to return an error with the 'NOTZONE' rcode string but it did not.")
	}
}

func TestRFC2136TsigClient(t *testing.T) {
	dns.HandleFunc(rfc2136TestZone, serverHandlerReturnSuccess)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", true)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	provider, err := NewDNSProviderRFC2136(addrstr, rfc2136TestZone, rfc2136TestTsigKey, rfc2136TestTsigSecret)
	if err != nil {
		t.Fatalf("Expected NewDNSProviderRFC2136() to return no error but the error was -> %v", err)
	}
	if err := provider.Present(rfc2136TestDomain, "", rfc2136TestKeyAuth); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}
}

func TestRFC2136ValidUpdatePacket(t *testing.T) {
	dns.HandleFunc(rfc2136TestZone, serverHandlerPassBackRequest)
	defer dns.HandleRemove(rfc2136TestZone)

	server, addrstr, err := runLocalDNSTestServer("127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer server.Shutdown()

	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{
		Name:   rfc2136TestFqdn,
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassINET,
		Ttl:    uint32(rfc2136TestTTL),
	}
	rr.Txt = []string{rfc2136TestValue}
	rrs := make([]dns.RR, 1)
	rrs[0] = rr
	m := new(dns.Msg)
	m.SetUpdate(dns.Fqdn(rfc2136TestZone))
	m.Insert(rrs)
	expectstr := m.String()
	expect, err := m.Pack()
	if err != nil {
		t.Fatalf("Error packing expect msg: %v", err)
	}

	provider, err := NewDNSProviderRFC2136(addrstr, rfc2136TestZone, "", "")
	if err != nil {
		t.Fatalf("Expected NewDNSProviderRFC2136() to return no error but the error was -> %v", err)
	}

	if err := provider.Present(rfc2136TestDomain, "", "1234d=="); err != nil {
		t.Errorf("Expected Present() to return no error but the error was -> %v", err)
	}

	rcvMsg := <-reqChan
	rcvMsg.Id = m.Id
	actual, err := rcvMsg.Pack()
	if err != nil {
		t.Fatalf("Error packing actual msg: %v", err)
	}

	if !bytes.Equal(actual, expect) {
		tmp := new(dns.Msg)
		if err := tmp.Unpack(actual); err != nil {
			t.Fatalf("Error unpacking actual msg: %v", err)
		}
		t.Errorf("Expected msg:\n%s", expectstr)
		t.Errorf("Actual msg:\n%v", tmp)
	}
}

func runLocalDNSTestServer(listenAddr string, tsig bool) (*dns.Server, string, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, "", err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}
	if tsig {
		server.TsigSecret = map[string]string{rfc2136TestTsigKey: rfc2136TestTsigSecret}
	}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		server.ActivateAndServe()
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), nil
}

func serverHandlerHello(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Extra = make([]dns.RR, 1)
	m.Extra[0] = &dns.TXT{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{"Hello world"},
	}
	w.WriteMsg(m)
}

func serverHandlerReturnSuccess(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	if t := req.IsTsig(); t != nil {
		if w.TsigStatus() == nil {
			// Validated
			m.SetTsig(rfc2136TestZone, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
}

func serverHandlerReturnErr(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeNotZone)
	w.WriteMsg(m)
}

func serverHandlerPassBackRequest(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	if t := req.IsTsig(); t != nil {
		if w.TsigStatus() == nil {
			// Validated
			m.SetTsig(rfc2136TestZone, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
	reqChan <- req
}
