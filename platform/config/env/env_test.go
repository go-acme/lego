package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWithFallback(t *testing.T) {
	var1Exist := os.Getenv("TEST_LEGO_VAR_EXIST_1")
	var2Exist := os.Getenv("TEST_LEGO_VAR_EXIST_2")
	var1Missing := os.Getenv("TEST_LEGO_VAR_MISSING_1")
	var2Missing := os.Getenv("TEST_LEGO_VAR_MISSING_2")

	t.Cleanup(func() {
		_ = os.Setenv("TEST_LEGO_VAR_EXIST_1", var1Exist)
		_ = os.Setenv("TEST_LEGO_VAR_EXIST_2", var2Exist)
		_ = os.Setenv("TEST_LEGO_VAR_MISSING_1", var1Missing)
		_ = os.Setenv("TEST_LEGO_VAR_MISSING_2", var2Missing)
	})

	err := os.Setenv("TEST_LEGO_VAR_EXIST_1", "VAR1")
	require.NoError(t, err)
	err = os.Setenv("TEST_LEGO_VAR_EXIST_2", "VAR2")
	require.NoError(t, err)
	err = os.Unsetenv("TEST_LEGO_VAR_MISSING_1")
	require.NoError(t, err)
	err = os.Unsetenv("TEST_LEGO_VAR_MISSING_2")
	require.NoError(t, err)

	type expected struct {
		value map[string]string
		error string
	}

	testCases := []struct {
		desc     string
		groups   [][]string
		expected expected
	}{
		{
			desc:   "no groups",
			groups: nil,
			expected: expected{
				value: map[string]string{},
			},
		},
		{
			desc:   "empty groups",
			groups: [][]string{{}, {}},
			expected: expected{
				error: "undefined environment variable names",
			},
		},
		{
			desc:   "missing env var",
			groups: [][]string{{"TEST_LEGO_VAR_MISSING_1"}},
			expected: expected{
				error: "some credentials information are missing: TEST_LEGO_VAR_MISSING_1",
			},
		},
		{
			desc:   "all env var in a groups are missing",
			groups: [][]string{{"TEST_LEGO_VAR_MISSING_1", "TEST_LEGO_VAR_MISSING_2"}},
			expected: expected{
				error: "some credentials information are missing: TEST_LEGO_VAR_MISSING_1",
			},
		},
		{
			desc:   "only the first env var have a value",
			groups: [][]string{{"TEST_LEGO_VAR_EXIST_1", "TEST_LEGO_VAR_MISSING_1"}},
			expected: expected{
				value: map[string]string{"TEST_LEGO_VAR_EXIST_1": "VAR1"},
			},
		},
		{
			desc:   "only the second env var have a value",
			groups: [][]string{{"TEST_LEGO_VAR_MISSING_1", "TEST_LEGO_VAR_EXIST_1"}},
			expected: expected{
				value: map[string]string{"TEST_LEGO_VAR_MISSING_1": "VAR1"},
			},
		},
		{
			desc:   "all env vars in a groups have a value",
			groups: [][]string{{"TEST_LEGO_VAR_EXIST_1", "TEST_LEGO_VAR_EXIST_2"}},
			expected: expected{
				value: map[string]string{"TEST_LEGO_VAR_EXIST_1": "VAR1"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			value, err := GetWithFallback(test.groups...)
			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.value, value)
			}
		})
	}
}

