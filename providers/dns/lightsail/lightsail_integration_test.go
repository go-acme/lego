package lightsail

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	"github.com/stretchr/testify/require"
)

func TestLiveTTL(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	domain := envTest.GetDomain()

	err = provider.Present(domain, "foo", "bar")
	require.NoError(t, err)

	// we need a separate Lightsail client here as the one in the DNS provider is unexported.
	fqdn := "_acme-challenge." + domain

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	require.NoError(t, err)

	svc := lightsail.NewFromConfig(cfg)
	require.NoError(t, err)

	defer func() {
		errC := provider.CleanUp(domain, "foo", "bar")
		if errC != nil {
			t.Log(errC)
		}
	}()

	params := &lightsail.GetDomainInput{
		DomainName: aws.String(domain),
	}

	resp, err := svc.GetDomain(ctx, params)
	require.NoError(t, err)

	entries := resp.Domain.DomainEntries
	for _, entry := range entries {
		if ptr.Deref(entry.Type) == "TXT" && ptr.Deref(entry.Name) == fqdn {
			return
		}
	}

	t.Fatalf("Could not find a TXT record for _acme-challenge.%s", domain)
}
