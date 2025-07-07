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
	provider := &DNSProvider{}

	// Mock zones with various domain levels to test hierarchical search
	mockZones := []idns.Zone{
		{
			Id:   idns.PtrInt32(1),
			Name: idns.PtrString("example.com"),
		},
		{
			Id:   idns.PtrInt32(2),
			Name: idns.PtrString("sub.example.com"),
		},
		{
			Id:   idns.PtrInt32(3),
			Name: idns.PtrString("test.example.org"),
		},
		{
			Id:   idns.PtrInt32(4),
			Name: idns.PtrString("example.org"),
		},
	}

	testCases := []struct {
		desc         string
		fqdn         string
		expectedZone *idns.Zone
		expectedErr  bool
	}{
		{
			desc: "exact match - example.com",
			fqdn: "example.com.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(1),
				Name: idns.PtrString("example.com"),
			},
		},
		{
			desc: "exact match - subdomain zone",
			fqdn: "sub.example.com.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(2),
				Name: idns.PtrString("sub.example.com"),
			},
		},
		{
			desc: "find closest parent - should find sub.example.com",
			fqdn: "_acme-challenge.api.sub.example.com.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(2),
				Name: idns.PtrString("sub.example.com"),
			},
		},
		{
			desc: "find parent when subdomain not registered - should find example.com",
			fqdn: "_acme-challenge.test.example.com.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(1),
				Name: idns.PtrString("example.com"),
			},
		},
		{
			desc: "deep subdomain - should find example.org",
			fqdn: "_acme-challenge.api.staging.example.org.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(4),
				Name: idns.PtrString("example.org"),
			},
		},
		{
			desc: "find specific subdomain zone - should find test.example.org",
			fqdn: "_acme-challenge.api.test.example.org.",
			expectedZone: &idns.Zone{
				Id:   idns.PtrInt32(3),
				Name: idns.PtrString("test.example.org"),
			},
		},
		{
			desc:        "no parent zone found",
			fqdn:        "_acme-challenge.notfound.net.",
			expectedErr: true,
		},
		{
			desc:        "empty zones list",
			fqdn:        "example.com.",
			expectedErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zones := mockZones
			if test.desc == "empty zones list" {
				zones = []idns.Zone{}
			}

			zone, err := provider.findParentZone(zones, test.fqdn)

			if test.expectedErr {
				require.Error(t, err)
				require.Nil(t, zone)
			} else {
				require.NoError(t, err)
				require.NotNil(t, zone)
				require.Equal(t, test.expectedZone.GetId(), zone.GetId())
				require.Equal(t, test.expectedZone.GetName(), zone.GetName())
			}
		})
	}
}

// TestFindParentZoneWildcardScenario tests the specific wildcard certificate scenario
// where only the parent domain is registered but wildcard certificates are requested.
func TestFindParentZoneWildcardScenario(t *testing.T) {
	provider := &DNSProvider{}

	// Simulate only example.com registered as zone (common real-world scenario)
	mockZones := []idns.Zone{
		{
			Id:   idns.PtrInt32(123),
			Name: idns.PtrString("example.com"),
		},
	}

	testCases := []struct {
		desc string
		fqdn string
	}{
		{
			desc: "wildcard certificate for *.test.example.com",
			fqdn: "_acme-challenge.test.example.com.",
		},
		{
			desc: "wildcard certificate for *.api.example.com",
			fqdn: "_acme-challenge.api.example.com.",
		},
		{
			desc: "deep subdomain certificate",
			fqdn: "_acme-challenge.api.v1.example.com.",
		},
		{
			desc: "multiple level subdomain",
			fqdn: "_acme-challenge.app.staging.test.example.com.",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := provider.findParentZone(mockZones, test.fqdn)

			// Should find the parent zone example.com
			require.NoError(t, err, "Should find parent zone example.com")
			require.NotNil(t, zone, "Zone should not be nil")
			require.Equal(t, int32(123), zone.GetId(), "Should find zone example.com (ID 123)")
			require.Equal(t, "example.com", zone.GetName(), "Zone name should be example.com")
		})
	}

	// Test negative case
	t.Run("domain not found", func(t *testing.T) {
		t.Parallel()

		zone, err := provider.findParentZone(mockZones, "_acme-challenge.notfound.net.")
		require.Error(t, err, "Should return error for domain not found")
		require.Nil(t, zone, "Zone should be nil for domain not found")
		require.Contains(t, err.Error(), "no parent zone found", "Error should mention no parent zone found")
	})
}
