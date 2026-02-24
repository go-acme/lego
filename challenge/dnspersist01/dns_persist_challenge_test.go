package dnspersist01

import (
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChallengeInfo(t *testing.T) {
	testCases := []struct {
		desc             string
		domain           string
		issuerDomainName string
		accountURI       string
		wildcard         bool
		persistUntil     *time.Time
		expected         ChallengeInfo
		expectErr        string
	}{
		{
			desc:             "basic",
			domain:           "example.com",
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:             "subdomain",
			domain:           "api.example.com",
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.api.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:             "wildcard with normalized issuer",
			domain:           "example.com",
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			wildcard:         true,
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123; policy=wildcard",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:             "uppercase issuer is rejected",
			domain:           "example.com",
			issuerDomainName: "Authority.Example.",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must be lowercase",
		},
		{
			desc:             "unicode issuer is rejected",
			domain:           "example.com",
			issuerDomainName: "bücher.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "must be a lowercase LDH label",
		},
		{
			desc:             "issuer with trailing dot is rejected",
			domain:           "example.com",
			issuerDomainName: "authority.example.",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must not have a trailing dot",
		},
		{
			desc:             "issuer with empty label is rejected",
			domain:           "example.com",
			issuerDomainName: "authority..example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name contains an empty label",
		},
		{
			desc:             "issuer label length over 63 is rejected",
			domain:           "example.com",
			issuerDomainName: strings.Repeat("a", 64) + ".example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name label exceeds maximum length of 63 octets",
		},
		{
			desc:             "issuer with malformed punycode a-label is rejected",
			domain:           "example.com",
			issuerDomainName: "xn--a.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must be represented in A-label format:",
		},
		{
			desc:             "includes persistUntil",
			domain:           "example.com",
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			wildcard:         true,
			persistUntil:     Pointer(time.Unix(4102444800, 0).UTC()),
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123; policy=wildcard; persistUntil=4102444800",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:             "empty domain",
			domain:           "",
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "domain cannot be empty",
		},
		{
			desc:             "empty account uri",
			domain:           "example.com",
			issuerDomainName: "authority.example",
			accountURI:       "",
			expectErr:        "ACME account URI cannot be empty",
		},
		{
			desc:             "invalid issuer",
			domain:           "example.com",
			issuerDomainName: "ca_.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "must be a lowercase LDH label",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actual, err := GetChallengeInfo(test.domain, test.issuerDomainName, test.accountURI, test.wildcard, test.persistUntil)
			if test.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestValidateIssuerDomainNames(t *testing.T) {
	testCases := []struct {
		desc      string
		issuers   []string
		expectErr bool
	}{
		{
			desc:      "missing issuers",
			expectErr: true,
		},
		{
			desc:      "too many issuers",
			issuers:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			expectErr: true,
		},
		{
			desc:    "valid issuer",
			issuers: []string{"ca.example"},
		},
		{
			desc:      "issuer all uppercase",
			issuers:   []string{"CA.EXAMPLE"},
			expectErr: true,
		},
		{
			desc:      "issuer contains underscore",
			issuers:   []string{"ca_.example"},
			expectErr: true,
		},
		{
			desc:      "issuer not in A-label format",
			issuers:   []string{"bücher.example"},
			expectErr: true,
		},
		{
			desc:      "issuer too long",
			issuers:   []string{strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 63)},
			expectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			err := validateIssuerDomainNames(acme.Challenge{IssuerDomainNames: test.issuers})
			if test.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestWithIssuerDomainName(t *testing.T) {
	testCases := []struct {
		desc      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			desc:     "normalizes uppercase and trailing dot",
			input:    "CA.EXAMPLE.",
			expected: "ca.example",
		},
		{
			desc:     "normalizes idna issuer",
			input:    "BÜCHER.example",
			expected: "xn--bcher-kva.example",
		},
		{
			desc:      "rejects invalid issuer",
			input:     "ca_.example",
			expectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := &Challenge{}

			err := WithIssuerDomainName(test.input)(chlg)
			if test.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, chlg.userSuppliedIssuerDomainName)
		})
	}
}

func TestChallenge_selectIssuerDomainName(t *testing.T) {
	testCases := []struct {
		desc                     string
		issuers                  []string
		records                  []TXTRecord
		wildcard                 bool
		overrideIssuerDomainName string
		expectIssuerDomainName   string
		expectErr                bool
	}{
		{
			desc:                   "default uses sorted first",
			issuers:                []string{"ca.example", "backup.example"},
			expectIssuerDomainName: "backup.example",
		},
		{
			desc: "default prefers existing matching record",
			issuers: []string{
				"ca.example", "backup.example",
			},
			records: []TXTRecord{
				{Value: BuildIssueValues("ca.example", "https://authority.example/acct/123", false, nil)},
			},
			expectIssuerDomainName: "ca.example",
		},
		{
			desc: "override still wins over matching existing record",
			issuers: []string{
				"ca.example", "backup.example",
			},
			records: []TXTRecord{
				{Value: BuildIssueValues("ca.example", "https://authority.example/acct/123", false, nil)},
			},
			overrideIssuerDomainName: "backup.example",
			expectIssuerDomainName:   "backup.example",
		},
		{
			desc:                     "override not offered in challenge",
			issuers:                  []string{"ca.example"},
			overrideIssuerDomainName: "other.example",
			expectErr:                true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := &Challenge{
				accountURI:                   "https://authority.example/acct/123",
				userSuppliedIssuerDomainName: test.overrideIssuerDomainName,
			}

			issuer, err := chlg.selectIssuerDomainName(test.issuers, test.records, test.wildcard)
			if test.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectIssuerDomainName, issuer)
		})
	}
}

func TestChallenge_hasMatchingRecord(t *testing.T) {
	expiredPersistUntil := time.Unix(1700000000, 0).UTC()
	futurePersistUntil := time.Unix(4102444800, 0).UTC()

	testCases := []struct {
		desc               string
		records            []TXTRecord
		issuer             string
		wildcard           bool
		requiredPersistUTC *time.Time
		expect             bool
	}{
		{
			desc:    "match basic",
			records: []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", false, nil)}},
			issuer:  "ca.example",
			expect:  true,
		},
		{
			desc:    "issuer mismatch",
			records: []TXTRecord{{Value: BuildIssueValues("other.example", "acc", false, nil)}},
			issuer:  "ca.example",
			expect:  false,
		},
		{
			desc:    "account mismatch",
			records: []TXTRecord{{Value: BuildIssueValues("ca.example", "other", false, nil)}},
			issuer:  "ca.example",
			expect:  false,
		},
		{
			desc:     "wildcard requires policy",
			records:  []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", false, nil)}},
			issuer:   "ca.example",
			wildcard: true,
			expect:   false,
		},
		{
			desc:     "wildcard match",
			records:  []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", true, nil)}},
			issuer:   "ca.example",
			wildcard: true,
			expect:   true,
		},
		{
			desc:     "policy wildcard allowed for non-wildcard",
			records:  []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", true, nil)}},
			issuer:   "ca.example",
			wildcard: false,
			expect:   true,
		},
		{
			desc: "matching malformed and matching valid record succeeds",
			records: []TXTRecord{
				{Value: "ca.example;accounturi=acc;accounturi=other"},
				{Value: "ca.example;accounturi=acc"},
			},
			issuer: "ca.example",
			expect: true,
		},
		{
			desc:     "wildcard accepts case-insensitive policy value",
			records:  []TXTRecord{{Value: "ca.example;accounturi=acc;policy=wIlDcArD"}},
			issuer:   "ca.example",
			wildcard: true,
			expect:   true,
		},
		{
			desc:     "wildcard policy mismatch is not a match",
			records:  []TXTRecord{{Value: "ca.example;accounturi=acc;policy=notwildcard"}},
			issuer:   "ca.example",
			wildcard: true,
			expect:   false,
		},
		{
			desc:    "persistUntil present without requirement is not a match",
			records: []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", false, &expiredPersistUntil)}},
			issuer:  "ca.example",
			expect:  false,
		},
		{
			desc:    "future persistUntil without requirement is not a match",
			records: []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", false, &futurePersistUntil)}},
			issuer:  "ca.example",
			expect:  false,
		},
		{
			desc:               "required persistUntil matches",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc;persistUntil=4102444800"}},
			issuer:             "ca.example",
			requiredPersistUTC: Pointer(time.Unix(4102444800, 0).UTC()),
			expect:             true,
		},
		{
			desc:               "required persistUntil matches even when expired",
			records:            []TXTRecord{{Value: BuildIssueValues("ca.example", "acc", false, &expiredPersistUntil)}},
			issuer:             "ca.example",
			requiredPersistUTC: Pointer(expiredPersistUntil),
			expect:             true,
		},
		{
			desc:               "required persistUntil mismatch",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc;persistUntil=4102444801"}},
			issuer:             "ca.example",
			requiredPersistUTC: Pointer(time.Unix(4102444800, 0).UTC()),
			expect:             false,
		},
		{
			desc:               "required persistUntil missing",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc"}},
			issuer:             "ca.example",
			requiredPersistUTC: Pointer(time.Unix(4102444800, 0).UTC()),
			expect:             false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := &Challenge{
				accountURI:   "acc",
				persistUntil: test.requiredPersistUTC,
			}
			actual := chlg.hasMatchingRecord(test.records, test.issuer, test.wildcard)
			assert.Equal(t, test.expect, actual)
		})
	}
}

// TODO(ldez) factorize
func Pointer[T any](v T) *T { return &v }
