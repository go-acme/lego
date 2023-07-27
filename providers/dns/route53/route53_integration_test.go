package route53

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
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

	// we need a separate R53 client here as the one in the DNS provider is unexported.
	fqdn := "_acme-challenge." + domain + "."

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	require.NoError(t, err)

	svc := route53.NewFromConfig(cfg)

	defer func() {
		errC := provider.CleanUp(domain, "foo", "bar")
		if errC != nil {
			t.Log(errC)
		}
	}()

	zoneID, err := provider.getHostedZoneID(context.Background(), fqdn)
	require.NoError(t, err)

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ListResourceRecordSets(ctx, params)
	require.NoError(t, err)

	for _, v := range resp.ResourceRecordSets {
		if deref(v.Name) == fqdn && v.Type == "TXT" && deref(v.TTL) == 10 {
			return
		}
	}

	t.Fatalf("Could not find a TXT record for _acme-challenge.%s with a TTL of 10", domain)
}
