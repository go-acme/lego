package challenge

// Type is a string that identifies a particular challenge type and version of ACME challenge.
type Type string

const (
	// HTTP01 is the "http-01" ACME challenge https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-8.3
	// Note: ChallengePath returns the URL path to fulfill this challenge
	HTTP01 = Type("http-01")

	// DNS01 is the "dns-01" ACME challenge https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-8.4
	// Note: GetRecord returns a DNS record which will fulfill this challenge
	DNS01 = Type("dns-01")

	// TLSALPN01 is the "tls-alpn-01" ACME challenge https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-05
	TLSALPN01 = Type("tls-alpn-01")
)
