package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetOrDefaultInt(t *testing.T) {
	testCases := []struct {
		desc         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			desc:         "valid value",
			envValue:     "100",
			defaultValue: 2,
			expected:     100,
		},
		{
			desc:         "invalid content, use default value",
			envValue:     "abc123",
			defaultValue: 2,
			expected:     2,
		},
		{
			desc:         "valid negative value",
			envValue:     "-111",
			defaultValue: 2,
			expected:     -111,
		},
		{
			desc:         "float: invalid type, use default value",
			envValue:     "1.11",
			defaultValue: 2,
			expected:     2,
		},
	}

	const key = "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer os.Unsetenv(key)
			err := os.Setenv(key, test.envValue)
			require.NoError(t, err)

			result := GetOrDefaultInt(key, test.defaultValue)
			assert.Equal(t, test.expected, result)
		})
	}
}
