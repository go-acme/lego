package dns01

// networkStack is used to indicate which IP stack should be used for DNS queries.
type networkStack int

const (
	dualStack networkStack = iota
	ipv4only
	ipv6only
)

// currentNetworkStack is used to define which IP stack will be used. The default is
// both IPv4 and IPv6. Set to IPv4Only or IPv6Only to select either version.
var currentNetworkStack = dualStack

// Network interprets the NetworkStack setting in relation to the desired
// protocol. The proto value should be either "udp" or "tcp".
func (s networkStack) Network(proto string) string {
	// The DNS client passes whatever value is set in (*dns.Client).Net to
	// the [net.Dialer](https://github.com/miekg/dns/blob/fe20d5d/client.go#L119-L141).
	// And the net.Dialer accepts strings such as "udp4" or "tcp6"
	// (https://cs.opensource.google/go/go/+/refs/tags/go1.18.9:src/net/dial.go;l=167-182).
	switch s {
	case ipv4only:
		return proto + "4"
	case ipv6only:
		return proto + "6"
	default:
		return proto
	}
}

// SetIPv4Only forces DNS queries to only happen over the IPv4 stack.
func SetIPv4Only() { currentNetworkStack = ipv4only }

// SetIPv6Only forces DNS queries to only happen over the IPv6 stack.
func SetIPv6Only() { currentNetworkStack = ipv6only }

// SetDualStack indicates that both IPv4 and IPv6 should be allowed.
// This setting lets the OS determine which IP stack to use.
func SetDualStack() { currentNetworkStack = dualStack }
