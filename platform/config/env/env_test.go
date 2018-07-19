package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetOrDefaultInt(t *testing.T) {
	tcs := []struct {
		EnvValue string
		Default  int
		Expected int
	}{
		{"100", 2, 100},
		{"abc123", 2, 2},
		{"-111", 2, -111},
		{"1.11", 2, 2},
	}

	const key = "LEGO_ENV_TC"
	for _, tc := range tcs {
		os.Setenv(key, tc.EnvValue)
		defer os.Unsetenv(key)

		res := GetOrDefaultInt(key, tc.Default)
		assert.Equal(t, tc.Expected, res)
	}
}
