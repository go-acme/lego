package internal

import (
	"sort"
	"testing"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/stretchr/testify/assert"
)

func TestGetNameservers(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		stack    challenge.NetworkStack
		expected []string
	}{
		{
			desc:     "with resolv.conf",
			path:     "fixtures/resolv.conf.1",
			stack:    challenge.DualStack,
			expected: []string{"10.200.3.249", "10.200.3.250:5353", "2001:4860:4860::8844", "[10.0.0.1]:5353"},
		},
		{
			desc:     "with nonexistent resolv.conf",
			path:     "fixtures/resolv.conf.nonexistant",
			stack:    challenge.DualStack,
			expected: []string{"1.0.0.1:53", "1.1.1.1:53", "[2606:4700:4700::1001]:53", "[2606:4700:4700::1111]:53"},
		},
		{
			desc:     "default with IPv4Only",
			path:     "resolv.conf.nonexistant",
			stack:    challenge.IPv4Only,
			expected: []string{"1.0.0.1:53", "1.1.1.1:53"},
		},
		{
			desc:     "default with IPv6Only",
			path:     "resolv.conf.nonexistant",
			stack:    challenge.IPv6Only,
			expected: []string{"[2606:4700:4700::1001]:53", "[2606:4700:4700::1111]:53"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			result := GetNameservers(test.path, test.stack)

			sort.Strings(result)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseNameservers(t *testing.T) {
	testCases := []struct {
		desc     string
		servers  []string
		expected []string
	}{
		{
			desc:     "without explicit port",
			servers:  []string{"ns1.example.com", "2001:db8::1"},
			expected: []string{"ns1.example.com:53", "[2001:db8::1]:53"},
		},
		{
			desc:     "with explicit port",
			servers:  []string{"ns1.example.com:53", "[2001:db8::1]:53"},
			expected: []string{"ns1.example.com:53", "[2001:db8::1]:53"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := ParseNameservers(test.servers)

			assert.Equal(t, test.expected, result)
		})
	}
}
