package dns01

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"testing"

	getport "github.com/jsumners/go-getport"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	testDNSHandler struct{}
	testDNSServer  struct {
		*dns.Server
		getport.PortResult
	}
)

func (handler *testDNSHandler) ServeDNS(writer dns.ResponseWriter, reply *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(reply)

	if reply.Question[0].Qtype == dns.TypeA {
		msg.Authoritative = true
		domain := msg.Question[0].Name
		msg.Answer = append(
			msg.Answer,
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
			},
		)
	}

	_ = writer.WriteMsg(&msg)
}

// getTestNameserver constructs a new DNS server on a local address, or set
// of addresses, that responds to an `A` query for `example.com`.
func getTestNameserver(t *testing.T, network string) testDNSServer {
	t.Helper()
	server := &dns.Server{
		Handler: new(testDNSHandler),
		Net:     network,
	}
	testServer := testDNSServer{
		Server: server,
	}

	var protocol getport.Protocol
	var address string
	switch network {
	case "tcp":
		protocol = getport.TCP
		address = "0.0.0.0"
	case "tcp4":
		protocol = getport.TCP4
		address = "127.0.0.1"
	case "tcp6":
		protocol = getport.TCP6
		address = "::1"
	case "udp":
		protocol = getport.UDP
		address = "0.0.0.0"
	case "udp4":
		protocol = getport.UDP4
		address = "127.0.0.1"
	case "udp6":
		protocol = getport.UDP6
		address = "::1"
	}
	portResult, portError := getport.GetPort(protocol, address)
	if portError != nil {
		t.Error(portError)
		return testServer
	}
	testServer.PortResult = portResult
	server.Addr = getport.PortResultToAddress(portResult)

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	fin := make(chan error, 1)
	go func() {
		fin <- server.ListenAndServe()
	}()

	waitLock.Lock()
	return testServer
}

func TestSendDNSQuery(t *testing.T) {
	currentNameservers := recursiveNameservers

	t.Cleanup(func() {
		recursiveNameservers = currentNameservers
		SetNetworkStack(DefaultNetworkStack)
	})

	t.Run("does udp4 only", func(t *testing.T) {
		SetNetworkStack(IPv4Only)
		nameserver := getTestNameserver(t, getNetwork("udp"))
		defer func() { _ = nameserver.Server.Shutdown() }()

		serverAddress := fmt.Sprintf("127.0.0.1:%d", nameserver.PortResult.Port)
		recursiveNameservers = ParseNameservers([]string{serverAddress})
		msg := createDNSMsg("example.com.", dns.TypeA, true)
		result, queryError := sendDNSQuery(msg, serverAddress)
		assert.NoError(t, queryError)
		assert.Equal(t, result.Answer[0].(*dns.A).A.String(), "127.0.0.1")
	})

	t.Run("does udp6 only", func(t *testing.T) {
		SetNetworkStack(IPv6Only)
		nameserver := getTestNameserver(t, getNetwork("udp"))
		defer func() { _ = nameserver.Server.Shutdown() }()

		serverAddress := fmt.Sprintf("[::1]:%d", nameserver.PortResult.Port)
		recursiveNameservers = ParseNameservers([]string{serverAddress})
		msg := createDNSMsg("example.com.", dns.TypeA, true)
		result, queryError := sendDNSQuery(msg, serverAddress)
		assert.NoError(t, queryError)
		assert.Equal(t, result.Answer[0].(*dns.A).A.String(), "127.0.0.1")
	})

	t.Run("does tcp4 and tcp6", func(t *testing.T) {
		SetNetworkStack(DefaultNetworkStack)
		nameserver := getTestNameserver(t, getNetwork("tcp"))
		defer func() { _ = nameserver.Server.Shutdown() }()

		serverAddress := fmt.Sprintf("[::1]:%d", nameserver.PortResult.Port)
		recursiveNameservers = ParseNameservers([]string{serverAddress})
		msg := createDNSMsg("example.com.", dns.TypeA, true)
		result, queryError := sendDNSQuery(msg, serverAddress)
		assert.NoError(t, queryError)
		assert.Equal(t, result.Answer[0].(*dns.A).A.String(), "127.0.0.1")

		serverAddress = fmt.Sprintf("127.0.0.1:%d", nameserver.PortResult.Port)
		recursiveNameservers = ParseNameservers([]string{serverAddress})
		msg = createDNSMsg("example.com.", dns.TypeA, true)
		result, queryError = sendDNSQuery(msg, serverAddress)
		assert.NoError(t, queryError)
		assert.Equal(t, result.Answer[0].(*dns.A).A.String(), "127.0.0.1")
	})
}

