package dns01

import (
	"sort"
	"testing"

	"github.com/go-acme/lego/v5/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_checkNameserversPropagationCustom_authoritativeNss(t *testing.T) {
	testCases := []struct {
		desc          string
		fqdn, value   string
		fakeDNSServer *dnsmock.Builder
		expectedError string
	}{
		{
			desc: "TXT RR w/ expected value",
			// NS: asnums.routeviews.org.
			fqdn:  "8.8.8.8.asn.routeviews.org.",
			value: "151698.8.8.024",
			fakeDNSServer: dnsmock.NewServer().
				Query("8.8.8.8.asn.routeviews.org. TXT",
					dnsmock.Answer(
						fakeTXT("8.8.8.8.asn.routeviews.org.", "151698.8.8.024"),
					),
				),
		},
		{
			desc: "TXT RR w/ unexpected value",
			// NS: asnums.routeviews.org.
			fqdn:  "8.8.8.8.asn.routeviews.org.",
			value: "fe01=",
			fakeDNSServer: dnsmock.NewServer().
				Query("8.8.8.8.asn.routeviews.org. TXT",
					dnsmock.Answer(
						fakeTXT("8.8.8.8.asn.routeviews.org.", "15169"),
						fakeTXT("8.8.8.8.asn.routeviews.org.", "8.8.8.0"),
						fakeTXT("8.8.8.8.asn.routeviews.org.", "24"),
					),
				),
			expectedError: "did not return the expected TXT record [fqdn: 8.8.8.8.asn.routeviews.org., value: fe01=]: 15169 ,8.8.8.0 ,24",
		},
		{
			desc: "No TXT RR",
			// NS: ns2.google.com.
			fqdn:  "ns1.google.com.",
			value: "fe01=",
			fakeDNSServer: dnsmock.NewServer().
				Query("ns1.google.com.", dnsmock.Noop),
			expectedError: "did not return the expected TXT record [fqdn: ns1.google.com., value: fe01=]: ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(nil)

			addr := test.fakeDNSServer.Build(t)

			ok, err := client.checkNameserversPropagationCustom(t.Context(), test.fqdn, test.value, []string{addr.String()}, false)

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.True(t, ok)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, test.expectedError)
				assert.False(t, ok)
			}
		})
	}
}

func TestClient_lookupAuthoritativeNameservers_OK(t *testing.T) {
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
			client := NewClient(&Options{RecursiveNameservers: []string{test.fakeDNSServer.Build(t).String()}})

			nss, err := client.lookupAuthoritativeNameservers(t.Context(), test.fqdn)
			require.NoError(t, err)

			sort.Strings(nss)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, nss)
		})
	}
}

func TestClient_lookupAuthoritativeNameservers_error(t *testing.T) {
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
			client := NewClient(&Options{RecursiveNameservers: []string{test.fakeDNSServer.Build(t).String()}})

			_, err := client.lookupAuthoritativeNameservers(t.Context(), test.fqdn)
			require.Error(t, err)
			assert.EqualError(t, err, test.error)
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
