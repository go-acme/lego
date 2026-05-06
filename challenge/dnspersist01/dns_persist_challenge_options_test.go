package dnspersist01

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			t.Parallel()

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
