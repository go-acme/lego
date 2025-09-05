package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_decodeEABHmac(t *testing.T) {
	testCases := []struct {
		desc string
		hmac string
	}{
		{
			desc: "RawURLEncoding",
			hmac: "BAEDAgQCBQcGCAUDDDMBAAIRAwQhEjEFQVFhEyJxgTIGFJGhsUIjJBVSwWIzNHKC0UMHJZJT8OHx",
		},
		{
			desc: "URLEncoding",
			hmac: "nKTo9Hu8fpCqWPXx-25LVbZrJWxcHISsr4qHrRR0j5U=",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			v, err := decodeEABHmac(test.hmac)
			require.NoError(t, err)

			assert.NotEmpty(t, v)
		})
	}
}
