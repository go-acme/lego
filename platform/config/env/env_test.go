package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrDefaultInt(t *testing.T) {
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

func TestGetOrDefaultSecond(t *testing.T) {
	testCases := []struct {
		desc         string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			desc:         "valid value",
			envValue:     "100",
			defaultValue: 2 * time.Second,
			expected:     100 * time.Second,
		},
		{
			desc:         "invalid content, use default value",
			envValue:     "abc123",
			defaultValue: 2 * time.Second,
			expected:     2 * time.Second,
		},
		{
			desc:         "invalid content, negative value",
			envValue:     "-111",
			defaultValue: 2 * time.Second,
			expected:     2 * time.Second,
		},
		{
			desc:         "float: invalid type, use default value",
			envValue:     "1.11",
			defaultValue: 2 * time.Second,
			expected:     2 * time.Second,
		},
	}

	var key = "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer os.Unsetenv(key)
			err := os.Setenv(key, test.envValue)
			require.NoError(t, err)

			result := GetOrDefaultSecond(key, test.defaultValue)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetOrDefaultString(t *testing.T) {
	testCases := []struct {
		desc         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			desc:         "missing env var",
			defaultValue: "foo",
			expected:     "foo",
		},
		{
			desc:         "with env var",
			envValue:     "bar",
			defaultValue: "foo",
			expected:     "bar",
		},
	}

	var key = "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer os.Unsetenv(key)
			err := os.Setenv(key, test.envValue)
			require.NoError(t, err)

			actual := GetOrDefaultString(key, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetOrDefaultBool(t *testing.T) {
	testCases := []struct {
		desc         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			desc:         "missing env var",
			defaultValue: true,
			expected:     true,
		},
		{
			desc:         "with env var",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			desc:         "invalid value",
			envValue:     "foo",
			defaultValue: false,
			expected:     false,
		},
	}

	var key = "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer os.Unsetenv(key)
			err := os.Setenv(key, test.envValue)
			require.NoError(t, err)

			actual := GetOrDefaultBool(key, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}
