package dns01

import (
	"testing"

	dnsmock2 "github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_preCheck_checkDNSPropagation(t *testing.T) {
	mockDefault(t,
		dnsmock2.NewServer().
			// This line is here to produce an error if the calls don't go on the right DNS server.
			Query("acme-staging.api.example.com. SOA", dnsmock2.Error(dns.RcodeNameError)).
			// This line is here to produce an error if the calls don't go on the right DNS server.
			Query("api.example.com. SOA", dnsmock2.Error(dns.RcodeNameError)).
			Query("example.com. SOA", dnsmock2.SOA("")).
			Query("example.com. NS",
				dnsmock2.Answer(
					fakeNS("example.com.", "ns0.lego.localhost."),
					fakeNS("example.com.", "ns1.lego.localhost."),
				),
			).
			Build(t),
		mockResolver(
			dnsmock2.NewServer().
				Query("ns0.lego.localhost. A",
					dnsmock2.Answer(fakeA("ns0.lego.localhost."))).
				Query("ns1.lego.localhost. A",
					dnsmock2.Answer(fakeA("ns1.lego.localhost."))).
				Query("example.com. TXT",
					dnsmock2.Answer(
						fakeTXT("example.com.", "one"),
						fakeTXT("example.com.", "two"),
						fakeTXT("example.com.", "three"),
						fakeTXT("example.com.", "four"),
						fakeTXT("example.com.", "five"),
					),
				).
				Build(t),
		),
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
			expectedError: "did not return the expected TXT record [fqdn: acme-staging.api.example.com., value: fe01=]: one, two, three, four, five",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			DefaultClient().ClearFqdnCache()

			check := newPreCheck()
			check.requireRecursiveNssPropagation = false

			ok, err := check.checkDNSPropagation(t.Context(), test.fqdn, test.value)
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

func Test_preCheck_checkDNSPropagation_requireRecursiveNssPropagation(t *testing.T) {
	// The 2 DNS servers must have the same data as we required full propagation.
	builder := dnsmock2.NewServer().
		Query("ns0.lego.localhost. A",
			dnsmock2.Answer(fakeA("ns0.lego.localhost."))).
		Query("ns1.lego.localhost. A",
			dnsmock2.Answer(fakeA("ns1.lego.localhost."))).
		Query("example.com. TXT",
			dnsmock2.Answer(
				fakeTXT("example.com.", "one"),
				fakeTXT("example.com.", "two"),
				fakeTXT("example.com.", "three"),
				fakeTXT("example.com.", "four"),
				fakeTXT("example.com.", "five"),
			),
		).
		Query("example.com. SOA", dnsmock2.SOA("")).
		Query("example.com. NS",
			dnsmock2.Answer(
				fakeNS("example.com.", "ns0.lego.localhost."),
				fakeNS("example.com.", "ns1.lego.localhost."),
			),
		)

	mockDefault(t, builder.Build(t), mockResolver(builder.Build(t)))

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
			expectedError: "did not return the expected TXT record [fqdn: acme-staging.api.example.com., value: fe01=]: one, two, three, four, five",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			DefaultClient().ClearFqdnCache()

			check := newPreCheck()

			ok, err := check.checkDNSPropagation(t.Context(), test.fqdn, test.value)
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
