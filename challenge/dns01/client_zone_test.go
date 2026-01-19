package dns01

import (
	"testing"

	"github.com/go-acme/lego/v5/platform/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestClient_FindZoneByFqdnCustom(t *testing.T) {
	for _, test := range lookupSoaByFqdnTestCases(t) {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(nil)

			zone, err := client.FindZoneByFqdnCustom(t.Context(), test.fqdn, test.nameservers)
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

func TestClient_FindZoneByFqdn(t *testing.T) {
	for _, test := range lookupSoaByFqdnTestCases(t) {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(nil)
			client.recursiveNameservers = test.nameservers

			zone, err := client.FindZoneByFqdn(t.Context(), test.fqdn)
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
