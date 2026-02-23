package dnspersist01

import (
	"testing"

	"github.com/go-acme/lego/v5/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_LookupTXT(t *testing.T) {
	fqdn := "_validation-persist.example.com."

	testCases := []struct {
		desc            string
		serverBuilder   *dnsmock.Builder
		expectedRecords []TXTRecord
		expectedChain   []string
	}{
		{
			desc: "direct TXT",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.Answer(fakeTXT(fqdn, "value", 120))),
			expectedRecords: []TXTRecord{{Value: "value", TTL: 120}},
		},
		{
			desc: "cname to txt",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock.Answer(fakeTXT("alias.example.com.", "value", 60))),
			expectedRecords: []TXTRecord{{Value: "value", TTL: 60}},
			expectedChain:   []string{"alias.example.com."},
		},
		{
			desc: "cname chain stops after one",
			serverBuilder: dnsmock.NewServer().
				Query(fqdn+" TXT", dnsmock.CNAME("alias.example.com.")).
				Query("alias.example.com. TXT", dnsmock.CNAME("alias2.example.com.")),
			expectedChain: []string{"alias.example.com."},
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
			resolver := NewResolver([]string{addr.String()})

			result, err := resolver.LookupTXT(fqdn)
			require.NoError(t, err)
			assert.Equal(t, test.expectedRecords, result.Records)
			assert.Equal(t, test.expectedChain, result.CNAMEChain)
		})
	}
}
