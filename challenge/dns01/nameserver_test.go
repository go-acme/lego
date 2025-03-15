package dns01

import (
	"errors"
	"sort"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupNameserversOK(t *testing.T) {
	testCases := []struct {
		fqdn string
		nss  []string
	}{
		{
			fqdn: "en.wikipedia.org.",
			nss:  []string{"ns0.wikimedia.org.", "ns1.wikimedia.org.", "ns2.wikimedia.org."},
		},
		{
			fqdn: "www.google.com.",
			nss:  []string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
		},
		{
			fqdn: "mail.proton.me.",
			nss:  []string{"ns1.proton.me.", "ns2.proton.me.", "ns3.proton.me."},
		},
	}

	for _, test := range testCases {
		t.Run(test.fqdn, func(t *testing.T) {
			t.Parallel()

			nss, err := lookupNameservers(test.fqdn)
			require.NoError(t, err)

			sort.Strings(nss)
			sort.Strings(test.nss)

			assert.EqualValues(t, test.nss, nss)
		})
	}
}

func TestLookupNameserversErr(t *testing.T) {
	testCases := []struct {
		desc  string
		fqdn  string
		error string
	}{
		{
			desc:  "invalid tld",
			fqdn:  "_null.n0n0.",
			error: "could not find zone",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := lookupNameservers(test.fqdn)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.error)
		})
	}
}

var findXByFqdnTestCases = []struct {
	desc          string
	fqdn          string
	zone          string
	primaryNs     string
	nameservers   []string
	expectedError string
}{
	{
		desc:        "domain is a CNAME",
		fqdn:        "mail.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a non-existent subdomain",
		fqdn:        "foo.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a eTLD",
		fqdn:        "example.com.ac.",
		zone:        "ac.",
		primaryNs:   "a0.nic.ac.",
		nameservers: recursiveNameservers,
	},
	{
		desc:        "domain is a cross-zone CNAME",
		fqdn:        "cross-zone-example.assets.sh.",
		zone:        "assets.sh.",
		primaryNs:   "gina.ns.cloudflare.com.",
		nameservers: recursiveNameservers,
	},
	{
		desc:          "NXDOMAIN",
		fqdn:          "test.lego.zz.",
		zone:          "lego.zz.",
		nameservers:   []string{"8.8.8.8:53"},
		expectedError: "[fqdn=test.lego.zz.] could not find the start of authority for 'test.lego.zz.' [question='zz. IN  SOA', code=NXDOMAIN]",
	},
	{
		desc:        "several non existent nameservers",
		fqdn:        "mail.google.com.",
		zone:        "google.com.",
		primaryNs:   "ns1.google.com.",
		nameservers: []string{":7053", ":8053", "8.8.8.8:53"},
	},
	{
		desc:        "only non-existent nameservers",
		fqdn:        "mail.google.com.",
		zone:        "google.com.",
		nameservers: []string{":7053", ":8053", ":9053"},
		// use only the start of the message because the port changes with each call: 127.0.0.1:XXXXX->127.0.0.1:7053.
		expectedError: "[fqdn=mail.google.com.] could not find the start of authority for 'mail.google.com.': DNS call error: read udp ",
	},
	{
		desc:          "no nameservers",
		fqdn:          "test.ldez.com.",
		zone:          "ldez.com.",
		nameservers:   []string{},
		expectedError: "[fqdn=test.ldez.com.] could not find the start of authority for 'test.ldez.com.': empty list of nameservers",
	},
}

func TestFindZoneByFqdnCustom(t *testing.T) {
	for _, test := range findXByFqdnTestCases {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			zone, err := FindZoneByFqdnCustom(test.fqdn, test.nameservers)
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

func TestFindPrimaryNsByFqdnCustom(t *testing.T) {
	for _, test := range findXByFqdnTestCases {
		t.Run(test.desc, func(t *testing.T) {
			ClearFqdnCache()

			ns, err := FindPrimaryNsByFqdnCustom(test.fqdn, test.nameservers)
			if test.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.primaryNs, ns)
			}
		})
	}
}

func TestResolveConfServers(t *testing.T) {
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

func TestDNSError_Error(t *testing.T) {
	msgIn := createDNSMsg("example.com.", dns.TypeTXT, true)

	msgOut := createDNSMsg("example.org.", dns.TypeSOA, true)
	msgOut.Rcode = dns.RcodeNameError

	testCases := []struct {
		desc     string
		err      *DNSError
		expected string
	}{
		{
			desc:     "empty error",
			err:      &DNSError{},
			expected: "DNS error",
		},
		{
			desc: "all fields",
			err: &DNSError{
				Message: "Oops",
				NS:      "example.com.",
				MsgIn:   msgIn,
				MsgOut:  msgOut,
				Err:     errors.New("I did it again"),
			},
			expected: "Oops: I did it again [ns=example.com., question='example.com. IN  TXT', code=NXDOMAIN]",
		},
		{
			desc: "only NS",
			err: &DNSError{
				NS: "example.com.",
			},
			expected: "DNS error [ns=example.com.]",
		},
		{
			desc: "only MsgIn",
			err: &DNSError{
				MsgIn: msgIn,
			},
			expected: "DNS error [question='example.com. IN  TXT']",
		},
		{
			desc: "only MsgOut",
			err: &DNSError{
				MsgOut: msgOut,
			},
			expected: "DNS error [question='example.org. IN  SOA', code=NXDOMAIN]",
		},
		{
			desc: "only Err",
			err: &DNSError{
				Err: errors.New("I did it again"),
			},
			expected: "DNS error: I did it again",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.EqualError(t, test.err, test.expected)
		})
	}
}
