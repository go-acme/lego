package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/config/env"
)

func TestRoute53TTL(t *testing.T) {
	config, err := env.Get("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "R53_DOMAIN")
	if err != nil {
		t.Skip(err.Error())
	}

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(config["R53_DOMAIN"], "foo", "bar")
	require.NoError(t, err)

	// we need a separate R53 client here as the one in the DNS provider is
	// unexported.
	fqdn := "_acme-challenge." + config["R53_DOMAIN"] + "."
	svc := route53.New(session.New())
	zoneID, err := provider.getHostedZoneID(fqdn)
	if err != nil {
		provider.CleanUp(config["R53_DOMAIN"], "foo", "bar")
		t.Fatal(err)
	}

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ListResourceRecordSets(params)
	if err != nil {
		provider.CleanUp(config["R53_DOMAIN"], "foo", "bar")
		t.Fatal(err)
	}

	for _, v := range resp.ResourceRecordSets {
		if aws.StringValue(v.Name) == fqdn && aws.StringValue(v.Type) == "TXT" && aws.Int64Value(v.TTL) == 10 {
			provider.CleanUp(config["R53_DOMAIN"], "foo", "bar")
			return
		}
	}

	provider.CleanUp(config["R53_DOMAIN"], "foo", "bar")
	t.Fatalf("Could not find a TXT record for _acme-challenge.%s with a TTL of 10", config["R53_DOMAIN"])
}
