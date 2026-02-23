package dnspersist01

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			expectErr: "issuer-domain-name label exceeds maximum length of 63 octets",
		},
		{
			desc:      "invalid a-label with idna error",
			name:      "xn--a.example",
			expectErr: "issuer-domain-name must be represented in A-label format:",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			err := validateIssuerDomainName(test.name)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectErr)
		})
	}
}

func TestValidateIssuerDomainName_ErrorNonCanonicalALabel(t *testing.T) {
	originalToASCII := issuerDomainNameToASCII
	t.Cleanup(func() {
		issuerDomainNameToASCII = originalToASCII
	})

	issuerDomainNameToASCII = func(string) (string, error) {
		return "different.example", nil
	}

	err := validateIssuerDomainName("authority.example")
	require.Error(t, err)
	assert.EqualError(t, err, "issuer-domain-name must be represented in A-label format")
}

func TestValidateIssuerDomainName_Valid(t *testing.T) {
	originalToASCII := issuerDomainNameToASCII
	t.Cleanup(func() {
		issuerDomainNameToASCII = originalToASCII
	})

	issuerDomainNameToASCII = func(name string) (string, error) {
		return name, nil
	}

	err := validateIssuerDomainName("authority.example")
	require.NoError(t, err)
}

func TestValidateIssuerDomainName_ErrorWrap(t *testing.T) {
	originalToASCII := issuerDomainNameToASCII
	t.Cleanup(func() {
		issuerDomainNameToASCII = originalToASCII
	})

	sentinelErr := errors.New("sentinel idna failure")
	issuerDomainNameToASCII = func(string) (string, error) {
		return "", sentinelErr
	}

	err := validateIssuerDomainName("authority.example")
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinelErr)
	assert.Contains(t, err.Error(), "issuer-domain-name must be represented in A-label format:")
}
