package dns01

import (
	"context"
	"sync/atomic"

	"github.com/go-acme/lego/v5/challenge/internal"
)

var defaultClient atomic.Pointer[Client]

func init() {
	defaultClient.Store(NewClient(nil))
}

func DefaultClient() *Client { return defaultClient.Load() }

func SetDefaultClient(c *Client) {
	defaultClient.Store(c)
}

type Options = internal.Options

type Client struct {
	core *internal.Client

	// authoritativeNSPort used by authoritative NS.
	// For testing purposes only.
	authoritativeNSPort string
}

func NewClient(opts *Options) *Client {
	return &Client{
		core: internal.NewClient(opts),

		authoritativeNSPort: "53",
	}
}

// FindZoneByFqdn determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func (c *Client) FindZoneByFqdn(ctx context.Context, fqdn string) (string, error) {
	return c.core.FindZoneByFqdn(ctx, fqdn)
}

// FindZoneByFqdnCustom determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func (c *Client) FindZoneByFqdnCustom(ctx context.Context, fqdn string, nameservers []string) (string, error) {
	return c.core.FindZoneByFqdnCustom(ctx, fqdn, nameservers)
}

// ClearFqdnCache clears the cache of fqdn to zone mappings. Primarily used in testing.
func (c *Client) ClearFqdnCache() {
	c.core.ClearFqdnCache()
}
