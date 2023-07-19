package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMainDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		prefix   string
		expected string
		errored  bool
	}{
		{
			desc:     "empty",
			domain:   "",
			expected: "",
			errored:  true,
		},
		{
			desc:     "missing sub domain",
			domain:   "home64.de",
			prefix:   "",
			expected: "",
			errored:  true,
		},
		{
			desc:     "explicit domain: sub domain",
			domain:   "_acme-challenge.sub.home64.de",
			prefix:   "_acme-challenge",
			expected: "sub.home64.de",
			errored:  false,
		},
		{
			desc:     "explicit domain: subsub domain",
			domain:   "_acme-challenge.my.sub.home64.de",
			prefix:   "_acme-challenge.my",
			expected: "sub.home64.de",
			errored:  false,
		},
		{
			desc:     "explicit domain: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub.home64.de",
			prefix:   "_acme-challenge.my.sub",
			expected: "sub.home64.de",
			errored:  false,
		},
		{
			desc:     "only subname: sub domain",
			domain:   "_acme-challenge.sub",
			expected: "",
			prefix:   "",
			errored:  true,
		},
		{
			desc:     "only subname: subsub domain",
			domain:   "_acme-challenge.my.sub",
			expected: "sub",
			prefix:   "",
			errored:  true,
		},
		{
			desc:     "only subname: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub",
			expected: "my.sub.sub",
			prefix:   "",
			errored:  false,
		},
		{
			desc:     "only subname: subsubsub domain",
			domain:   "_acme-challenge...net",
			expected: "..net",
			prefix:   "",
			errored:  true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			prefix, wDomain, err := getPrefix(test.domain)

			println(wDomain)
			assert.Equal(t, test.prefix, prefix)
			assert.Equal(t, test.errored, err != nil)
			assert.Equal(t, test.expected, wDomain)
		})
	}
}
