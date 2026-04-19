package dotenv

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		desc      string
		filenames []string
		expected  []string
	}{
		{
			desc: "no file",
		},
		{
			desc:      "non-existing file",
			filenames: []string{filepath.Join("testdata", ".env.lego.non-existing")},
		},
		{
			desc:      "simple",
			filenames: []string{filepath.Join("testdata", ".env.lego.bar")},
			expected:  []string{"LEGO_TEST_ENV_A=aGlobal", "LEGO_TEST_ENV_B=bGlobal"},
		},
		{
			desc:      "multiple files",
			filenames: []string{filepath.Join("testdata", ".env.lego.bar"), filepath.Join("testdata", ".env.lego.foo")},
			expected:  []string{"LEGO_TEST_ENV_A=aLocal", "LEGO_TEST_ENV_B=bGlobal"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cleanUp, err := Load(test.filenames...)

			t.Cleanup(cleanUp)

			require.NoError(t, err)

			assert.Equal(t, test.expected, getTestEnviron())

			cleanUp()

			assert.Empty(t, getTestEnviron())
		})
	}
}

func getTestEnviron() []string {
	var result []string

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "LEGO_TEST_ENV_") {
			result = append(result, v)
		}
	}

	slices.Sort(result)

	return result
}
