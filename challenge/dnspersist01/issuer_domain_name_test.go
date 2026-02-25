package dnspersist01

import (
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateIssuerDomainNames(t *testing.T) {
	testCases := []struct {
		desc    string
		issuers []string
		assert  assert.ErrorAssertionFunc
	}{
		{
			desc:   "missing issuers",
			assert: assert.Error,
		},
		{
			desc:    "too many issuers",
			issuers: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			assert:  assert.Error,
		},
		{
			desc:    "valid issuer",
			issuers: []string{"ca.example"},
			assert:  assert.NoError,
		},
		{
			desc:    "issuer all uppercase",
			issuers: []string{"CA.EXAMPLE"},
			assert:  assert.Error,
		},
		{
			desc:    "issuer contains underscore",
			issuers: []string{"ca_.example"},
			assert:  assert.Error,
		},
		{
			desc:    "issuer not in A-label format",
			issuers: []string{"bücher.example"},
			assert:  assert.Error,
		},
		{
			desc:    "issuer too long",
			issuers: []string{strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 63)},
			assert:  assert.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := validateIssuerDomainNames(acme.Challenge{IssuerDomainNames: test.issuers})
			test.assert(t, err)
		})
	}
}

func Test_validateIssuerDomainName_errors(t *testing.T) {
	testCases := []struct {
		desc      string
		name      string
		expectErr string
	}{
		{
			desc:      "trailing dot",
			name:      "authority.example.",
			expectErr: "issuer-domain-name must not have a trailing dot",
		},
		{
			desc:      "empty label",
			name:      "authority..example",
			expectErr: "issuer-domain-name contains an empty label",
		},
		{
			desc:      "label too long",
			name:      strings.Repeat("a", 64) + ".example",
			expectErr: "issuer-domain-name label exceeds the maximum length of 63 octets",
		},
		{
			desc:      "invalid a-label with idna error",
			name:      "xn--a.example",
			expectErr: `issuer-domain-name must be represented in A-label format: idna: invalid label "\u0080"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := validateIssuerDomainName(test.name)
			require.EqualError(t, err, test.expectErr)
		})
	}
}

func Test_validateIssuerDomainName_errorNonCanonicalALabel(t *testing.T) {
	mockIssuerDomainNameToASCII(t, func(string) (string, error) {
		return "different.example", nil
	})

	err := validateIssuerDomainName("authority.example")
	require.EqualError(t, err, "issuer-domain-name must be represented in A-label format")
}

func Test_validateIssuerDomainName_Valid(t *testing.T) {
	mockIssuerDomainNameToASCII(t, func(name string) (string, error) {
		return name, nil
	})

	err := validateIssuerDomainName("authority.example")
	require.NoError(t, err)
}

func mockIssuerDomainNameToASCII(t *testing.T, fn func(string) (string, error)) {
	t.Helper()

	originalToASCII := issuerDomainNameToASCII

	t.Cleanup(func() {
		issuerDomainNameToASCII = originalToASCII
	})

	issuerDomainNameToASCII = fn
}
