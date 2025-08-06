package edgeone

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	teo "github.com/go-acme/tencentedgdeone/v20220901"
)

func (d *DNSProvider) getHostedZone(ctx context.Context, domain string) (*teo.Zone, error) {
	request := teo.NewDescribeZonesRequest()

	var domains []*teo.Zone

	for {
		response, err := teo.DescribeZonesWithContext(ctx, d.client, request)
		if err != nil {
			return nil, fmt.Errorf("API call failed: %w", err)
		}

		domains = append(domains, response.Response.Zones...)

		if int64(len(domains)) >= ptr.Deref(response.Response.TotalCount) {
			break
		}

		request.Offset = ptr.Pointer(int64(len(domains)))
	}

	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	var hostedZone *teo.Zone
	for _, zone := range domains {
		unfqdn := dns01.UnFqdn(authZone)
		if ptr.Deref(zone.ZoneName) == unfqdn {
			hostedZone = zone
		}
	}

	if hostedZone == nil {
		return nil, fmt.Errorf("zone %s not found in dnspod for domain %s", authZone, domain)
	}

	return hostedZone, nil
}
