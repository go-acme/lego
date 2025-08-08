package rfc2136

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

const (
	fakeDomain     = "123456789.www.example.com"
	fakeKeyAuth    = "123d=="
	fakeValue      = "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"
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

func TestDNSProvider_Present_success(t *testing.T) {
	dns01.ClearFqdnCache()

	addr := dnsmock.NewServer().
		Query(fakeZone+" SOA", dnsmock.SOA("")).
		Update(fakeZone+" SOA", dnsmock.Noop).
		Build(t)

	config := NewDefaultConfig()
	config.Nameserver = addr.String()

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.NoError(t, err)
}

func TestDNSProvider_Present_success_updatePacket(t *testing.T) {
	dns01.ClearFqdnCache()

	reqChan := make(chan *dns.Msg, 1)

	addr := dnsmock.NewServer().
		Query("_acme-challenge.123456789.www.example.com. SOA", dnsmock.SOA(fakeZone)).
		Update(fakeZone+" SOA", func(w dns.ResponseWriter, req *dns.Msg) {
			dnsmock.Noop(w, req)

			// Only talk back when it is not the SOA RR.
			reqChan <- req
		}).
		Build(t)

	config := NewDefaultConfig()
	config.Nameserver = addr.String()

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.NoError(t, err)

	select {
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for request")

	case rcvMsg := <-reqChan:
		txtRR := &dns.TXT{
			Hdr: dns.RR_Header{Name: fakeFqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: fakeTTL},
			Txt: []string{fakeValue},
		}

		m := new(dns.Msg).SetUpdate(fakeZone)

		m.RemoveRRset([]dns.RR{txtRR})
		m.Insert([]dns.RR{txtRR})

		expected, err := m.Pack()
		require.NoError(t, err, "error packing")

		rcvMsg.Id = m.Id

		actual, err := rcvMsg.Pack()
		require.NoError(t, err, "error packing")

		if !bytes.Equal(actual, expected) {
			tmp := new(dns.Msg)
			require.NoError(t, tmp.Unpack(actual))

			t.Errorf("Expected msg:\n%s", m)
			t.Errorf("Actual msg:\n%s", tmp)
		}
	}
}

func TestDNSProvider_Present_error(t *testing.T) {
	dns01.ClearFqdnCache()

	addr := dnsmock.NewServer().
		Query(fakeZone+" SOA", dnsmock.Error(dns.RcodeNotZone)).
		Build(t)

	config := NewDefaultConfig()
	config.Nameserver = addr.String()

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.Error(t, err)
	if !strings.Contains(err.Error(), "NOTZONE") {
		t.Errorf("Expected Present() to return an error with the 'NOTZONE' rcode string, but it did not: %v", err)
	}
}

func TestDNSProvider_Present_tsig_success(t *testing.T) {
	dns01.ClearFqdnCache()

	addr := dnsmock.NewServer().
		Query(fakeZone+" SOA", dnsmock.SOA("")).
		Update(fakeZone+" SOA", handleTSIG).
		Build(t, func(server *dns.Server) error {
			server.TsigSecret = map[string]string{fakeTsigKey: fakeTsigSecret}

			return nil
		})

	config := NewDefaultConfig()
	config.Nameserver = addr.String()
	config.TSIGKey = fakeTsigKey
	config.TSIGSecret = fakeTsigSecret

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.NoError(t, err)
}

func TestDNSProvider_Present_tsig_error(t *testing.T) {
	dns01.ClearFqdnCache()

	addr := dnsmock.NewServer().
		Query(fakeZone+" SOA", dnsmock.SOA("")).
		Update(fakeZone+" SOA", handleTSIG).
		Build(t, func(server *dns.Server) error {
			server.TsigSecret = map[string]string{"example.org": fakeTsigSecret}

			return nil
		})

	config := NewDefaultConfig()
	config.Nameserver = addr.String()
	config.TSIGKey = fakeTsigKey
	config.TSIGSecret = fakeTsigSecret

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(fakeDomain, "", fakeKeyAuth)
	require.Error(t, err)
	require.EqualError(t, err, "rfc2136: failed to insert: DNS update failed: server replied: NOTZONE")
}

func handleTSIG(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)

	tsig := req.IsTsig()
	if tsig == nil {
		_ = w.WriteMsg(m.SetRcode(req, dns.RcodeRefused))
		return
	}

	err := w.TsigStatus()
	if err != nil {
		_ = w.WriteMsg(m.SetRcode(req, dns.RcodeNotZone))

		return
	}

	// Validated
	_ = w.WriteMsg(m.
		SetReply(req).
		SetTsig(tsig.Hdr.Name, tsig.Algorithm, tsig.Fudge, time.Now().Unix()),
	)
}
