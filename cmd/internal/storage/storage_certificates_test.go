package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SanitizedName(t *testing.T) {
	// IDN examples from https://www.iana.org/domains/reserved
	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "emojis",
			value:    "❤️🧑🏿‍❤️‍💋‍🧑🏻.com",
			expected: "xn--1ugaa726eba24804aca46741c4a81le84aha.com",
		},
		{
			desc:     "Tamil",
			value:    "பரிட்சை.com",
			expected: "xn--hlcj6aya9esc7a.com",
		},
		{
			desc:     "Katakana",
			value:    "テスト.com",
			expected: "xn--zckzah.com",
		},
		{
			desc:     "Devanagari (Nagari)",
			value:    "परीक्षा.com",
			expected: "xn--11b5bs3a9aj6g.com",
		},
		{
			desc:     "Han (Traditional variant)",
			value:    "測試.com",
			expected: "xn--g6w251d.com",
		},
		{
			desc:     "Han (Simplified variant)",
			value:    "测试.com",
			expected: "xn--0zwm56d.com",
		},
		{
			desc:     "slash and anti-slash",
			value:    "bar\\foo./example.com",
			expected: "barfoo.example.com",
		},
		{
			desc:     "whitespaces",
			value:    "bar\tfoo .ex\nam\rple.com",
			expected: "barfoo.example.com",
		},
		{
			desc:     "wildcard",
			value:    "*.example.com",
			expected: "_.example.com",
		},
		{
			desc:     "hyphen",
			value:    "foo-bar.example.com",
			expected: "foo-bar.example.com",
		},
		{
			desc:     "special chars",
			value:    "foo${}bar&().ex[]ample.com",
			expected: "foobar.example.com",
		},
		{
			desc:     "ip",
			value:    "127.0.0.1",
			expected: "127.0.0.1",
		},
		{
			desc:     "email",
			value:    "fée@example.com",
			expected: "xn--fe@example-b7a.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, SanitizedName(test.value))
		})
	}
}
