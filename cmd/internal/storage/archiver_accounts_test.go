package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/mattn/go-zglob"
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

	accounts, err := archiver.ListArchivedAccounts()
	assert.NoError(t, err)

	require.Len(t, accounts, 1)
	assert.Regexp(t, `.+_\d+\.zip`, accounts[0])

	// clean

	err = archiver.Accounts(cfg)
	require.NoError(t, err)

	accounts, err = archiver.ListArchivedAccounts()
	require.NoError(t, err)

	require.Empty(t, accounts)
}

func TestArchiver_Restore_accounts(t *testing.T) {
	cfg := &configuration.Configuration{
		Storage: t.TempDir(),
		Accounts: map[string]*configuration.Account{
			"foo": {
				Server:  "https://ca.example.org/dir",
				KeyType: "EC256",
			},
		},
	}

	archiver := NewArchiver(cfg.Storage)
	archiver.maxTimeBeforeCleaning = 0

	err := os.MkdirAll(archiver.accountsBasePath, 0o700)
	require.NoError(t, err)

	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.org", "foo")
	generateFakeAccountFiles(t, archiver.accountsBasePath, "ca.example.org", "bar")

	// archive

	err = archiver.Accounts(cfg)
	require.NoError(t, err)

	accounts, err := archiver.ListArchivedAccounts()
	assert.NoError(t, err)

	require.Len(t, accounts, 1)
	assert.Regexp(t, `.+_\d+\.zip`, accounts[0])

	// restore

	err = archiver.Restore(accounts[0])
	require.NoError(t, err)

	matches, err := zglob.Glob(filepath.Join(archiver.accountsBasePath, "**", "*.json"))
	require.NoError(t, err)

	assert.Len(t, matches, 2)

	accounts, err = archiver.ListArchivedAccounts()
	require.NoError(t, err)

	require.Empty(t, accounts)
}

func generateFakeAccountFiles(t *testing.T, accountsBasePath, server, accountID string) {
	t.Helper()

	filename := filepath.Join(accountsBasePath, server, accountID, "account.json")

	err := os.MkdirAll(filepath.Dir(filename), 0o700)
	require.NoError(t, err)

	file, err := os.Create(filename)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	r := Account{
		ID:      accountID,
		KeyType: certcrypto.EC256,
		Server:  "https://ca.example.com/dir",
		Origin:  OriginConfiguration,
	}

	err = json.NewEncoder(file).Encode(r)
	require.NoError(t, err)
}
