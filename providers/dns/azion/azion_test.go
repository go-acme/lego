package azion

import (
	"testing"

	"github.com/aziontech/azionapi-go-sdk/idns"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvPersonalToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvPersonalToken: "token",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvPersonalToken: "",
			},
			expected: "azion: some credentials information are missing: AZION_PERSONAL_TOKEN",
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
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		token    string
		expected string
	}{
		{
			desc:  "success",
			token: "token",
		},
		{
			desc:     "missing credentials",
			expected: "azion: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PersonalToken = test.token

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestFindParentZone(t *testing.T) {
	// Create a mock DNSProvider
	provider := &DNSProvider{}

	// Create mock zones - simulating that only example.com is registered
	mockZones := []idns.Zone{
		{
			Id:   idns.PtrInt32(1),
			Name: idns.PtrString("example.com"),
		},
		{
			Id:   idns.PtrInt32(2),
			Name: idns.PtrString("another.org"),
		},
	}

	testCases := []struct {
		desc        string
		fqdn        string
		expectedID  int32
		expectedErr bool
	}{
		{
			desc:        "exact match - example.com",
			fqdn:        "example.com.",
			expectedID:  1,
			expectedErr: false,
		},
		{
			desc:        "subdomain - test.example.com should find example.com",
			fqdn:        "test.example.com.",
			expectedID:  1,
			expectedErr: false,
		},
		{
			desc:        "deep subdomain - _acme-challenge.test.example.com should find example.com",
			fqdn:        "_acme-challenge.test.example.com.",
			expectedID:  1,
			expectedErr: false,
		},
		{
			desc:        "wildcard subdomain - *.test.example.com should find example.com",
			fqdn:        "_acme-challenge.test.example.com.",
			expectedID:  1,
			expectedErr: false,
		},
		{
			desc:        "no parent zone found",
			fqdn:        "notfound.net.",
			expectedID:  0,
			expectedErr: true,
		},
		{
			desc:        "another domain exact match",
			fqdn:        "another.org.",
			expectedID:  2,
			expectedErr: false,
		},
		{
			desc:        "subdomain of another domain",
			fqdn:        "sub.another.org.",
			expectedID:  2,
			expectedErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			zone, err := provider.findParentZone(mockZones, test.fqdn)

			if test.expectedErr {
				require.Error(t, err)
				require.Nil(t, zone)
			} else {
				require.NoError(t, err)
				require.NotNil(t, zone)
				require.Equal(t, test.expectedID, zone.GetId())
			}
		})
	}
}
