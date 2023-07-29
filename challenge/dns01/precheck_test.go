package dns01

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDNSPropagation(t *testing.T) {
	testCases := []struct {
		desc        string
		fqdn        string
		value       string
		expectError bool
	}{
		{
			desc:  "success",
			fqdn:  "postman-echo.com.",
			value: "postman-domain-verification=c85de626cb79d941310696e06558e2e790223802f3697dfbdcaf65510152d52c",
		},
		{
			desc:        "no TXT record",
			fqdn:        "acme-staging.api.letsencrypt.org.",
			value:       "fe01=",
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			ClearFqdnCache()

			check := newPreCheck()

			ok, err := check.checkDNSPropagation(test.fqdn, test.value)
			if test.expectError {
				assert.Errorf(t, err, "PreCheckDNS must fail for %s", test.fqdn)
				assert.False(t, ok, "PreCheckDNS must fail for %s", test.fqdn)
			} else {
				assert.NoErrorf(t, err, "PreCheckDNS failed for %s", test.fqdn)
				assert.True(t, ok, "PreCheckDNS failed for %s", test.fqdn)
			}
		})
	}
}

func TestCheckAuthoritativeNss(t *testing.T) {
	testCases := []struct {
		desc        string
		fqdn, value string
		ns          []string
		expected    bool
	}{
		{
			desc:     "TXT RR w/ expected value",
			fqdn:     "8.8.8.8.asn.routeviews.org.",
			value:    "151698.8.8.024",
			ns:       []string{"asnums.routeviews.org."},
			expected: true,
		},
		{
			desc: "No TXT RR",
			fqdn: "ns1.google.com.",
			ns:   []string{"ns2.google.com."},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			ClearFqdnCache()

			ok, _ := checkAuthoritativeNss(test.fqdn, test.value, test.ns)
			assert.Equal(t, test.expected, ok, test.fqdn)
		})
	}
}

func TestCheckAuthoritativeNssErr(t *testing.T) {
	testCases := []struct {
		desc        string
		fqdn, value string
		ns          []string
		error       string
	}{
		{
			desc:  "TXT RR /w unexpected value",
			fqdn:  "8.8.8.8.asn.routeviews.org.",
			value: "fe01=",
			ns:    []string{"asnums.routeviews.org."},
			error: "did not return the expected TXT record",
		},
		{
			desc:  "No TXT RR",
			fqdn:  "ns1.google.com.",
			value: "fe01=",
			ns:    []string{"ns2.google.com."},
			error: "did not return the expected TXT record",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			ClearFqdnCache()

			_, err := checkAuthoritativeNss(test.fqdn, test.value, test.ns)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.error)
		})
	}
}
