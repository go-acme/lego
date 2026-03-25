package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiver_Accounts(t *testing.T) {
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

	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.com", "foo")
	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.com", "bar")

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

func generateFakeAccountFiles(t *testing.T, accountsBasePath, server, accountID string) {
	t.Helper()

	filename := filepath.Join(accountsBasePath, server, accountID, "account.json")

	err := os.MkdirAll(filepath.Dir(filename), 0o700)
	require.NoError(t, err)

	file, err := os.Create(filename)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	r := Account{ID: accountID}

	err = json.NewEncoder(file).Encode(r)
	require.NoError(t, err)
}
