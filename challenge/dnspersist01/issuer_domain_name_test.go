package dnspersist01

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateIssuerDomainName_Errors(t *testing.T) {
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

func TestValidateIssuerDomainName_ErrorNonCanonicalALabel(t *testing.T) {
	mockIssuerDomainNameToASCII(t, func(string) (string, error) {
		return "different.example", nil
	})

	err := validateIssuerDomainName("authority.example")
	require.EqualError(t, err, "issuer-domain-name must be represented in A-label format")
}

func TestValidateIssuerDomainName_Valid(t *testing.T) {
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
