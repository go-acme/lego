package challenge

type NetworkStack int

const (
	dualStack NetworkStack = iota
	ipv4only
	ipv6only
)

func (s NetworkStack) Network(proto string) string {
	switch s {
	case ipv4only:
		return proto + "4"
	case ipv6only:
		return proto + "6"
	default:
		return proto
	}
}