func TestLookupNameserversOK(t *testing.T) {
	testCases := []struct {
		fqdn string
		nss  []string
	}{
		{
			fqdn: "books.google.com.ng.",
			nss:  []string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
		},
		{
			fqdn: "www.google.com.",
			nss:  []string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
		},
		{
			fqdn: "physics.georgetown.edu.",
			nss:  []string{"ns4.georgetown.edu.", "ns5.georgetown.edu.", "ns6.georgetown.edu."},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.fqdn, func(t *testing.T) {
			t.Parallel()

			nss, err := lookupNameservers(test.fqdn)
			require.NoError(t, err)

			sort.Strings(nss)
			sort.Strings(test.nss)

			assert.EqualValues(t, test.nss, nss)
		})
	}
}

func TestLookupNameserversErr(t *testing.T) {
	testCases := []struct {
		desc  string
		fqdn  string
		error string
	}{
		{
			desc:  "invalid tld",
			fqdn:  "_null.n0n0.",
			error: "could not determine the zone",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := lookupNameservers(test.fqdn)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.error)
		})
	}
}

var findXByFqdnTestCases = []struct {
	desc          string
	fqdn          string
	zone          string
	primaryNs     string
	nameservers   []string
	expectedError string
}{
	{
		desc:        "domain is a CNAME",
		fqdn:        "mail.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a non-existent subdomain",
		fqdn:        "foo.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a eTLD",
		fqdn:        "example.com.ac.",
		zone:        "ac.",
		primaryNs:   "a0.nic.ac.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a cross-zone CNAME",
		fqdn:        "cross-zone-example.assets.sh.",
		zone:        "assets.sh.",
		primaryNs:   "gina.ns.cloudflare.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:          "NXDOMAIN",
		fqdn:          "test.lego.zz.",
		zone:          "lego.zz.",
		nameservers:   []string{"8.8.8.8:53"},
		expectedError: "could not find the start of authority for test.lego.zz.: NXDOMAIN",
	},
	{
		desc:        "several non existent nameservers",
		fqdn:        "mail.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: []string{":7053", ":8053", "8.8.8.8:53"},
	},
	{
		desc:          "only non-existent nameservers",
		fqdn:          "mail.google.com.",
		zone:          "google.com.",
		nameservers:   []string{":7053", ":8053", ":9053"},
		expectedError: "could not find the start of authority for mail.google.com.: dial tcp :9053: connect:",
	},
	{
		desc:          "no nameservers",
		fqdn:          "test.ldez.com.",
		zone:          "ldez.com.",
		nameservers:   []string{},
		expectedError: "could not find the start of authority for test.ldez.com.",
	},
}

func TestFindZoneByFqdnCustom(t *testing.T) {
	for _, test := range findXByFqdnTestCases {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			zone, err := FindZoneByFqdnCustom(test.fqdn, test.nameservers)
			if test.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.zone, zone)
			}
		})
	}
}

func TestFindPrimaryNsByFqdnCustom(t *testing.T) {
	for _, test := range findXByFqdnTestCases {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			ns, err := FindPrimaryNsByFqdnCustom(test.fqdn, test.nameservers)
			if test.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.primaryNs, ns)
			}
		})
	}
}

func TestResolveConfServers(t *testing.T) {
	testCases := []struct {
		fixture  string
		expected []string
		defaults []string
	}{
		{
			fixture:  "fixtures/resolv.conf.1",
			defaults: []string{"127.0.0.1:53"},
			expected: []string{"10.200.3.249:53", "10.200.3.250:5353", "[2001:4860:4860::8844]:53", "[10.0.0.1]:5353"},
		},
		{
			fixture:  "fixtures/resolv.conf.nonexistant",
			defaults: []string{"127.0.0.1:53"},
			expected: []string{"127.0.0.1:53"},
		},
	}

	for _, test := range testCases {
		t.Run(test.fixture, func(t *testing.T) {
			result := getNameservers(test.fixture, test.defaults)

			sort.Strings(result)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, result)
		})
	}
}
