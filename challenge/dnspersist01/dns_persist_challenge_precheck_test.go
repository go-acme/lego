package dnspersist01

import (
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_preCheck_checkDNSPropagation(t *testing.T) {
	addr := dnsmock.NewServer().
		Query("ns0.lego.localhost. A",
			dnsmock.Answer(fakeA("ns0.lego.localhost.", "127.0.0.1"))).
		Query("ns1.lego.localhost. A",
			dnsmock.Answer(fakeA("ns1.lego.localhost.", "127.0.0.1"))).
		Query("example.com. TXT",
			dnsmock.Answer(
				fakeTXT("example.com.", "one", 10),
				fakeTXT("example.com.", "two", 10),
				fakeTXT("example.com.", "three", 10),
				fakeTXT("example.com.", "four", 10),
				fakeTXT("example.com.", "five", 10),
			),
		).
		Query("acme-staging.api.example.com. TXT",
			dnsmock.Answer(
				fakeTXT("acme-staging.api.example.com.", "one", 10),
				fakeTXT("acme-staging.api.example.com.", "two", 10),
				fakeTXT("acme-staging.api.example.com.", "three", 10),
				fakeTXT("acme-staging.api.example.com.", "four", 10),
				fakeTXT("acme-staging.api.example.com.", "five", 10),
			),
		).
		Query("acme-staging.api.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
		Query("api.example.com. SOA", dnsmock.Error(dns.RcodeNameError)).
		Query("example.com. SOA", dnsmock.SOA("")).
		Query("example.com. NS",
			dnsmock.Answer(
				fakeNS("example.com.", "ns0.lego.localhost."),
				fakeNS("example.com.", "ns1.lego.localhost."),
			),
		).
		Build(t)

	mockDefault(t, addr, mockResolver(addr))

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
			expectedError: "did not return a matching TXT record [fqdn: acme-staging.api.example.com.]: one,two,three,four,five",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			match := func(records []TXTRecord) bool {
				for _, record := range records {
					if record.Value == test.value {
						return true
					}
				}

				return false
			}

			check := newPreCheck()
			check.requireRecursiveNssPropagation = false

			ok, err := check.checkDNSPropagation(t.Context(), test.fqdn, match)
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
