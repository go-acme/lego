package easydns

import (
	"testing"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("EASYDNS_HOSTNAME", "EASYDNS_TOKEN", "EASYDNS_SECRET")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"EASYDNS_TOKEN":  "TOKEN",
				"EASYDNS_SECRET": "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				"EASYDNS_SECRET": "SECRET",
			},
			expected: "easydns: the API token is missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				"EASYDNS_TOKEN": "TOKEN",
			},
			expected: "easydns: the API secret is missing: EASYDNS_SECRET",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestGetHost(t *testing.T) {
	testCases := []struct {
		desc           string
		fqdn           string
		expectedHost   string
		expectedDoamin string
	}{
		{
			desc:           "single-part host",
			fqdn:           "_acme-challenge.domain.com",
			expectedHost:   "_acme-challenge",
			expectedDoamin: "domain.com",
		},
		{
			desc:           "multi-part host",
			fqdn:           "_acme-challenge.sub.domain.com",
			expectedHost:   "_acme-challenge.sub",
			expectedDoamin: "domain.com",
		},
		{
			desc:           "trailing dot",
			fqdn:           "_acme-challenge.sub.domain.com.",
			expectedHost:   "_acme-challenge.sub",
			expectedDoamin: "domain.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actualHost, actualDomain := getHost(test.fqdn)

			require.Equal(t, test.expectedHost, actualHost)
			require.Equal(t, test.expectedDoamin, actualDomain)
		})
	}
}
