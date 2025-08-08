package dns01

import (
	"errors"
	"sort"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lookupNameserversOK(t *testing.T) {
	testCases := []struct {
		desc          string
		fakeDNSServer *dnsmock.Builder
		fqdn          string
		expected      []string
	}{
		{
			fqdn: "en.wikipedia.org.localhost.",
			fakeDNSServer: dnsmock.NewServer().
				Query("en.wikipedia.org.localhost SOA", dnsmock.CNAME("dyna.wikimedia.org.localhost")).
				Query("wikipedia.org.localhost SOA", dnsmock.SOA("")).
				Query("wikipedia.org.localhost NS",
					dnsmock.Answer(
						fakeNS("wikipedia.org.localhost.", "ns0.wikimedia.org.localhost."),
						fakeNS("wikipedia.org.localhost.", "ns1.wikimedia.org.localhost."),
						fakeNS("wikipedia.org.localhost.", "ns2.wikimedia.org.localhost."),
					),
				),
			expected: []string{"ns0.wikimedia.org.localhost.", "ns1.wikimedia.org.localhost.", "ns2.wikimedia.org.localhost."},
		},
		{
			fqdn: "www.google.com.localhost.",
			fakeDNSServer: dnsmock.NewServer().
				Query("www.google.com.localhost. SOA", dnsmock.Noop).
				Query("google.com.localhost. SOA", dnsmock.SOA("")).
				Query("google.com.localhost. NS",
					dnsmock.Answer(
						fakeNS("google.com.localhost.", "ns1.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns2.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns3.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns4.google.com.localhost."),
					),
				),
			expected: []string{"ns1.google.com.localhost.", "ns2.google.com.localhost.", "ns3.google.com.localhost.", "ns4.google.com.localhost."},
		},
		{
			fqdn: "mail.proton.me.localhost.",
			fakeDNSServer: dnsmock.NewServer().
				Query("mail.proton.me.localhost. SOA", dnsmock.Noop).
				Query("proton.me.localhost. SOA", dnsmock.SOA("")).
				Query("proton.me.localhost. NS",
					dnsmock.Answer(
						fakeNS("proton.me.localhost.", "ns1.proton.me.localhost."),
						fakeNS("proton.me.localhost.", "ns2.proton.me.localhost."),
						fakeNS("proton.me.localhost.", "ns3.proton.me.localhost."),
					),
				),
			expected: []string{"ns1.proton.me.localhost.", "ns2.proton.me.localhost.", "ns3.proton.me.localhost."},
		},
	}

	for _, test := range testCases {
		t.Run(test.fqdn, func(t *testing.T) {
			useAsNameserver(t, test.fakeDNSServer.Build(t))

			nss, err := lookupNameservers(test.fqdn)
			require.NoError(t, err)

			sort.Strings(nss)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, nss)
		})
	}
}

func Test_lookupNameserversErr(t *testing.T) {
	testCases := []struct {
		desc          string
		fqdn          string
		fakeDNSServer *dnsmock.Builder
		error         string
	}{
		{
			desc: "NXDOMAIN",
			fqdn: "example.invalid.",
			fakeDNSServer: dnsmock.NewServer().
				Query(". SOA", dnsmock.Error(dns.RcodeNameError)),
			error: "could not find zone: [fqdn=example.invalid.] could not find the start of authority for 'example.invalid.' [question='invalid. IN  SOA', code=NXDOMAIN]",
		},
		{
			desc: "NS error",
			fqdn: "example.com.",
			fakeDNSServer: dnsmock.NewServer().
				Query("example.com. SOA", dnsmock.SOA("")).
				Query("example.com. NS", dnsmock.Error(dns.RcodeServerFailure)),
			error: "[zone=example.com.] could not determine authoritative nameservers",
		},
		{
			desc: "empty NS",
			fqdn: "example.com.",
			fakeDNSServer: dnsmock.NewServer().
				Query("example.com. SOA", dnsmock.SOA("")).
				Query("example.me NS", dnsmock.Noop),
			error: "[zone=example.com.] could not determine authoritative nameservers",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			useAsNameserver(t, test.fakeDNSServer.Build(t))

			_, err := lookupNameservers(test.fqdn)
			require.Error(t, err)
			assert.EqualError(t, err, test.error)
		})
	}
}

type lookupSoaByFqdnTestCase struct {
	desc          string
	fqdn          string
	zone          string
	primaryNs     string
	nameservers   []string
	expectedError string
}

