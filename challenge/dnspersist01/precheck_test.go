package dnspersist01

import (
	"testing"

	dnsmock2 "github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_preCheck_checkDNSPropagation(t *testing.T) {
	addr := dnsmock2.NewServer().
		Query("ns0.lego.localhost. A",
			dnsmock2.Answer(fakeA("ns0.lego.localhost.", "127.0.0.1"))).
		Query("ns1.lego.localhost. A",
			dnsmock2.Answer(fakeA("ns1.lego.localhost.", "127.0.0.1"))).
		Query("example.com. TXT",
			dnsmock2.Answer(
				fakeTXT("example.com.", "one", 10),
				fakeTXT("example.com.", "two", 10),
				fakeTXT("example.com.", "three", 10),
				fakeTXT("example.com.", "four", 10),
				fakeTXT("example.com.", "five", 10),
			),
		).
		Query("acme-staging.api.example.com. TXT",
			dnsmock2.Answer(
				fakeTXT("acme-staging.api.example.com.", "one", 10),
				fakeTXT("acme-staging.api.example.com.", "two", 10),
				fakeTXT("acme-staging.api.example.com.", "three", 10),
				fakeTXT("acme-staging.api.example.com.", "four", 10),
				fakeTXT("acme-staging.api.example.com.", "five", 10),
			),
		).
		Query("acme-staging.api.example.com. SOA", dnsmock2.Error(dns.RcodeNameError)).
		Query("api.example.com. SOA", dnsmock2.Error(dns.RcodeNameError)).
		Query("example.com. SOA", dnsmock2.SOA("")).
		Query("example.com. NS",
			dnsmock2.Answer(
				fakeNS("example.com.", "ns0.lego.localhost."),
				fakeNS("example.com.", "ns1.lego.localhost."),
			),
		).
		Build(t)

	chlg := &Challenge{
		resolver:             NewResolver([]string{addr.String()}),
		preCheck:             newPreCheck(),
		recursiveNameservers: ParseNameservers([]string{addr.String()}),
		authoritativeNSPort:  mockResolver(t, addr),
	}

	testCases := []struct {
		desc          string
		fqdn          string
		value         string
		expectedError bool
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
			expectedError: true,
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

			ok, err := chlg.checkDNSPropagation(test.fqdn, match)
			if test.expectedError {
				require.Error(t, err)
				assert.False(t, ok)
			} else {
				require.NoError(t, err)
				assert.True(t, ok)
			}
		})
	}
}
