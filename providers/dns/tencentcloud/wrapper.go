package tencentcloud

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"golang.org/x/net/idna"
)

func extractRecordName(fqdn, zone string) (string, error) {
	asciiDomain, err := idna.ToASCII(zone)
	if err != nil {
		return "", fmt.Errorf("fail to convert punycode: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, asciiDomain)
	if err != nil {
		return "", err
	}

	return subDomain, nil
}
