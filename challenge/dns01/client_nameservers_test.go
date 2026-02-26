package dns01

import (
	"testing"

	dnsmock2 "github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_checkNameserversPropagationCustom_authoritativeNss(t *testing.T) {
	testCases := []struct {
		desc          string
		fqdn, value   string
		fakeDNSServer *dnsmock2.Builder
		expectedError string
	}{
		{
			desc: "TXT RR w/ expected value",
			// NS: asnums.routeviews.org.
			fqdn:  "8.8.8.8.asn.routeviews.org.",
			value: "151698.8.8.024",
			fakeDNSServer: dnsmock2.NewServer().
				Query("8.8.8.8.asn.routeviews.org. TXT",
					dnsmock2.Answer(
						fakeTXT("8.8.8.8.asn.routeviews.org.", "151698.8.8.024"),
					),
				),
		},
		{
			desc: "TXT RR w/ unexpected value",
			// NS: asnums.routeviews.org.
			fqdn:  "8.8.8.8.asn.routeviews.org.",
			value: "fe01=",
			fakeDNSServer: dnsmock2.NewServer().
				Query("8.8.8.8.asn.routeviews.org. TXT",
					dnsmock2.Answer(
						fakeTXT("8.8.8.8.asn.routeviews.org.", "15169"),
						fakeTXT("8.8.8.8.asn.routeviews.org.", "8.8.8.0"),
						fakeTXT("8.8.8.8.asn.routeviews.org.", "24"),
					),
				),
			expectedError: "did not return the expected TXT record [fqdn: 8.8.8.8.asn.routeviews.org., value: fe01=]: 15169, 8.8.8.0, 24",
		},
		{
			desc: "No TXT RR",
			// NS: ns2.google.com.
			fqdn:  "ns1.google.com.",
			value: "fe01=",
			fakeDNSServer: dnsmock2.NewServer().
				Query("ns1.google.com.", dnsmock2.Noop),
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
