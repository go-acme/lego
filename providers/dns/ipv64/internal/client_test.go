package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMainDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "empty",
			domain:   "",
			expected: "",
		},
		{
			desc:     "missing sub domain",
			domain:   "home64.de",
			expected: "",
		},
		{
			desc:     "explicit domain: sub domain",
			domain:   "_acme-challenge.sub.home64.de",
			expected: "sub.home64.de",
		},
		{
			desc:     "explicit domain: subsub domain",
			domain:   "_acme-challenge.my.sub.home64.de",
			expected: "sub.home64.de",
		},
		{
			desc:     "explicit domain: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub.home64.de",
			expected: "sub.home64.de",
		},
		{
			desc:     "only subname: sub domain",
			domain:   "_acme-challenge.sub",
			expected: "sub",
		},
		{
			desc:     "only subname: subsub domain",
			domain:   "_acme-challenge.my.sub",
			expected: "sub",
		},
		{
			desc:     "only subname: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub",
			expected: "sub",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			wDomain := getPrefix(test.domain)
			assert.Equal(t, test.expected, wDomain)
		})
	}
}
