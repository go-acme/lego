package dnspersist01

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/idna"
)

//nolint:gochecknoglobals // test seam for injecting IDNA conversion failures/variants.
var issuerDomainNameToASCII = idna.Lookup.ToASCII

// validateIssuerDomainName validates a single issuer-domain-name according to
// the following rules:
//   - lowercase only
//   - no trailing dot
//   - max 253 octets overall
//   - non-empty labels, each max 63 octets
//   - lowercase LDH label syntax
//   - A-label (Punycode, RFC5890)
func validateIssuerDomainName(name string) error {
	if name == "" {
		return errors.New("issuer-domain-name cannot be empty")
	}

	if strings.ToLower(name) != name {
		return errors.New("issuer-domain-name must be lowercase")
	}

	if strings.HasSuffix(name, ".") {
		return errors.New("issuer-domain-name must not have a trailing dot")
	}

	if len(name) > 253 {
		return errors.New("issuer-domain-name exceeds maximum length of 253 octets")
	}

	labels := strings.SplitSeq(name, ".")
	for label := range labels {
		if label == "" {
			return errors.New("issuer-domain-name contains an empty label")
		}

		if len(label) > 63 {
			return errors.New("issuer-domain-name label exceeds maximum length of 63 octets")
		}

		if !isLDHLabel(label) {
			return fmt.Errorf("issuer-domain-name label %q must be a lowercase LDH label", label)
		}
	}

	ascii, err := issuerDomainNameToASCII(name)
	if err != nil {
		return fmt.Errorf("issuer-domain-name must be represented in A-label format: %w", err)
	}

	if ascii != name {
		return errors.New("issuer-domain-name must be represented in A-label format")
	}

	return nil
}

func isLDHLabel(label string) bool {
	if label == "" {
		return false
	}

	if !isLowerAlphaNum(label[0]) || !isLowerAlphaNum(label[len(label)-1]) {
		return false
	}

	for i := range len(label) {
		c := label[i]
		if isLowerAlphaNum(c) || c == '-' {
			continue
		}

		return false
	}

	return true
}

func isLowerAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// normalizeUserSuppliedIssuerDomainName normalizes a user supplied
// issuer-domain-name for comparison. Note: DO NOT normalize issuer-domain-names
// from the challenge, as they are expected to already be in the correct format.
func normalizeUserSuppliedIssuerDomainName(name string) (string, error) {
	n := strings.TrimSpace(strings.TrimSuffix(name, "."))
	n = strings.ToLower(n)

	ascii, err := idna.Lookup.ToASCII(n)
	if err != nil {
		return "", fmt.Errorf("normalizing supplied issuer-domain-name %q: %w", n, err)
	}

	return ascii, nil
}