func lookupSoaByFqdnTestCases(t *testing.T) []lookupSoaByFqdnTestCase {
	t.Helper()

	return []lookupSoaByFqdnTestCase{
		{
			desc:      "domain is a CNAME",
			fqdn:      "mail.example.com.",
			zone:      "example.com.",
			primaryNs: "ns1.example.com.",
			nameservers: []string{
				dnsmock.NewServer().
					Query("mail.example.com. SOA", dnsmock.CNAME("example.com.")).
					Query("example.com. SOA", dnsmock.SOA("")).
					Build(t).
					String(),
			},
		},
		{
			desc:      "domain is a non-existent subdomain",
			fqdn:      "foo.example.com.",
			zone:      "example.com.",
			primaryNs: "ns1.example.com.",
			nameservers: []string{
				dnsmock.NewServer().
					Query("foo.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
					Query("example.com. SOA", dnsmock.SOA("")).
					Build(t).
					String(),
			},
		},
		{
			desc:      "domain is a eTLD",
			fqdn:      "example.com.ac.",
			zone:      "ac.",
			primaryNs: "ns1.nic.ac.",
			nameservers: []string{
				dnsmock.NewServer().
					Query("example.com.ac. SOA", dnsmock.Error(dns.RcodeNameError)).
					Query("com.ac. SOA", dnsmock.Error(dns.RcodeNameError)).
					Query("ac. SOA", dnsmock.SOA("")).
					Build(t).
					String(),
			},
		},
		{
			desc:      "domain is a cross-zone CNAME",
			fqdn:      "cross-zone-example.example.com.",
			zone:      "example.com.",
			primaryNs: "ns1.example.com.",
			nameservers: []string{
				dnsmock.NewServer().
					Query("cross-zone-example.example.com. SOA", dnsmock.CNAME("example.org.")).
					Query("example.com. SOA", dnsmock.SOA("")).
					Build(t).
					String(),
			},
		},
		{
			desc: "NXDOMAIN",
			fqdn: "test.lego.invalid.",
			zone: "lego.invalid.",
			nameservers: []string{
				dnsmock.NewServer().
					Query("test.lego.invalid. SOA", dnsmock.Error(dns.RcodeNameError)).
					Query("lego.invalid. SOA", dnsmock.Error(dns.RcodeNameError)).
					Query("invalid. SOA", dnsmock.Error(dns.RcodeNameError)).
					Build(t).
					String(),
			},
			expectedError: `[fqdn=test.lego.invalid.] could not find the start of authority for 'test.lego.invalid.' [question='invalid. IN  SOA', code=NXDOMAIN]`,
		},
		{
			desc:      "several non existent nameservers",
			fqdn:      "mail.example.com.",
			zone:      "example.com.",
			primaryNs: "ns1.example.com.",
			nameservers: []string{
				":7053",
				":8053",
				dnsmock.NewServer().
					Query("mail.example.com. SOA", dnsmock.CNAME("example.com.")).
					Query("example.com. SOA", dnsmock.SOA("")).
					Build(t).
					String(),
			},
		},
		{
			desc:        "only non-existent nameservers",
			fqdn:        "mail.example.com.",
			zone:        "example.com.",
			nameservers: []string{":7053", ":8053", ":9053"},
			// use only the start of the message because the port changes with each call: 127.0.0.1:XXXXX->127.0.0.1:7053.
			expectedError: "[fqdn=mail.example.com.] could not find the start of authority for 'mail.example.com.': DNS call error: read udp ",
		},
		{
			desc:          "no nameservers",
			fqdn:          "test.example.com.",
			zone:          "example.com.",
			nameservers:   []string{},
			expectedError: "[fqdn=test.example.com.] could not find the start of authority for 'test.example.com.': empty list of nameservers",
		},
	}
}

func TestFindZoneByFqdnCustom(t *testing.T) {
	for _, test := range lookupSoaByFqdnTestCases(t) {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			zone, err := FindZoneByFqdnCustom(test.fqdn, test.nameservers)
			if test.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.zone, zone)
			}
		})
	}
}

func TestFindPrimaryNsByFqdnCustom(t *testing.T) {
	for _, test := range lookupSoaByFqdnTestCases(t) {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			ns, err := FindPrimaryNsByFqdnCustom(test.fqdn, test.nameservers)
			if test.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.primaryNs, ns)
			}
		})
	}
}

func Test_getNameservers_ResolveConfServers(t *testing.T) {
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

func TestDNSError_Error(t *testing.T) {
	msgIn := createDNSMsg("example.com.", dns.TypeTXT, true)

	msgOut := createDNSMsg("example.org.", dns.TypeSOA, true)
	msgOut.Rcode = dns.RcodeNameError

	testCases := []struct {
		desc     string
		err      *DNSError
		expected string
	}{
		{
			desc:     "empty error",
			err:      &DNSError{},
			expected: "DNS error",
		},
		{
			desc: "all fields",
			err: &DNSError{
				Message: "Oops",
				NS:      "example.com.",
				MsgIn:   msgIn,
				MsgOut:  msgOut,
				Err:     errors.New("I did it again"),
			},
			expected: "Oops: I did it again [ns=example.com., question='example.com. IN  TXT', code=NXDOMAIN]",
		},
		{
			desc: "only NS",
			err: &DNSError{
				NS: "example.com.",
			},
			expected: "DNS error [ns=example.com.]",
		},
		{
			desc: "only MsgIn",
			err: &DNSError{
				MsgIn: msgIn,
			},
			expected: "DNS error [question='example.com. IN  TXT']",
		},
		{
			desc: "only MsgOut",
			err: &DNSError{
				MsgOut: msgOut,
			},
			expected: "DNS error [question='example.org. IN  SOA', code=NXDOMAIN]",
		},
		{
			desc: "only Err",
			err: &DNSError{
				Err: errors.New("I did it again"),
			},
			expected: "DNS error: I did it again",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.EqualError(t, test.err, test.expected)
		})
	}
}
