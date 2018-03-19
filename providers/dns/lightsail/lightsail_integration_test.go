package lightsail

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
)

func TestLightsailTTL(t *testing.T) {

	m, err := testGetAndPreCheck()
	if err != nil {
		t.Skip(err.Error())
	}

	provider, err := NewDNSProvider()
	if err != nil {
		t.Fatalf("Fatal: %s", err.Error())
	}

	err = provider.Present(m["lightsailDomain"], "foo", "bar")
	if err != nil {
		t.Fatalf("Fatal: %s", err.Error())
	}
	// we need a separate Lightshail client here as the one in the DNS provider is
	// unexported.
	fqdn := "_acme-challenge." + m["lightsailDomain"]
	svc := lightsail.New(session.New())
	if err != nil {
		provider.CleanUp(m["lightsailDomain"], "foo", "bar")
		t.Fatalf("Fatal: %s", err.Error())
	}
	params := &lightsail.GetDomainInput{
		DomainName: aws.String(m["lightsailDomain"]),
	}
	resp, err := svc.GetDomain(params)
	if err != nil {
		provider.CleanUp(m["lightsailDomain"], "foo", "bar")
		t.Fatalf("Fatal: %s", err.Error())
	}
	entries := resp.Domain.DomainEntries
	for _, entry := range entries {
		if *entry.Type == "TXT" && *entry.Name == fqdn {
			provider.CleanUp(m["lightsailDomain"], "foo", "bar")
			return
		}
	}
	provider.CleanUp(m["lightsailDomain"], "foo", "bar")
	t.Fatalf("Could not find a TXT record for _acme-challenge.%s", m["lightsailDomain"])
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
