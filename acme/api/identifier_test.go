package api

import (
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/stretchr/testify/assert"
)

func Test_compareIdentifiers(t *testing.T) {
	testCases := []struct {
		desc     string
		a, b     []acme.Identifier
		expected int
	}{
		{
			desc: "identical identifiers",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			expected: 0,
		},
		{
			desc: "identical identifiers but different order",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "*.example.com"},
				{Type: "dns", Value: "example.com"},
			},
			expected: 0,
		},
		{
			desc: "duplicate identifiers",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "example.com"},
			},
			expected: -1,
		},
		{
			desc: "different identifier values",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.org"},
			},
			expected: -1,
		},
		{
			desc: "different identifier types",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "ip", Value: "*.example.com"},
			},
			expected: -1,
		},
		{
			desc: "different number of identifiers a>b",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
				{Type: "dns", Value: "example.org"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			expected: 1,
		},
		{
			desc: "different number of identifiers b>a",
			a: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
			},
			b: []acme.Identifier{
				{Type: "dns", Value: "example.com"},
				{Type: "dns", Value: "*.example.com"},
				{Type: "dns", Value: "example.org"},
			},
			expected: -1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, compareIdentifiers(test.a, test.b))
		})
	}
}
