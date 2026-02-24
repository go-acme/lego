package dnspersist01

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	policyWildcard    = "wildcard"
	paramAccountURI   = "accounturi"
	paramPolicy       = "policy"
	paramPersistUntil = "persistuntil"
)

// IssueValue represents a parsed dns-persist-01 issue-value.
type IssueValue struct {
	IssuerDomainName string
	AccountURI       string
	Policy           string
	PersistUntil     *time.Time
}

// BuildIssueValue constructs an RFC 8659 issue-value for a dns-persist-01 TXT
// record. issuerDomainName and accountURI are required. wildcard and
// persistUntil are optional.
func BuildIssueValue(issuerDomainName, accountURI string, wildcard bool, persistUntil *time.Time) (string, error) {
	if accountURI == "" {
		return "", errors.New("dnspersist01: ACME account URI cannot be empty")
	}

	err := validateIssuerDomainName(issuerDomainName)
	if err != nil {
		return "", fmt.Errorf("dnspersist01: %w", err)
	}

	value := issuerDomainName + "; " + paramAccountURI + "=" + accountURI

	if wildcard {
		value += "; " + paramPolicy + "=" + policyWildcard
	}

	if persistUntil != nil {
		value += fmt.Sprintf("; persistUntil=%d", persistUntil.UTC().Unix())
	}

	return value, nil
}

// trimWSP trims RFC 5234 WSP (SP / HTAB) characters from both ends of a
// string, as referenced by RFC 8659.
func trimWSP(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
}

// ParseIssueValue parses the issuer-domain-name and parameters for an RFC
// 8659 issue-value TXT record and returns the extracted fields. It returns
// an error if any portion of the value is malformed.
//
//nolint:gocyclo // parsing and validating tagged parameters requires branching
func ParseIssueValue(value string) (IssueValue, error) {
	fields := strings.Split(value, ";")

	issuerDomainName := trimWSP(fields[0])
	if issuerDomainName == "" {
		return IssueValue{}, errors.New("missing issuer-domain-name")
	}

	parsed := IssueValue{
		IssuerDomainName: issuerDomainName,
	}

	// Parse parameters (with optional surrounding WSP).
	seenTags := map[string]bool{}

	for _, raw := range fields[1:] {
		part := trimWSP(raw)
		if part == "" {
			return IssueValue{}, errors.New("empty parameter or trailing semicolon provided")
		}

		// Capture each tag=value pair.
		tagValue := strings.SplitN(part, "=", 2)
		if len(tagValue) != 2 {
			return IssueValue{}, fmt.Errorf("malformed parameter %q should be tag=value pair", part)
		}

		tag := trimWSP(tagValue[0])
		val := trimWSP(tagValue[1])

		if tag == "" {
			return IssueValue{}, fmt.Errorf("malformed parameter %q, empty tag", part)
		}

		canonicalTag := strings.ToLower(tag)
		if seenTags[canonicalTag] {
			return IssueValue{}, fmt.Errorf("duplicate parameter %q", tag)
		}

		seenTags[canonicalTag] = true
		// Ensure values contain no whitespace/control/non-ASCII characters.
		for _, r := range val {
			if (r >= 0x21 && r <= 0x3A) || (r >= 0x3C && r <= 0x7E) {
				continue
			}

			return IssueValue{}, fmt.Errorf("malformed value %q for tag %q", val, tag)
		}

		// Finally, capture expected tag values.
		//
		// Note: according to RFC 8659 matching of tags is case insensitive.
		switch canonicalTag {
		case paramAccountURI:
			if val == "" {
				return IssueValue{}, errors.New("empty value provided for mandatory accounturi")
			}

			parsed.AccountURI = val
		case paramPolicy:
			// Per the dns-persist-01 specification, if the policy tag is
			// present parameter's tag and defined values MUST be treated as
			// case-insensitive.
			if val != "" && !strings.EqualFold(val, policyWildcard) {
				// If the policy parameter's value is anything other than
				// "wildcard", the a CA MUST proceed as if the policy parameter
				// were not present.
				val = ""
			}

			parsed.Policy = val
		case paramPersistUntil:
			ts, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return IssueValue{}, fmt.Errorf("malformed persistUntil timestamp %q", val)
			}

			persistUntil := time.Unix(ts, 0).UTC()
			parsed.PersistUntil = &persistUntil
		default:
			// Unknown parameters are permitted but not currently consumed.
		}
	}

	return parsed, nil
}
