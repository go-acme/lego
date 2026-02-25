package dnspersist01

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIssueValue(t *testing.T) {
	testCases := []struct {
		desc              string
		issuer            string
		accountURI        string
		wildcard          bool
		persistUTC        time.Time
		expect            string
		expectErrContains string
	}{
		{
			desc:       "basic",
			issuer:     "authority.example",
			accountURI: "https://authority.example/acct/123",
			expect:     "authority.example; accounturi=https://authority.example/acct/123",
		},
		{
			desc:       "with persistUntil",
			issuer:     "authority.example",
			accountURI: "https://authority.example/acct/123",
			wildcard:   true,
			persistUTC: time.Unix(4102444800, 0).UTC(),
			expect:     "authority.example; accounturi=https://authority.example/acct/123; policy=wildcard; persistUntil=4102444800",
		},
		{
			desc:              "missing account uri",
			issuer:            "authority.example",
			expectErrContains: "ACME account URI cannot be empty",
		},
		{
			desc:              "invalid issuer",
			issuer:            "Authority.Example.",
			accountURI:        "https://authority.example/acct/123",
			expectErrContains: "issuer-domain-name must be lowercase",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := BuildIssueValue(test.issuer, test.accountURI, test.wildcard, test.persistUTC)
			if test.expectErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectErrContains)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expect, actual)
		})
	}
}

func TestParseIssueValue(t *testing.T) {
	testCases := []struct {
		desc              string
		value             string
		expected          IssueValue
		expectErrContains string
	}{
		{
			desc:  "basic",
			value: "authority.example; accounturi=https://authority.example/acct/123",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
			},
		},
		{
			desc:  "wildcard policy is case-insensitive",
			value: "authority.example; accounturi=https://authority.example/acct/123; policy=wIlDcArD",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Policy:           "wIlDcArD",
			},
		},
		{
			desc:  "unknown param",
			value: "authority.example; accounturi=https://authority.example/acct/123; extra=value",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
			},
		},
		{
			desc:  "unknown tag with empty value",
			value: "authority.example; accounturi=https://authority.example/acct/123; foo=",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
			},
		},
		{
			desc:  "unknown tags with unusual formatting are ignored",
			value: "authority.example;accounturi=https://authority.example/acct/123;bad tag=value;\nweird=\\x01337",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
			},
		},
		{
			desc:  "all known fields with heavy whitespace",
			value: "   authority.example   ;   accounturi   =   https://authority.example/acct/123   ;   policy   =   wildcard   ;   persistUntil   =   4102444800   ",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Policy:           "wildcard",
				PersistUntil:     time.Unix(4102444800, 0).UTC(),
			},
		},
		{
			desc:  "policy other than wildcard is treated as absent",
			value: "authority.example; accounturi=https://authority.example/acct/123; policy=notwildcard",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
			},
		},
		{
			desc:  "missing accounturi",
			value: "authority.example",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
			},
		},
		{
			desc:              "missing issuer",
			value:             "; accounturi=https://authority.example/acct/123",
			expectErrContains: "missing issuer-domain-name",
		},
		{
			desc:              "invalid parameter",
			value:             "authority.example; badparam",
			expectErrContains: `malformed parameter "badparam" should be tag=value pair`,
		},
		{
			desc:              "empty tag is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; =abc",
			expectErrContains: `malformed parameter "=abc", empty tag`,
		},
		{
			desc:              "empty accounturi is malformed",
			value:             "authority.example; accounturi=",
			expectErrContains: `empty value provided for mandatory "accounturi"`,
		},
		{
			desc:              "invalid value character is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; policy=wild card",
			expectErrContains: `malformed value "wild card" for tag "policy"`,
		},
		{
			desc:              "persistUntil non unix timestamp is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; persistUntil=not-a-unix-timestamp",
			expectErrContains: `malformed "persistuntil": strconv.ParseInt: parsing "not-a-unix-timestamp": invalid syntax`,
		},
		{
			desc:              "duplicate unknown parameter is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; foo=bar; foo=baz",
			expectErrContains: `duplicate parameter "foo"`,
		},
		{
			desc:              "duplicate parameter is case-insensitive",
			value:             "authority.example; ACCOUNTURI=https://authority.example/acct/123; accounturi=https://authority.example/acct/456",
			expectErrContains: `duplicate parameter "accounturi"`,
		},
		{
			desc:              "trailing semicolon is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123;",
			expectErrContains: "empty parameter or trailing semicolon provided",
		},
		{
			desc:              "empty persistUntil is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; persistUntil=",
			expectErrContains: `malformed "persistuntil": strconv.ParseInt: parsing "": invalid syntax`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseIssueValue(test.value)
			if test.expectErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.expectErrContains)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expected, parsed)
		})
	}
}
