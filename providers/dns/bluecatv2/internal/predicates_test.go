package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPredicate(t *testing.T) {
	testCases := []struct {
		desc      string
		predicate fmt.Stringer
		expected  string
	}{
		{
			desc:      "Equals",
			predicate: Eq("foo", "bar"),
			expected:  "foo:eq('bar')",
		},
		{
			desc:      "Contains",
			predicate: Contains("foo", "bar"),
			expected:  "foo:contains('bar')",
		},
		{
			desc:      "Starts with",
			predicate: StartsWith("foo", "bar"),
			expected:  "foo:startsWith('bar')",
		},
		{
			desc:      "Ends with",
			predicate: EndsWith("foo", "bar"),
			expected:  "foo:endsWith('bar')",
		},
		{
			desc:      "Match a list of values",
			predicate: In("foo", "bar", "bir"),
			expected:  "foo:in('bar', 'bir')",
		},
		{
			desc:      "Combined: and",
			predicate: And(Eq("foo", "bar"), Eq("fii", "bir")),
			expected:  "foo:eq('bar') and fii:eq('bir')",
		},
		{
			desc: "Combined: multiple and",
			predicate: And(
				Eq("foo", "bar"),
				Eq("fii", "bir"),
				Eq("fuu", "bur"),
			),
			expected: "foo:eq('bar') and fii:eq('bir') and fuu:eq('bur')",
		},
		{
			desc:      "Combined: or",
			predicate: Or(Eq("foo", "bar"), Eq("foo", "bir")),
			expected:  "foo:eq('bar') or foo:eq('bir')",
		},
		{
			desc: "Combined: multiple or",
			predicate: Or(
				Eq("foo", "bar"),
				Eq("foo", "bir"),
				Eq("foo", "bur"),
			),
			expected: "foo:eq('bar') or foo:eq('bir') or foo:eq('bur')",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, test.predicate.String())
		})
	}
}
