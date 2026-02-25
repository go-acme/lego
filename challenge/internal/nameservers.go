package internal

import (
	"net"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/miekg/dns"
)

const DefaultResolvConf = "/etc/resolv.conf"

// GetNameservers attempts to get systems nameservers before falling back to the defaults.
func GetNameservers(path string, stack challenge.NetworkStack) []string {
	config, err := dns.ClientConfigFromFile(path)
	if err == nil && len(config.Servers) > 0 {
		return config.Servers
	}

	switch stack {
	case challenge.IPv4Only:
		return []string{
			"1.1.1.1:53",
			"1.0.0.1:53",
		}

	case challenge.IPv6Only:
		return []string{
			"[2606:4700:4700::1111]:53",
			"[2606:4700:4700::1001]:53",
		}

	default:
		return []string{
			"1.1.1.1:53",
			"1.0.0.1:53",
			"[2606:4700:4700::1111]:53",
			"[2606:4700:4700::1001]:53",
		}
	}
}

func ParseNameservers(servers []string) []string {
	var resolvers []string

	for _, resolver := range servers {
		// ensure all servers have a port number
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolvers = append(resolvers, net.JoinHostPort(resolver, "53"))
		} else {
			resolvers = append(resolvers, resolver)
		}
	}

	return resolvers
}
