package internal

import (
	"github.com/go-acme/lego/v5/log"
	"golang.org/x/net/idna"
)

// PunycodeDomains converts domains to their Punycode encoded versions.
//
// https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.4
// The domain name MUST be encoded in the form in which it would appear in a certificate.
// That is, it MUST be encoded according to the rules in Section 7 of [RFC5280].
//
// https://www.rfc-editor.org/rfc/rfc5280.html#section-7
func PunycodeDomains(domains []string) []string {
	var punycodedDomains []string

	for _, domain := range domains {
		punycodedDomain, err := idna.ToASCII(domain)
		if err != nil {
			log.Warn("skip domain: unable to Punycode encode.",
				log.DomainAttr(domain),
				log.ErrorAttr(err),
			)
		} else {
			punycodedDomains = append(punycodedDomains, punycodedDomain)
		}
	}

	return punycodedDomains
}
