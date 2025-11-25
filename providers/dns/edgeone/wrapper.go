package edgeone

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	teo "github.com/go-acme/tencentedgdeone/v20220901"
)

func (d *DNSProvider) getHostedZoneID(ctx context.Context, domain string) (string, error) {
	if d.config.ZoneID != "" {
		return d.config.ZoneID, nil
	}

	request := teo.NewDescribeZonesRequest()

	var hostedZones []*teo.Zone

	for {
		response, err := teo.DescribeZonesWithContext(ctx, d.client, request)
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}

		hostedZones = append(hostedZones, response.Response.Zones...)

		if int64(len(hostedZones)) >= ptr.Deref(response.Response.TotalCount) {
			break
		}

		request.Offset = ptr.Pointer(int64(len(hostedZones)))
	}

	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	var hostedZone *teo.Zone

	for _, zone := range hostedZones {
		unfqdn := dns01.UnFqdn(authZone)
		if ptr.Deref(zone.ZoneName) == unfqdn {
			hostedZone = zone
			break
		}
	}

	if hostedZone == nil {
		return "", fmt.Errorf("zone %s not found in edgeone for domain %s", authZone, domain)
	}

	return ptr.Deref(hostedZone.ZoneId), nil
}
