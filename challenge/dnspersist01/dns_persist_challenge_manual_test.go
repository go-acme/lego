package dnspersist01

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatTXTValue(t *testing.T) {
	longValue := strings.Repeat("z", 256)

	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "single quoted string",
			value:    "abc",
			expected: `"abc"`,
		},
		{
			desc:     "split and quoted across chunks",
			value:    longValue,
			expected: fmt.Sprintf("%q %q", longValue[:255], longValue[255:]),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := formatTXTValue(test.value)

			assert.Equal(t, test.expected, actual)
		})
	}
}
