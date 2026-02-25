package dnspersist01

import (
	"testing"

	dnsmock2 "github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_LookupTXT(t *testing.T) {
	fqdn := "_validation-persist.example.com."

	testCases := []struct {
		desc          string
		serverBuilder *dnsmock2.Builder
		expected      TXTResult
	}{
		{
			desc: "direct TXT",
			serverBuilder: dnsmock2.NewServer().
				Query(fqdn+" TXT", dnsmock2.Answer(fakeTXT(fqdn, "value", 120))),
			expected: TXTResult{
				Records: []TXTRecord{{Value: "value", TTL: 120}},
			},
		},
		{
			desc: "cname to txt",
			serverBuilder: dnsmock2.NewServer().
				Query(fqdn+" TXT", dnsmock2.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock2.Answer(fakeTXT("alias.example.com.", "value", 60))),
			expected: TXTResult{
				Records:    []TXTRecord{{Value: "value", TTL: 60}},
				CNAMEChain: []string{"alias.example.com."},
			},
		},
		{
			desc: "cname chain follows multiple hops",
			serverBuilder: dnsmock2.NewServer().
				Query(fqdn+" TXT", dnsmock2.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock2.CNAME("alias2.example.com.")).
				Query("alias2.example.com. TXT", dnsmock2.Answer(fakeTXT("alias2.example.com.", "value", 30))),
			expected: TXTResult{
				Records:    []TXTRecord{{Value: "value", TTL: 30}},
				CNAMEChain: []string{"alias.example.com.", "alias2.example.com."},
			},
		},
		{
			desc: "nxdomain",
			serverBuilder: dnsmock2.NewServer().
				Query(fqdn+" TXT", dnsmock2.Error(dns.RcodeNameError)),
		},
		{
			desc: "empty answer",
			serverBuilder: dnsmock2.NewServer().
				Query(fqdn+" TXT", dnsmock2.Noop),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			addr := test.serverBuilder.Build(t)

			resolver := NewResolver([]string{addr.String()})

			result, err := resolver.LookupTXT(fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, result)
		})
	}
}
