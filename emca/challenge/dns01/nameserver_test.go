package dns01

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupNameserversOK(t *testing.T) {
	testCases := []struct {
		fqdn string
		nss  []string
	}{
		{
			fqdn: "books.google.com.ng.",
			nss:  []string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
		},
		{
			fqdn: "www.google.com.",
			nss:  []string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
		},
		{
			fqdn: "physics.georgetown.edu.",
			nss:  []string{"ns1.georgetown.edu.", "ns2.georgetown.edu.", "ns3.georgetown.edu."},
		},
	}

	for _, test := range testCases {
		test := test
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
			error: "could not determine the zone",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := lookupNameservers(test.fqdn)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.error)
		})
	}
}

func TestFindZoneByFqdn(t *testing.T) {
	testCases := []struct {
		desc string
		fqdn string
		zone string
	}{
		{
			desc: "domain is a CNAME",
			fqdn: "mail.google.com.",
			zone: "google.com.",
		},
		{
			desc: "domain is a non-existent subdomain",
			fqdn: "foo.google.com.",
			zone: "google.com.",
		},
		{
			desc: "domain is a eTLD",
			fqdn: "example.com.ac.",
			zone: "ac.",
		},
		{
			desc: "domain is a cross-zone CNAME",
			fqdn: "cross-zone-example.assets.sh.",
			zone: "assets.sh.",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := FindZoneByFqdn(test.fqdn, RecursiveNameservers)
			require.NoError(t, err)

			assert.Equal(t, test.zone, zone)
		})
	}
}

func TestResolveConfServers(t *testing.T) {
	var testCases = []struct {
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
