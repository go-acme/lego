package dnspersist01

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIssueValues(t *testing.T) {
	testCases := []struct {
		desc       string
		issuer     string
		accountURI string
		wildcard   bool
		persistUTC *time.Time
		expect     string
	}{
		{
			desc:       "basic",
			issuer:     "authority.example",
			accountURI: "https://authority.example/acct/123",
			expect:     "authority.example; accounturi=https://authority.example/acct/123",
		},
		{
			desc:     "no account",
			issuer:   "authority.example",
			wildcard: true,
			expect:   "authority.example; policy=wildcard",
		},
		{
			desc:       "with persistUntil",
			issuer:     "authority.example",
			accountURI: "https://authority.example/acct/123",
			wildcard:   true,
			persistUTC: Pointer(time.Unix(4102444800, 0).UTC()),
			expect:     "authority.example; accounturi=https://authority.example/acct/123; policy=wildcard; persistUntil=4102444800",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actual := BuildIssueValues(test.issuer, test.accountURI, test.wildcard, test.persistUTC)
			assert.Equal(t, test.expect, actual)
		})
	}
}

func TestParseIssueValues(t *testing.T) {
	testCases := []struct {
		desc               string
		value              string
		expected           IssueValue
		expectedPersistUTC *time.Time
		expectErrContains  string
	}{
		{
			desc:  "basic",
			value: "authority.example; accounturi=https://authority.example/acct/123",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Params:           map[string]string{},
			},
		},
		{
			desc:  "wildcard policy is case-insensitive",
			value: "authority.example; accounturi=https://authority.example/acct/123; policy=wIlDcArD",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Policy:           "wIlDcArD",
				Params:           map[string]string{},
			},
		},
		{
			desc:  "unknown param",
			value: "authority.example; accounturi=https://authority.example/acct/123; extra=value",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Params:           map[string]string{"extra": "value"},
			},
		},
		{
			desc:  "unknown tag with empty value",
			value: "authority.example; accounturi=https://authority.example/acct/123; foo=",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Params:           map[string]string{"foo": ""},
			},
		},
		{
			desc:  "unknown tags with unusual formatting are ignored",
			value: "authority.example;accounturi=https://authority.example/acct/123;bad tag=value;\nweird=\\x01337",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Params: map[string]string{
					"bad tag": "value",
					"\nweird": "\\x01337",
				},
			},
		},
		{
			desc:  "all known fields with heavy whitespace",
			value: "   authority.example   ;   accounturi   =   https://authority.example/acct/123   ;   policy   =   wildcard   ;   persistUntil   =   4102444800   ",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Policy:           "wildcard",
				Params:           map[string]string{},
			},
			expectedPersistUTC: Pointer(time.Unix(4102444800, 0).UTC()),
		},
		{
			desc:  "policy other than wildcard is treated as absent",
			value: "authority.example; accounturi=https://authority.example/acct/123; policy=notwildcard",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				AccountURI:       "https://authority.example/acct/123",
				Params:           map[string]string{},
			},
		},
		{
			desc:  "missing accounturi",
			value: "authority.example",
			expected: IssueValue{
				IssuerDomainName: "authority.example",
				Params:           map[string]string{},
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
			expectErrContains: "empty value provided for mandatory accounturi",
		},
		{
			desc:              "invalid value character is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; policy=wild card",
			expectErrContains: `malformed value "wild card" for tag "policy"`,
		},
		{
			desc:              "persistUntil non unix timestamp is malformed",
			value:             "authority.example; accounturi=https://authority.example/acct/123; persistUntil=not-a-unix-timestamp",
			expectErrContains: `malformed persistUntil timestamp "not-a-unix-timestamp"`,
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
			expectErrContains: `malformed persistUntil timestamp`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			parsed, err := ParseIssueValues(test.value)
			if test.expectErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectErrContains)

				return
			}

			require.NoError(t, err)

			expected := test.expected
			expected.PersistUntil = test.expectedPersistUTC
			assert.Equal(t, expected, parsed)
		})
	}
}
