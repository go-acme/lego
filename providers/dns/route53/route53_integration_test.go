package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/config/env"
)

func TestLiveTTL(t *testing.T) {
	config, err := env.Get("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "R53_DOMAIN")
	if err != nil {
		t.Skip(err.Error())
	}

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	domain := config["R53_DOMAIN"]

	err = provider.Present(domain, "foo", "bar")
	require.NoError(t, err)

	// we need a separate R53 client here as the one in the DNS provider is unexported.
	fqdn := "_acme-challenge." + domain + "."
	sess, err := session.NewSession()
	require.NoError(t, err)
	svc := route53.New(sess)

	defer func() {
		errC := provider.CleanUp(domain, "foo", "bar")
		if errC != nil {
			t.Log(errC)
		}
	}()

	zoneID, err := provider.getHostedZoneID(fqdn)
	require.NoError(t, err)

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ListResourceRecordSets(params)
	require.NoError(t, err)

	for _, v := range resp.ResourceRecordSets {
		if aws.StringValue(v.Name) == fqdn && aws.StringValue(v.Type) == "TXT" && aws.Int64Value(v.TTL) == 10 {
			return
		}
	}

	t.Fatalf("Could not find a TXT record for _acme-challenge.%s with a TTL of 10", domain)
}
