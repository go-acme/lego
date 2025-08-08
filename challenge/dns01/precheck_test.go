package dns01

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_preCheck_checkDNSPropagation(t *testing.T) {
	mockResolver(t,
		dnsmock.NewServer().
			Query("ns0.lego.localhost. A",
				dnsmock.Answer(fakeA("ns0.lego.localhost.", "127.0.0.1"))).
			Query("ns1.lego.localhost. A",
				dnsmock.Answer(fakeA("ns1.lego.localhost.", "127.0.0.1"))).
			Query("example.com. TXT",
				dnsmock.Answer(
					fakeTXT("example.com.", "one"),
					fakeTXT("example.com.", "two"),
					fakeTXT("example.com.", "three"),
					fakeTXT("example.com.", "four"),
					fakeTXT("example.com.", "five"),
				),
			).
			Build(t),
	)

	useAsNameserver(t,
		dnsmock.NewServer().
			Query("acme-staging.api.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
			Query("api.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
			Query("example.com. SOA", dnsmock.SOA("")).
			Query("example.com. NS",
				dnsmock.Answer(
					fakeNS("example.com.", "ns0.lego.localhost."),
					fakeNS("example.com.", "ns1.lego.localhost."),
				),
			).
			Build(t),
	)

	testCases := []struct {
		desc          string
		fqdn          string
		value         string
		expectedError string
	}{
		{
			desc:  "success",
			fqdn:  "example.com.",
			value: "four",
		},
		{
			desc:          "no matching TXT record",
			fqdn:          "acme-staging.api.example.com.",
			value:         "fe01=",
			expectedError: "did not return the expected TXT record [fqdn: acme-staging.api.example.com., value: fe01=]: one ,two ,three ,four ,five",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			check := newPreCheck()

			ok, err := check.checkDNSPropagation(test.fqdn, test.value)
			if test.expectedError != "" {
				assert.ErrorContainsf(t, err, test.expectedError, "PreCheckDNS must fail for %s", test.fqdn)
				assert.False(t, ok, "PreCheckDNS must fail for %s", test.fqdn)
			} else {
				assert.NoErrorf(t, err, "PreCheckDNS failed for %s", test.fqdn)
				assert.True(t, ok, "PreCheckDNS failed for %s", test.fqdn)
			}
		})
	}
}

func Test_checkNameserversPropagation_authoritativeNss(t *testing.T) {
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
			ClearFqdnCache()

			addr := test.fakeDNSServer.Build(t)

			ok, err := checkNameserversPropagation(test.fqdn, test.value, []string{addr.String()}, false)

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
