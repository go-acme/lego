package dnspersist01

import (
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_LookupTXT(t *testing.T) {
	fqdn := "_validation-persist.example.com."

	testCases := []struct {
		desc          string
		serverBuilder *dnsmock.Builder
		expected      TXTResult
	}{
		{
			desc: "direct TXT",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.Answer(fakeTXT(fqdn, "value", 120))),
			expected: TXTResult{
				Records: []TXTRecord{{Value: "value", TTL: 120}},
			},
		},
		{
			desc: "cname to txt",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock.Answer(fakeTXT("alias.example.com.", "value", 60))),
			expected: TXTResult{
				Records:    []TXTRecord{{Value: "value", TTL: 60}},
				CNAMEChain: []string{"alias.example.com."},
			},
		},
		{
			desc: "cname chain follows multiple hops",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock.CNAME("alias2.example.com.")).
				Query("alias2.example.com. TXT", dnsmock.Answer(fakeTXT("alias2.example.com.", "value", 30))),
			expected: TXTResult{
				Records:    []TXTRecord{{Value: "value", TTL: 30}},
				CNAMEChain: []string{"alias.example.com.", "alias2.example.com."},
			},
		},
		{
			desc: "nxdomain",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.Error(dns.RcodeNameError)),
		},
		{
			desc: "empty answer",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.Noop),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			addr := test.serverBuilder.Build(t)

			client := NewClient(&Options{
				RecursiveNameservers: []string{addr.String()},
			})

			result, err := client.LookupTXT(t.Context(), fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, result)
		})
	}
}
