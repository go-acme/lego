package internal

import (
	"path/filepath"
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
			expected: &Key{Name: "lego", Algorithm: "hmac-sha256.", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "data before the key",
			filename: "text_before.conf",
			expected: &Key{Name: "lego", Algorithm: "hmac-sha256.", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "data after the key",
			filename: "text_after.conf",
			expected: &Key{Name: "lego", Algorithm: "hmac-sha256.", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
		},
		{
			desc:     "missing secret",
			filename: "missing_secret.conf",
			expected: &Key{Name: "lego", Algorithm: "hmac-sha256."},
		},
		{
			desc:     "missing algorithm",
			filename: "mising_algo.conf",
			expected: &Key{Name: "lego", Secret: "TCG5A6/lOHUGbW0e/9RYYbzWDFMlj1pIxCvybLBayBg="},
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
	_, err := ReadTSIGFile(filepath.Join("fixtures", "invalid_key.conf"))
	require.Error(t, err)
}
