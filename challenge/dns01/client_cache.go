package dns01

import (
	"time"

	"github.com/miekg/dns"
)

// soaCacheEntry holds a cached SOA record (only selected fields).
type soaCacheEntry struct {
	zone      string    // zone apex (a domain name)
	primaryNs string    // primary nameserver for the zone apex
	expires   time.Time // time when this cache entry should be evicted
}

func newSoaCacheEntry(soa *dns.SOA) *soaCacheEntry {
	return &soaCacheEntry{
		zone:      soa.Hdr.Name,
		primaryNs: soa.Ns,
		expires:   time.Now().Add(time.Duration(soa.Refresh) * time.Second),
	}
}

// isExpired checks whether a cache entry should be considered expired.
func (cache *soaCacheEntry) isExpired() bool {
	return time.Now().After(cache.expires)
}

// ClearFqdnCache clears the cache of fqdn to zone mappings. Primarily used in testing.
func (c *Client) ClearFqdnCache() {
	c.muFqdnSoaCache.Lock()
	c.fqdnSoaCache = map[string]*soaCacheEntry{}
	c.muFqdnSoaCache.Unlock()
}
