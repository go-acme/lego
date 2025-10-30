package selfhostde

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvRecordsMapping).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername:       "user",
				EnvPassword:       "secret",
				EnvRecordsMapping: "example.com:123",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvPassword:       "secret",
				EnvRecordsMapping: "example.com:123",
			},
			expected: "selfhostde: some credentials information are missing: SELFHOSTDE_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername:       "user",
				EnvRecordsMapping: "example.com:123",
			},
			expected: "selfhostde: some credentials information are missing: SELFHOSTDE_PASSWORD",
		},
		{
			desc: "missing records mapping",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
			},
			expected: "selfhostde: some credentials information are missing: SELFHOSTDE_RECORDS_MAPPING",
		},
		{
			desc: "invalid records mapping",
			envVars: map[string]string{
				EnvUsername:       "user",
				EnvPassword:       "secret",
				EnvRecordsMapping: "example.com",
			},
			expected: `selfhostde: malformed records mapping: missing ":": example.com`,
		},
		{
			desc:     "missing information",
			envVars:  map[string]string{},
			expected: "selfhostde: some credentials information are missing: SELFHOSTDE_USERNAME,SELFHOSTDE_PASSWORD,SELFHOSTDE_RECORDS_MAPPING",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		username      string
		password      string
		recordMapping map[string]*Seq
		expected      string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
			recordMapping: map[string]*Seq{
				"example.com": NewSeq("123"),
			},
		},
		{
			desc:     "missing username",
			password: "secret",
			recordMapping: map[string]*Seq{
				"example.com": NewSeq("123"),
			},
			expected: "selfhostde: credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			recordMapping: map[string]*Seq{
				"example.com": NewSeq("123"),
			},
			expected: "selfhostde: credentials missing",
		},
		{
			desc:     "missing sequence",
			username: "user",
			password: "secret",
			recordMapping: map[string]*Seq{
				"example.com": nil,
			},
			expected: `selfhostde: missing record ID for "example.com"`,
		},
		{
			desc:     "empty sequence",
			username: "user",
			password: "secret",
			recordMapping: map[string]*Seq{
				"example.com": NewSeq(),
			},
			expected: `selfhostde: missing record ID for "example.com"`,
		},
		{
			desc:     "missing records mapping",
			username: "user",
			password: "secret",
			expected: "selfhostde: missing record mapping",
		},
		{
			desc:          "empty records mapping",
			username:      "user",
			password:      "secret",
			recordMapping: map[string]*Seq{},
			expected:      "selfhostde: missing record mapping",
		},
		{
			desc:     "missing information",
			expected: "selfhostde: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.RecordsMapping = test.recordMapping

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
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
