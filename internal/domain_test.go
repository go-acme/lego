package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPunycodeDomains(t *testing.T) {
	testCases := []struct {
		desc     string
		domains  []string
		expected []string
	}{
		{
			desc: "empty slice",
		},
		{
			desc:     "not IDN domains",
			domains:  []string{"example.com", "example.org"},
			expected: []string{"example.com", "example.org"},
		},
		{
			desc: "IDN domains, already encoded",
			// https://www.iana.org/domains/reserved
			domains:  []string{"xn--9t4b11yi5a", "xn--zckzah", "xn--jxalpdlp"},
			expected: []string{"xn--9t4b11yi5a", "xn--zckzah", "xn--jxalpdlp"},
		},
		{
			desc: "IDN domains, to encode",
			// https://www.iana.org/domains/reserved
			domains:  []string{"테스트", "テスト", "δοκιμή"},
			expected: []string{"xn--9t4b11yi5a", "xn--zckzah", "xn--jxalpdlp"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := PunycodeDomains(test.domains)

			assert.Equal(t, test.expected, actual)
		})
	}
}
