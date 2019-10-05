package http01

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseForwardedHeader(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []map[string]string
		err   string
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "simple case",
			input: `for=1.2.3.4;host=example.com; by=127.0.0.1`,
			want: []map[string]string{
				{"for": "1.2.3.4", "host": "example.com", "by": "127.0.0.1"},
			},
		},
		{
			name:  "quoted-string",
			input: `foo="bar"`,
			want: []map[string]string{
				{"foo": "bar"},
			},
		},
		{
			name:  "multiple entries",
			input: `a=1, b=2; c=3, d=4`,
			want: []map[string]string{
				{"a": "1"},
				{"b": "2", "c": "3"},
				{"d": "4"},
			},
		},
		{
			name:  "whitespace",
			input: "   a =  1,\tb\n=\r\n2,c=\"   untrimmed  \"",
			want: []map[string]string{
				{"a": "1"},
				{"b": "2"},
				{"c": "   untrimmed  "},
			},
		},
		{
			name:  "unterminated quote",
			input: `x="y`,
			err:   "unterminated quoted-string",
		},
		{
			name:  "unexpected quote",
			input: `"x=y"`,
			err:   "unexpected quote",
		},
		{
			name:  "invalid token",
			input: `a=b, ipv6=[fe80::1], x=y`,
			err:   "invalid token character at pos 10: [",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseForwardedHeader(test.input)
			if test.err == "" {
				require.NoError(t, err)
				assert.EqualValues(t, test.want, actual)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
			}
		})
	}
}
