package edgeone

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	teo "github.com/go-acme/tencentedgdeone/v20220901"
)

func (d *DNSProvider) getHostedZoneID(ctx context.Context, domain string) (*string, error) {
	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	if d.config.ZonesMapping != nil {
		zoneID, ok := d.config.ZonesMapping[authZone]
		if ok {
			return ptr.Pointer(zoneID), nil
		}
	}

	request := teo.NewDescribeZonesRequest()

	var zones []*teo.Zone

	for {
		response, err := teo.DescribeZonesWithContext(ctx, d.client, request)
		if err != nil {
			return nil, fmt.Errorf("API call failed: %w", err)
		}

		zones = append(zones, response.Response.Zones...)

		if int64(len(zones)) >= ptr.Deref(response.Response.TotalCount) {
			break
		}

		request.Offset = ptr.Pointer(int64(len(zones)))
	}

	var hostedZone *teo.Zone

	for _, zone := range zones {
		unfqdn := dns01.UnFqdn(authZone)
		if ptr.Deref(zone.ZoneName) == unfqdn {
			hostedZone = zone
		}
	}

	if hostedZone == nil {
		return nil, fmt.Errorf("zone %s not found for domain %s", authZone, domain)
	}

	return hostedZone.ZoneId, nil
}
