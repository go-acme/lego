package lightsail

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/stretchr/testify/require"
)

func TestLightsailTTL(t *testing.T) {
	m, err := testGetAndPreCheck()
	if err != nil {
		t.Skip(err.Error())
	}

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	lightsailDomain := m["lightsailDomain"]

	err = provider.Present(lightsailDomain, "foo", "bar")
	require.NoError(t, err)

	// we need a separate Lightshail client here as the one in the DNS provider is
	// unexported.
	fqdn := "_acme-challenge." + lightsailDomain
	sess, err := session.NewSession()
	require.NoError(t, err)

	svc := lightsail.New(sess)
	require.NoError(t, err)

	defer func() {
		errC := provider.CleanUp(lightsailDomain, "foo", "bar")
		if errC != nil {
			t.Log(errC)
		}
	}()

	params := &lightsail.GetDomainInput{
		DomainName: aws.String(lightsailDomain),
	}

	resp, err := svc.GetDomain(params)
	require.NoError(t, err)

	entries := resp.Domain.DomainEntries
	for _, entry := range entries {
		if *entry.Type == "TXT" && *entry.Name == fqdn {
			return
		}
	}

	t.Fatalf("Could not find a TXT record for _acme-challenge.%s", lightsailDomain)
}

func testGetAndPreCheck() (map[string]string, error) {
	m := map[string]string{
		"lightsailKey":    os.Getenv("AWS_ACCESS_KEY_ID"),
		"lightsailSecret": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"lightsailDomain": os.Getenv("DNS_ZONE"),
	}
	for _, v := range m {
		if v == "" {
			return nil, fmt.Errorf("AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and R53_DOMAIN are needed to run this test")
		}
	}
	return m, nil
}
