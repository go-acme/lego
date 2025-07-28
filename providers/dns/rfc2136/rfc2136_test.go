package rfc2136

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fakeDomain     = "123456789.www.example.com"
	fakeKeyAuth    = "123d=="
	fakeValue      = "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo"
	fakeFqdn       = "_acme-challenge.123456789.www.example.com."
	fakeZone       = "example.com."
	fakeTTL        = 120
	fakeTsigKey    = "example.com."
	fakeTsigSecret = "IwBTJx9wrDp4Y1RyC3H0gA=="
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvTSIGFile,
	EnvTSIGKey,
	EnvTSIGSecret,
	EnvTSIGAlgorithm,
	EnvNameserver,
	EnvDNSTimeout,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvNameserver: "example.com",
			},
		},
		{
			desc: "missing nameserver",
			envVars: map[string]string{
				EnvNameserver: "",
			},
			expected: "rfc2136: some credentials information are missing: RFC2136_NAMESERVER",
		},
		{
			desc: "invalid algorithm",
			envVars: map[string]string{
				EnvNameserver:    "example.com",
				EnvTSIGKey:       "",
				EnvTSIGSecret:    "",
				EnvTSIGAlgorithm: "foo",
			},
			expected: "rfc2136: unsupported TSIG algorithm: foo.",
		},
		{
			desc: "valid TSIG file",
			envVars: map[string]string{
				EnvNameserver: "example.com",
				EnvTSIGFile:   "./internal/fixtures/sample.conf",
			},
		},
		{
			desc: "invalid TSIG file",
			envVars: map[string]string{
				EnvNameserver: "example.com",
				EnvTSIGFile:   "./internal/fixtures/invalid_key.conf",
			},
			expected: "rfc2136: read TSIG file ./internal/fixtures/invalid_key.conf: invalid key line: key {",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		expected      string
		nameserver    string
		tsigFile      string
		tsigAlgorithm string
		tsigKey       string
		tsigSecret    string
	}{
		{
			desc:       "success",
			nameserver: "example.com",
		},
		{
			desc:     "missing nameserver",
			expected: "rfc2136: nameserver missing",
		},
		{
			desc:          "invalid algorithm",
			nameserver:    "example.com",
			tsigAlgorithm: "foo",
			expected:      "rfc2136: unsupported TSIG algorithm: foo.",
		},
		{
			desc:       "valid TSIG file",
			nameserver: "example.com",
			tsigFile:   "./internal/fixtures/sample.conf",
		},
		{
			desc:       "invalid TSIG file",
			nameserver: "example.com",
			tsigFile:   "./internal/fixtures/invalid_key.conf",
			expected:   "rfc2136: read TSIG file ./internal/fixtures/invalid_key.conf: invalid key line: key {",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Nameserver = test.nameserver
			config.TSIGFile = test.tsigFile
			config.TSIGAlgorithm = test.tsigAlgorithm
			config.TSIGKey = test.tsigKey
			config.TSIGSecret = test.tsigSecret

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestCanaryLocalTestServer(t *testing.T) {
	dns01.ClearFqdnCache()

	mux, addr := runLocalDNSTestServer(t, false)
	mux.HandleFunc("example.com.", serverHandlerHello)

	c := new(dns.Client)
	m := new(dns.Msg)

	m.SetQuestion("example.com.", dns.TypeTXT)

	r, _, err := c.Exchange(m, addr)
	require.NoError(t, err, "Failed to communicate with test server")
	assert.Len(t, r.Extra, 1, "Failed to communicate with test server")

	txt := r.Extra[0].(*dns.TXT).Txt[0]
	assert.Equal(t, "Hello world", txt)
}

func TestServerSuccess(t *testing.T) {
	dns01.ClearFqdnCache()

	mux, addr := runLocalDNSTestServer(t, false)
	mux.HandleFunc(fakeZone, serverHandlerReturnSuccess)

	config := NewDefaultConfig()
	config.Nameserver = addr

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.NoError(t, err)
}

func TestServerError(t *testing.T) {
	dns01.ClearFqdnCache()

	mux, addr := runLocalDNSTestServer(t, false)
	mux.HandleFunc(fakeZone, serverHandlerReturnErr)

	config := NewDefaultConfig()
	config.Nameserver = addr

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.Error(t, err)
	if !strings.Contains(err.Error(), "NOTZONE") {
		t.Errorf("Expected Present() to return an error with the 'NOTZONE' rcode string, but it did not: %v", err)
	}
}

func TestTsigClient(t *testing.T) {
	dns01.ClearFqdnCache()

	mux, addr := runLocalDNSTestServer(t, true)
	mux.HandleFunc(fakeZone, serverHandlerReturnSuccess)

	config := NewDefaultConfig()
	config.Nameserver = addr
	config.TSIGKey = fakeTsigKey
	config.TSIGSecret = fakeTsigSecret

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.NoError(t, err)
}

func TestValidUpdatePacket(t *testing.T) {
	reqChan := make(chan *dns.Msg, 10)

	dns01.ClearFqdnCache()

	mux, addr := runLocalDNSTestServer(t, false)
	mux.HandleFunc(fakeZone, serverHandlerPassBackRequest(reqChan))

	txtRR, _ := dns.NewRR(fmt.Sprintf("%s %d IN TXT %s", fakeFqdn, fakeTTL, fakeValue))

	m := new(dns.Msg)
	m.SetUpdate(fakeZone)
	m.RemoveRRset([]dns.RR{txtRR})
	m.Insert([]dns.RR{txtRR})

	expectStr := m.String()

	expect, err := m.Pack()
	require.NoError(t, err, "error packing")

	config := NewDefaultConfig()
	config.Nameserver = addr

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", "1234d==")
	require.NoError(t, err)

	rcvMsg := <-reqChan
	rcvMsg.Id = m.Id

	actual, err := rcvMsg.Pack()
	require.NoError(t, err, "error packing")

	if !bytes.Equal(actual, expect) {
		tmp := new(dns.Msg)
		if err := tmp.Unpack(actual); err != nil {
			t.Fatalf("Error unpacking actual msg: %v", err)
		}
		t.Errorf("Expected msg:\n%s", expectStr)
		t.Errorf("Actual msg:\n%v", tmp)
	}
}

func runLocalDNSTestServer(t *testing.T, tsig bool) (*dns.ServeMux, string) {
	t.Helper()

	mux := dns.NewServeMux()

	server := &dns.Server{
		Addr:         "127.0.0.1:0",
		Net:          "udp",
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
		MsgAcceptFunc: func(dh dns.Header) dns.MsgAcceptAction {
			// bypass defaultMsgAcceptFunc to allow dynamic update (https://github.com/miekg/dns/pull/830)
			return dns.MsgAccept
		},
		Handler: mux,
	}

	t.Cleanup(func() {
		_ = server.Shutdown()
	})

	if tsig {
		server.TsigSecret = map[string]string{fakeTsigKey: fakeTsigSecret}
	}

	waitLock := sync.Mutex{}
	waitLock.Lock()

	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			t.Log(err)
		}
	}()

	waitLock.Lock()

	return mux, server.PacketConn.LocalAddr().String()
}

