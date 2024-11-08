package internal

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTSIGFile(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		expected *Key
	}{
		{
			desc:     "basic",
			filename: "sample.conf",
			expected: &Key{Name: "example.com", Algorithm: "hmac-sha256", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "data before the key",
			filename: "text_before.conf",
			expected: &Key{Name: "example.com", Algorithm: "hmac-sha256", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "data after the key",
			filename: "text_after.conf",
			expected: &Key{Name: "example.com", Algorithm: "hmac-sha256", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "ignore missing secret",
			filename: "missing_secret.conf",
			expected: &Key{Name: "example.com", Algorithm: "hmac-sha256"},
		},
		{
			desc:     "ignore missing algorithm",
			filename: "mising_algo.conf",
			expected: &Key{Name: "example.com", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "ignore invalid field format",
			filename: "invalid_field.conf",
			expected: &Key{Name: "example.com", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			key, err := ReadTSIGFile(filepath.Join("fixtures", test.filename))
			require.NoError(t, err)

			assert.Equal(t, test.expected, key)
		})
	}
}

func TestReadTSIGFile_error(t *testing.T) {
	if runtime.GOOS != "linux" {
		// Because error messages are different on Windows.
		t.Skip("only for UNIX systems")
	}

	testCases := []struct {
		desc     string
		filename string
		expected string
	}{
		{
			desc:     "missing file",
			filename: "missing.conf",
			expected: "open file: open fixtures/missing.conf: no such file or directory",
		},
		{
			desc:     "invalid key format",
			filename: "invalid_key.conf",
			expected: "invalid key line: key {",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := ReadTSIGFile(filepath.Join("fixtures", test.filename))
			require.Error(t, err)

			require.EqualError(t, err, test.expected)
		})
	}
}
