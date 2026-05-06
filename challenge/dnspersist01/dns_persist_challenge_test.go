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
		authz            acme.Authorization
		issuerDomainName string
		accountURI       string
		persistUntil     time.Time
		expected         ChallengeInfo
		expectErr        string
	}{
		{
			desc: "basic",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc: "subdomain",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "api.example.com"},
			},
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.api.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc: "wildcard with normalized issuer",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
				Wildcard:   true,
			},
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123; policy=wildcard",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc: "uppercase issuer is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "Authority.Example.",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must be lowercase",
		},
		{
			desc: "unicode issuer is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "bücher.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "must be a lowercase LDH label",
		},
		{
			desc: "issuer with trailing dot is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "authority.example.",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must not have a trailing dot",
		},
		{
			desc: "issuer with empty label is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "authority..example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name contains an empty label",
		},
		{
			desc: "issuer label length over 63 is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: strings.Repeat("a", 64) + ".example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name label exceeds the maximum length of 63 octets",
		},
		{
			desc: "issuer with malformed punycode a-label is rejected",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "xn--a.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "issuer-domain-name must be represented in A-label format:",
		},
		{
			desc: "includes persistUntil",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
				Wildcard:   true,
			},
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			persistUntil:     time.Unix(4102444800, 0).UTC(),
			expected: ChallengeInfo{
				FQDN:             "_validation-persist.example.com.",
				Value:            "authority.example; accounturi=https://ca.example/acct/123; policy=wildcard; persistUntil=4102444800",
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:             "empty domain",
			authz:            acme.Authorization{},
			issuerDomainName: "authority.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "domain cannot be empty",
		},
		{
			desc: "empty account uri",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "authority.example",
			accountURI:       "",
			expectErr:        "ACME account URI cannot be empty",
		},
		{
			desc: "invalid issuer",
			authz: acme.Authorization{
				Identifier: acme.Identifier{Value: "example.com"},
			},
			issuerDomainName: "ca_.example",
			accountURI:       "https://ca.example/acct/123",
			expectErr:        "must be a lowercase LDH label",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := GetChallengeInfo(test.authz, test.issuerDomainName, test.accountURI, test.persistUntil)
			if test.expectErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectErr)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestChallenge_selectIssuerDomainName(t *testing.T) {
	accountURI := "https://authority.example/acct/123"

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
				{Value: mustChallengeValue(t, "ca.example", accountURI, false, time.Time{})},
			},
			expectIssuerDomainName: "ca.example",
		},
		{
			desc: "override still wins over matching existing record",
			issuers: []string{
				"ca.example", "backup.example",
			},
			records: []TXTRecord{
				{Value: mustChallengeValue(t, "ca.example", accountURI, false, time.Time{})},
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
				userSuppliedIssuerDomainName: test.overrideIssuerDomainName,
			}

			issuer, err := chlg.selectIssuerDomainName(test.issuers, test.records, accountURI, test.wildcard)
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

	accountURI := "acc"

	testCases := []struct {
		desc               string
		records            []TXTRecord
		issuer             string
		wildcard           bool
		requiredPersistUTC time.Time
		assert             assert.BoolAssertionFunc
	}{
		{
			desc:    "match basic",
			records: []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, false, time.Time{})}},
			issuer:  "ca.example",
			assert:  assert.True,
		},
		{
			desc:    "issuer mismatch",
			records: []TXTRecord{{Value: mustChallengeValue(t, "other.example", accountURI, false, time.Time{})}},
			issuer:  "ca.example",
			assert:  assert.False,
		},
		{
			desc:    "account mismatch",
			records: []TXTRecord{{Value: mustChallengeValue(t, "ca.example", "other", false, time.Time{})}},
			issuer:  "ca.example",
			assert:  assert.False,
		},
		{
			desc:     "wildcard requires policy",
			records:  []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, false, time.Time{})}},
			issuer:   "ca.example",
			wildcard: true,
			assert:   assert.False,
		},
		{
			desc:     "wildcard match",
			records:  []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, true, time.Time{})}},
			issuer:   "ca.example",
			wildcard: true,
			assert:   assert.True,
		},
		{
			desc:     "policy wildcard allowed for non-wildcard",
			records:  []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, true, time.Time{})}},
			issuer:   "ca.example",
			wildcard: false,
			assert:   assert.True,
		},
		{
			desc: "matching malformed and matching valid record succeeds",
			records: []TXTRecord{
				{Value: "ca.example;accounturi=acc;accounturi=other"},
				{Value: "ca.example;accounturi=acc"},
			},
			issuer: "ca.example",
			assert: assert.True,
		},
		{
			desc:     "wildcard accepts case-insensitive policy value",
			records:  []TXTRecord{{Value: "ca.example;accounturi=acc;policy=wIlDcArD"}},
			issuer:   "ca.example",
			wildcard: true,
			assert:   assert.True,
		},
		{
			desc:     "wildcard policy mismatch is not a match",
			records:  []TXTRecord{{Value: "ca.example;accounturi=acc;policy=notwildcard"}},
			issuer:   "ca.example",
			wildcard: true,
			assert:   assert.False,
		},
		{
			desc:    "persistUntil present without requirement is not a match",
			records: []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, false, expiredPersistUntil)}},
			issuer:  "ca.example",
			assert:  assert.False,
		},
		{
			desc:    "future persistUntil without requirement is not a match",
			records: []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, false, futurePersistUntil)}},
			issuer:  "ca.example",
			assert:  assert.False,
		},
		{
			desc:               "required persistUntil matches",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc;persistUntil=4102444800"}},
			issuer:             "ca.example",
			requiredPersistUTC: time.Unix(4102444800, 0).UTC(),
			assert:             assert.True,
		},
		{
			desc:               "required persistUntil matches even when expired",
			records:            []TXTRecord{{Value: mustChallengeValue(t, "ca.example", accountURI, false, expiredPersistUntil)}},
			issuer:             "ca.example",
			requiredPersistUTC: expiredPersistUntil,
			assert:             assert.True,
		},
		{
			desc:               "required persistUntil mismatch",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc;persistUntil=4102444801"}},
			issuer:             "ca.example",
			requiredPersistUTC: time.Unix(4102444800, 0).UTC(),
			assert:             assert.False,
		},
		{
			desc:               "required persistUntil missing",
			records:            []TXTRecord{{Value: "ca.example;accounturi=acc"}},
			issuer:             "ca.example",
			requiredPersistUTC: time.Unix(4102444800, 0).UTC(),
			assert:             assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			chlg := &Challenge{
				persistUntil: test.requiredPersistUTC,
			}

			match := chlg.hasMatchingRecord(test.records, test.issuer, accountURI, test.wildcard)

			test.assert(t, match)
		})
	}
}

func mustChallengeValue(t *testing.T, issuerDomainName, accountURI string, wildcard bool, persistUntil time.Time) string {
	t.Helper()

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "example.com",
		},
		Wildcard: wildcard,
	}

	info, err := GetChallengeInfo(authz, issuerDomainName, accountURI, persistUntil)
	require.NoError(t, err)

	return info.Value
}
