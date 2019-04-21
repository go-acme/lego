package joker

import (
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("JOKER_API_KEY").WithDomain("JOKER_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"JOKER_API_KEY": "123",
			},
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"JOKER_API_KEY": "",
			},
			expected: "joker: some credentials information are missing: JOKER_API_KEY",
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
				assert.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		authKey  string
		expected string
	}{
		{
			desc:    "success",
			authKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "joker: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.authKey

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestRemoveTxtEntryFromZone(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
		modified bool
	}{
		{
			desc:     "empty zone",
			input:    "",
			expected: "",
			modified: false,
		},
		{
			desc:     "zone with only A entry",
			input:    "@ A 0 192.0.2.2 3600",
			expected: "@ A 0 192.0.2.2 3600",
			modified: false,
		},
		{
			desc:     "zone with only clenup entry",
			input:    "_acme-challenge TXT 0  \"old \" 120",
			expected: "",
			modified: true,
		},
		{
			desc:     "zone with one A and one cleanup entries",
			input:    "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0  \"old \" 120",
			expected: "@ A 0 192.0.2.2 3600",
			modified: true,
		},
		{
			desc:     "zone with one A and multiple cleanup entries",
			input:    "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0  \"old \" 120\n_acme-challenge TXT 0  \"another \" 120",
			expected: "@ A 0 192.0.2.2 3600",
			modified: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			zone, modified := removeTxtEntryFromZone(test.input, "_acme-challenge")
			assert.Equal(t, zone, test.expected)
			assert.Equal(t, modified, test.modified)
		})
	}

}

func TestAddTxtEntryToZone(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "empty zone",
			input:    "",
			expected: "_acme-challenge TXT 0 \"test\" 120",
		},
		{
			desc:     "zone with A entry",
			input:    "@ A 0 192.0.2.2 3600",
			expected: "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0 \"test\" 120",
		},
		{
			desc:     "zone with required clenup entry",
			input:    "_acme-challenge TXT 0  \"old \" 120",
			expected: "_acme-challenge TXT 0 \"old\" 120\n_acme-challenge TXT 0 \"test\" 120",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			zone := addTxtEntryToZone(test.input, "_acme-challenge", "test", 120)
			assert.Equal(t, zone, test.expected)
		})
	}

}

func TestFixTxtLines(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "clean-up",
			input:    `_acme-challenge TXT 0  "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE " 120`,
			expected: `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
		},
		{
			desc:     "already cleaned",
			input:    `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
			expected: `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
		},
		{
			desc:     "special DNS entry",
			input:    "$dyndns=yes:username:password",
			expected: "$dyndns=yes:username:password",
		},
		{
			desc:     "SRV entry",
			input:    "_jabber._tcp SRV 20/0 xmpp-server1.l.google.com:5269 300",
			expected: "_jabber._tcp SRV 20/0 xmpp-server1.l.google.com:5269 300",
		},
		{
			desc:     "MX entry",
			input:    "@ MX 10 ASPMX.L.GOOGLE.COM 300",
			expected: "@ MX 10 ASPMX.L.GOOGLE.COM 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			line := fixTxtLines(test.input)
			assert.Equal(t, line, test.expected)
		})
	}

}
