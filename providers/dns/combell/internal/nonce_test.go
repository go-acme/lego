package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonce_Generate(t *testing.T) {
	nonce := NewNonce()

	testCases := []struct {
		desc   string
		length int
	}{
		{
			desc:   "one letter",
			length: 1,
		},
		{
			desc:   "10 letters",
			length: 10,
		},
		{
			desc:   "100 letters",
			length: 100,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := nonce.Generate(test.length)
			assert.Len(t, result, test.length)
			assert.Regexp(t, `[a-zA-Z]+`, result)
		})
	}
}
