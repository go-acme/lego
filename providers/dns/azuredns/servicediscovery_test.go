package azuredns

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createGraphQuery(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Config
		expected string
	}{
		{
			desc: "empty configuration (public)",
			cfg:  &Config{},
			expected: `
resources
| where type =~ "microsoft.network/dnszones"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "SubscriptionID (public)",
			cfg: &Config{
				SubscriptionID: "123",
			},
			expected: `
resources
| where type =~ "microsoft.network/dnszones"
| where subscriptionId =~ "123"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "ResourceGroup (public)",
			cfg: &Config{
				ResourceGroup: "123",
			},
			expected: `
resources
| where type =~ "microsoft.network/dnszones"
| where resourceGroup =~ "123"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "ServiceDiscoveryFilter (public)",
			cfg: &Config{
				ServiceDiscoveryFilter: "123",
			},
			expected: `
resources
| where type =~ "microsoft.network/dnszones"
| 123
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "empty configuration (private)",
			cfg: &Config{
				PrivateZone: true,
			},
			expected: `
resources
| where type =~ "microsoft.network/privatednszones"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "SubscriptionID (private)",
			cfg: &Config{
				SubscriptionID: "123",
				PrivateZone:    true,
			},
			expected: `
resources
| where type =~ "microsoft.network/privatednszones"
| where subscriptionId =~ "123"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "ResourceGroup (private)",
			cfg: &Config{
				ResourceGroup: "123",
				PrivateZone:   true,
			},
			expected: `
resources
| where type =~ "microsoft.network/privatednszones"
| where resourceGroup =~ "123"
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "ServiceDiscoveryFilter (private)",
			cfg: &Config{
				ServiceDiscoveryFilter: "123",
				PrivateZone:            true,
			},
			expected: `
resources
| where type =~ "microsoft.network/privatednszones"
| 123
| project subscriptionId, resourceGroup, name`,
		},
		{
			desc: "all (private)",
			cfg: &Config{
				SubscriptionID:         "123",
				ResourceGroup:          "456",
				ServiceDiscoveryFilter: "789",
				PrivateZone:            true,
			},
			expected: `
resources
| where type =~ "microsoft.network/privatednszones"
| where subscriptionId =~ "123"
| where resourceGroup =~ "456"
| 789
| project subscriptionId, resourceGroup, name`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			query := createGraphQuery(test.cfg)
			assert.Equal(t, strings.ReplaceAll(test.expected, "\r", ""), query)
		})
	}
}
