package sakuracloud

import (
	"fmt"
	"sync"
	"testing"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) {
	t.Helper()

	t.Setenv("SAKURACLOUD_FAKE_MODE", "1")

	createDummyZone(t, fakeCaller())
}

func fakeCaller() iaas.APICaller {
	return api.NewCallerWithOptions(&api.CallerOptions{
		Options: &client.Options{
			AccessToken:       "dummy",
			AccessTokenSecret: "dummy",
		},
		FakeMode: true,
	})
}

func createDummyZone(t *testing.T, caller iaas.APICaller) {
	t.Helper()

	ctx := t.Context()

	dnsOp := iaas.NewDNSOp(caller)

	// cleanup
	zones, err := dnsOp.Find(ctx, &iaas.FindCondition{})
	require.NoError(t, err)

	for _, zone := range zones.DNS {
		if zone.Name == "example.com" {
			err = dnsOp.Delete(ctx, zone.ID)
			require.NoError(t, err)

			break
		}
	}

	// create dummy zone
	_, err = iaas.NewDNSOp(caller).Create(t.Context(), &iaas.DNSCreateRequest{Name: "example.com"})
	require.NoError(t, err)
}

func TestDNSProvider_addAndCleanupRecords(t *testing.T) {
	setupTest(t)

	config := NewDefaultConfig()
	config.Token = "token1"
	config.Secret = "secret1"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	t.Run("addTXTRecord", func(t *testing.T) {
		ctx := t.Context()

		err = p.addTXTRecord(ctx, "test.example.com.", "dummyValue", 10)
		require.NoError(t, err)

		updZone, e := p.getHostedZone(ctx, "test.example.com.")
		require.NoError(t, e)
		require.NotNil(t, updZone)

		require.Len(t, updZone.Records, 1)
	})

	t.Run("cleanupTXTRecord", func(t *testing.T) {
		ctx := t.Context()

		err = p.cleanupTXTRecord(ctx, "test.example.com.", "dummyValue")
		require.NoError(t, err)

		updZone, e := p.getHostedZone(ctx, "test.example.com.")
		require.NoError(t, e)
		require.NotNil(t, updZone)

		require.Empty(t, updZone.Records)
	})
}

func TestDNSProvider_concurrentAddAndCleanupRecords(t *testing.T) {
	setupTest(t)

	dummyRecordCount := 10

	var providers []*DNSProvider

	for range dummyRecordCount {
		config := NewDefaultConfig()
		config.Token = "token3"
		config.Secret = "secret3"

		p, err := NewDNSProviderConfig(config)
		require.NoError(t, err)

		providers = append(providers, p)
	}

	var wg sync.WaitGroup

	t.Run("addTXTRecord", func(t *testing.T) {
		wg.Add(len(providers))

		ctx := t.Context()

		for i, p := range providers {
			go func(j int, client *DNSProvider) {
				err := client.addTXTRecord(ctx, fmt.Sprintf("test%d.example.com.", j), "dummyValue", 10)
				require.NoError(t, err)
				wg.Done()
			}(i, p)
		}

		wg.Wait()

		updZone, err := providers[0].getHostedZone(ctx, "example.com.")
		require.NoError(t, err)
		require.NotNil(t, updZone)

		require.Len(t, updZone.Records, dummyRecordCount)
	})

	t.Run("cleanupTXTRecord", func(t *testing.T) {
		wg.Add(len(providers))

		ctx := t.Context()

		for i, p := range providers {
			go func(i int, client *DNSProvider) {
				err := client.cleanupTXTRecord(ctx, fmt.Sprintf("test%d.example.com.", i), "dummyValue")
				require.NoError(t, err)
				wg.Done()
			}(i, p)
		}

		wg.Wait()

		updZone, err := providers[0].getHostedZone(ctx, "example.com.")
		require.NoError(t, err)
		require.NotNil(t, updZone)

		require.Empty(t, updZone.Records)
	})
}