func serverHandlerHello(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	m.Extra = make([]dns.RR, 1)
	m.Extra[0] = &dns.TXT{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{"Hello world"},
	}

	_ = w.WriteMsg(m)
}

func serverHandlerReturnSuccess(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	if req.Opcode == dns.OpcodeQuery && req.Question[0].Qtype == dns.TypeSOA && req.Question[0].Qclass == dns.ClassINET {
		// Return SOA to appease findZoneByFqdn()
		m.Answer = []dns.RR{fakeSOAAnswer()}
	}

	if t := req.IsTsig(); t != nil {
		if w.TsigStatus() == nil {
			// Validated
			m.SetTsig(fakeZone, dns.HmacSHA1, 300, time.Now().Unix())
		}
	}

	_ = w.WriteMsg(m)
}

func serverHandlerReturnErr(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeNotZone)

	_ = w.WriteMsg(m)
}

func serverHandlerPassBackRequest(reqChan chan *dns.Msg) func(w dns.ResponseWriter, req *dns.Msg) {
	return func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)

		if req.Opcode == dns.OpcodeQuery && req.Question[0].Qtype == dns.TypeSOA && req.Question[0].Qclass == dns.ClassINET {
			// Return SOA to appease findZoneByFqdn()
			m.Answer = []dns.RR{fakeSOAAnswer()}
		}

		if t := req.IsTsig(); t != nil {
			if w.TsigStatus() == nil {
				// Validated
				m.SetTsig(fakeZone, dns.HmacSHA1, 300, time.Now().Unix())
			}
		}

		_ = w.WriteMsg(m)

		if req.Opcode != dns.OpcodeQuery || req.Question[0].Qtype != dns.TypeSOA || req.Question[0].Qclass != dns.ClassINET {
			// Only talk back when it is not the SOA RR.
			reqChan <- req
		}
	}
}

func fakeSOAAnswer() *dns.SOA {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: fakeZone, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: fakeTTL},
		Ns:      "ns1." + fakeZone,
		Mbox:    "admin." + fakeZone,
		Serial:  2016022801,
		Refresh: 28800,
		Retry:   7200,
		Expire:  2419200,
		Minttl:  1200,
	}
}
