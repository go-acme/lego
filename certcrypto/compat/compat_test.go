package compat

import (
	"encoding/json"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyTypeCompat_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		desc     string
		raw      string
		expected KeyTypeCompat
	}{
		{
			desc:     "RSA8192: new value",
			raw:      `"RSA8192"`,
			expected: RSA8192,
		},
		{
			desc:     "RSA8192: compatibility",
			raw:      `"8192"`,
			expected: RSA8192,
		},
		{
			desc:     "RSA4096: new value",
			raw:      `"RSA4096"`,
			expected: RSA4096,
		},
		{
			desc:     "RSA4096: compatibility",
			raw:      `"4096"`,
			expected: RSA4096,
		},
		{
			desc:     "RSA3072: new value",
			raw:      `"RSA3072"`,
			expected: RSA3072,
		},
		{
			desc:     "RSA3072: compatibility",
			raw:      `"3072"`,
			expected: RSA3072,
		},
		{
			desc:     "RSA2048: new value",
			raw:      `"RSA2048"`,
			expected: RSA2048,
		},
		{
			desc:     "RSA2048: compatibility",
			raw:      `"2048"`,
			expected: RSA2048,
		},
		{
			desc:     "EC384: new value",
			raw:      `"EC384"`,
			expected: EC384,
		},
		{
			desc:     "EC384: compatibility",
			raw:      `"P384"`,
			expected: EC384,
		},
		{
			desc:     "EC256: new value",
			raw:      `"EC256"`,
			expected: EC256,
		},
		{
			desc:     "EC256: compatibility",
			raw:      `"P256"`,
			expected: EC256,
		},
		{
			desc:     "RSA4096: new value (compat type)",
			raw:      `"RSA4096"`,
			expected: KeyTypeCompat(certcrypto.RSA4096),
		},
		{
			desc:     "RSA4096: compatibility (compat type)",
			raw:      `"4096"`,
			expected: KeyTypeCompat(certcrypto.RSA4096),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var r KeyTypeCompat

			err := json.Unmarshal([]byte(test.raw), &r)
			require.NoError(t, err)

			assert.Equal(t, test.expected, r)
		})
	}
}
