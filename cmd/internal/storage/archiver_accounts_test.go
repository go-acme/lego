package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiver_Accounts(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The error is:
		// TempDir RemoveAll cleanup: unlinkat C:\Users\RUNNER~1\AppData\Local\Temp\xxx: The process cannot access the file because it is being used by another process.
		t.Skip("skipping test on Windows")
	}

	cfg := &configuration.Configuration{
		Storage: t.TempDir(),
		Accounts: map[string]*configuration.Account{
			"foo": {
				Server:  "https://ca.example.com/dir",
				KeyType: "EC256",
			},
		},
	}

	archiver := NewArchiver(cfg.Storage)
	archiver.maxTimeBeforeCleaning = 0

	err := os.MkdirAll(archiver.accountsBasePath, 0o700)
	require.NoError(t, err)

	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.com", "EC256", "foo")
	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.com", "EC256", "bar")

	// archive

	err = archiver.Accounts(cfg)
	require.NoError(t, err)

	entries, err := os.ReadDir(archiver.accountsArchivePath)
	require.NoError(t, err)

	assert.Len(t, entries, 1)

	// clean

	err = archiver.Accounts(cfg)
	require.NoError(t, err)

	entries, err = os.ReadDir(archiver.accountsArchivePath)
	require.NoError(t, err)

	assert.Empty(t, entries)
}

func generateFakeAccountFiles(t *testing.T, accountsBasePath, server, keyType, accountID string) {
	t.Helper()

	filename := filepath.Join(accountsBasePath, server, accountID, keyType, "account.json")

	err := os.MkdirAll(filepath.Dir(filename), 0o700)
	require.NoError(t, err)

	file, err := os.Create(filename)
	require.NoError(t, err)

	r := Account{ID: accountID}

	err = json.NewEncoder(file).Encode(r)
	require.NoError(t, err)
}
