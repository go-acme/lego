package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_merge(t *testing.T) {
	testCases := []struct {
		desc        string
		prevDomains []string
		nextDomains []string
		expected    []string
	}{
		{
			desc:        "all empty",
			prevDomains: []string{},
			nextDomains: []string{},
			expected:    []string{},
		},
		{
			desc:        "next empty",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{},
			expected:    []string{"a", "b", "c"},
		},
		{
			desc:        "prev empty",
			prevDomains: []string{},
			nextDomains: []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
		},
		{
			desc:        "merge append",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{"a", "c", "d"},
			expected:    []string{"a", "b", "c", "d"},
		},
		{
			desc:        "merge same",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := merge(test.prevDomains, test.nextDomains)
			assert.Equal(t, test.expected, actual)
		})
	}
}