func TestGetOneWithFallback(t *testing.T) {
	var1Exist := os.Getenv("TEST_LEGO_VAR_EXIST_1")
	var2Exist := os.Getenv("TEST_LEGO_VAR_EXIST_2")
	var1Missing := os.Getenv("TEST_LEGO_VAR_MISSING_1")
	var2Missing := os.Getenv("TEST_LEGO_VAR_MISSING_2")

	t.Cleanup(func() {
		_ = os.Setenv("TEST_LEGO_VAR_EXIST_1", var1Exist)
		_ = os.Setenv("TEST_LEGO_VAR_EXIST_2", var2Exist)
		_ = os.Setenv("TEST_LEGO_VAR_MISSING_1", var1Missing)
		_ = os.Setenv("TEST_LEGO_VAR_MISSING_2", var2Missing)
	})

	err := os.Setenv("TEST_LEGO_VAR_EXIST_1", "VAR1")
	require.NoError(t, err)
	err = os.Setenv("TEST_LEGO_VAR_EXIST_2", "VAR2")
	require.NoError(t, err)
	err = os.Unsetenv("TEST_LEGO_VAR_MISSING_1")
	require.NoError(t, err)
	err = os.Unsetenv("TEST_LEGO_VAR_MISSING_2")
	require.NoError(t, err)

	testCases := []struct {
		desc         string
		main         string
		defaultValue string
		alts         []string
		expected     string
	}{
		{
			desc:         "with value and no alternative",
			main:         "TEST_LEGO_VAR_EXIST_1",
			defaultValue: "oops",
			expected:     "VAR1",
		},
		{
			desc:         "with value and alternatives",
			main:         "TEST_LEGO_VAR_EXIST_1",
			defaultValue: "oops",
			alts:         []string{"TEST_LEGO_VAR_MISSING_1"},
			expected:     "VAR1",
		},
		{
			desc:         "without value and no alternatives",
			main:         "TEST_LEGO_VAR_MISSING_1",
			defaultValue: "oops",
			expected:     "oops",
		},
		{
			desc:         "without value and alternatives",
			main:         "TEST_LEGO_VAR_MISSING_1",
			defaultValue: "oops",
			alts:         []string{"TEST_LEGO_VAR_EXIST_1"},
			expected:     "VAR1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			value := GetOneWithFallback(test.main, test.defaultValue, ParseString, test.alts...)
			assert.Equal(t, test.expected, value)
		})
	}
}

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
			t.Setenv(key, test.envValue)

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

	key := "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Setenv(key, test.envValue)

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

	key := "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Setenv(key, test.envValue)

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

	key := "LEGO_ENV_TC"

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Setenv(key, test.envValue)

			actual := GetOrDefaultBool(key, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetOrFile_ReadsEnvVars(t *testing.T) {
	t.Setenv("TEST_LEGO_ENV_VAR", "lego_env")

	value := GetOrFile("TEST_LEGO_ENV_VAR")

	assert.Equal(t, "lego_env", value)
}

func TestGetOrFile_ReadsFiles(t *testing.T) {
	varEnvFileName := "TEST_LEGO_ENV_VAR_FILE"
	varEnvName := "TEST_LEGO_ENV_VAR"

	testCases := []struct {
		desc        string
		fileContent []byte
	}{
		{
			desc:        "simple",
			fileContent: []byte("lego_file"),
		},
		{
			desc:        "with an empty last line",
			fileContent: []byte("lego_file\n"),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			err := os.Unsetenv(varEnvFileName)
			require.NoError(t, err)
			err = os.Unsetenv(varEnvName)
			require.NoError(t, err)

			file, err := os.CreateTemp("", "lego")
			require.NoError(t, err)
			defer os.Remove(file.Name())

			err = os.WriteFile(file.Name(), []byte("lego_file\n"), 0o644)
			require.NoError(t, err)

			t.Setenv(varEnvFileName, file.Name())

			value := GetOrFile(varEnvName)

			assert.Equal(t, "lego_file", value)
		})
	}
}

func TestGetOrFile_PrefersEnvVars(t *testing.T) {
	varEnvFileName := "TEST_LEGO_ENV_VAR_FILE"
	varEnvName := "TEST_LEGO_ENV_VAR"

	err := os.Unsetenv(varEnvFileName)
	require.NoError(t, err)
	err = os.Unsetenv(varEnvName)
	require.NoError(t, err)

	file, err := os.CreateTemp("", "lego")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	err = os.WriteFile(file.Name(), []byte("lego_file"), 0o644)
	require.NoError(t, err)

	t.Setenv(varEnvFileName, file.Name())
	t.Setenv(varEnvName, "lego_env")

	value := GetOrFile(varEnvName)

	assert.Equal(t, "lego_env", value)
}
