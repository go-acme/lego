package challenge

type NetworkStack int

const (
	DualStack NetworkStack = iota
	IPv4Only
	IPv6Only
)

func (s NetworkStack) Network(proto string) string {
	switch s {
	case IPv4Only:
		return proto + "4"
	case IPv6Only:
		return proto + "6"
	default:
		return proto
	}
}
