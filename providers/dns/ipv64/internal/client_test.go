package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMainDomain(t *testing.T) {
	testCases := []struct {
		desc           string
		domain         string
		prefix         string
		expectedDomain string
		errored        bool
	}{
		{
			desc:           "empty",
			domain:         "",
			expectedDomain: "",
			errored:        true,
		},
		{
			desc:           "missing sub domain",
			domain:         "home64.de",
			prefix:         "",
			expectedDomain: "",
			errored:        true,
		},
		{
			desc:           "explicit domain: sub domain",
			domain:         "_acme-challenge.sub.home64.de",
			prefix:         "_acme-challenge",
			expectedDomain: "sub.home64.de",
			errored:        false,
		},
		{
			desc:           "explicit domain: subsub domain",
			domain:         "_acme-challenge.my.sub.home64.de",
			prefix:         "_acme-challenge.my",
			expectedDomain: "sub.home64.de",
			errored:        false,
		},
		{
			desc:           "explicit domain: subsubsub domain",
			domain:         "_acme-challenge.my.sub.sub.home64.de",
			prefix:         "_acme-challenge.my.sub",
			expectedDomain: "sub.home64.de",
			errored:        false,
		},
		{
			desc:           "only subname: sub domain",
			domain:         "_acme-challenge.sub",
			expectedDomain: "",
			prefix:         "",
			errored:        true,
		},
		{
			desc:           "only subname: subsub domain",
			domain:         "_acme-challenge.my.sub",
			expectedDomain: "_acme-challenge.my.sub",
			prefix:         "",
			errored:        false,
		},
		{
			desc:           "only subname: subsubsub domain",
			domain:         "_acme-challenge.my.sub.sub",
			expectedDomain: "my.sub.sub",
			prefix:         "_acme-challenge",
			errored:        false,
		},
		{
			desc:           "only subname: subsubsub domain with only dots",
			domain:         "_acme-challenge...net",
			expectedDomain: "..net",
			prefix:         "_acme-challenge",
			errored:        true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			prefix, wDomain, err := getPrefix(test.domain)

			assert.Equal(t, test.prefix, prefix)
			assert.Equal(t, test.errored, err != nil)
			assert.Equal(t, test.expectedDomain, wDomain)
		})
	}
}
