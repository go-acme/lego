package dnspersist01

import (
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

// ClearFqdnCache clears the cache of fqdn to zone mappings. Primarily used in testing.
func (c *Client) ClearFqdnCache() {
	c.core.ClearFqdnCache()
}
