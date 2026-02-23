package dnspersist01

import (
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
	Params           map[string]string
}

// BuildIssueValues constructs an issue-value string for dns-persist-01 and
// optionally includes the persistUntil parameter.
func BuildIssueValues(issuerDomainName, accountURI string, wildcard bool, persistUntil *time.Time) string {
	parts := []string{issuerDomainName}

	if accountURI != "" {
		parts = append(parts, fmt.Sprintf("%s=%s", paramAccountURI, accountURI))
	}

	if wildcard {
		parts = append(parts, fmt.Sprintf("%s=%s", paramPolicy, policyWildcard))
	}

	if persistUntil != nil {
		parts = append(parts, fmt.Sprintf("persistUntil=%d", persistUntil.UTC().Unix()))
	}

	return strings.Join(parts, "; ")
}

func trimWSP(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
}

// ParseIssueValues parses an issue-value string. Unknown parameters are
// preserved in Params.
func ParseIssueValues(value string) (IssueValue, error) {
	fields := strings.Split(value, ";")

	issuerDomainName := trimWSP(fields[0])
	if issuerDomainName == "" {
		return IssueValue{}, fmt.Errorf("missing issuer-domain-name")
	}

	parsed := IssueValue{
		IssuerDomainName: issuerDomainName,
		Params:           map[string]string{},
	}

	seenTags := map[string]bool{}

	for _, raw := range fields[1:] {
		part := trimWSP(raw)
		if part == "" {
			return IssueValue{}, fmt.Errorf("empty parameter or trailing semicolon provided")
		}

		tagValue := strings.SplitN(part, "=", 2)
		if len(tagValue) != 2 {
			return IssueValue{}, fmt.Errorf("malformed parameter %q should be tag=value pair", part)
		}

		tag := trimWSP(tagValue[0])
		val := trimWSP(tagValue[1])
		if tag == "" {
			return IssueValue{}, fmt.Errorf("malformed parameter %q, empty tag", part)
		}

		key := strings.ToLower(tag)
		if seenTags[key] {
			return IssueValue{}, fmt.Errorf("duplicate parameter %q", tag)
		}
		seenTags[key] = true

		for _, r := range val {
			if (r >= 0x21 && r <= 0x3A) || (r >= 0x3C && r <= 0x7E) {
				continue
			}

			return IssueValue{}, fmt.Errorf("malformed value %q for tag %q", val, tag)
		}

		switch key {
		case paramAccountURI:
			if val == "" {
				return IssueValue{}, fmt.Errorf("empty value provided for mandatory accounturi")
			}
			parsed.AccountURI = val
		case paramPolicy:
			if val != "" && strings.ToLower(val) != policyWildcard {
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
			parsed.Params[key] = val
		}
	}

	return parsed, nil
}
