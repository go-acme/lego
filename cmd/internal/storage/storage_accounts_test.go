package storage

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountsStorage_GetRootPath(t *testing.T) {
	basePath := t.TempDir()

	storage := NewAccountsStorage(basePath)

	assert.Truef(t, strings.HasPrefix(storage.GetRootPath(), basePath),
		"%s is not prefixed with %s", storage.GetRootPath(), basePath)

	rootPath, err := filepath.Rel(basePath, storage.GetRootPath())
	require.NoError(t, err)

	assert.Equal(t, baseAccountsRootFolderName, rootPath)
}

func TestAccountsStorage_Save(t *testing.T) {
	basePath := t.TempDir()

	storage := NewAccountsStorage(basePath)

	accountID := "test@example.com"
	keyType := certcrypto.RSA4096

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	require.NoError(t, err)

	account := &Account{
		Email:   "account@example.com",
		ID:      accountID,
		Server:  "https://example.com/dir",
		KeyType: keyType,
		Registration: &acme.ExtendedAccount{
			Account: acme.Account{
				Status:                 "valid",
				Contact:                []string{"contact@example.com"},
				TermsOfServiceAgreed:   true,
				Orders:                 "https://ame.example.com/orders/123456",
				OnlyReturnExisting:     true,
				ExternalAccountBinding: []byte(`"EAB"`),
			},
			Location: "https://ame.example.com",
		},
		key: privateKey,
	}

	server, err := url.Parse(account.Server)
	require.NoError(t, err)

	accountFilePath := storage.getAccountFilePath(server, accountID)

	err = os.MkdirAll(filepath.Dir(accountFilePath), 0o755)
	require.NoError(t, err)

	err = storage.Save(account)
	require.NoError(t, err)

	require.FileExists(t, accountFilePath)
	assert.NoFileExists(t, storage.getAccountKeyPath(server, accountID))

	file, err := os.ReadFile(accountFilePath)
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", accountFileName))
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(file))
}

func TestAccountsStorage_Get_newAccount(t *testing.T) {
	storage := NewAccountsStorage(t.TempDir())

	email := "test@example.com"
	keyType := certcrypto.RSA4096

	server, err := url.Parse("https://example.com/dir")
	require.NoError(t, err)

	account, err := storage.Get(server.String(), keyType, email, "")
	require.NoError(t, err)

	assert.Equal(t, email, account.GetEmail())
	assert.Equal(t, email, account.GetID())
	assert.Nil(t, account.GetRegistration())
	assert.NotNil(t, account.GetPrivateKey())
	assert.False(t, account.NeedsRecovery)

	assert.FileExists(t, storage.getAccountFilePath(server, email))
	assert.FileExists(t, storage.getAccountKeyPath(server, email))
}

func TestAccountsStorage_Get_existingAccount(t *testing.T) {
	storage := NewAccountsStorage(t.TempDir())

	accountID := "test@example.com"
	email := "account@example.com"
	keyType := certcrypto.RSA4096

	server, err := url.Parse("https://example.com/dir")
	require.NoError(t, err)

	existingAccount := &Account{
		ID:      accountID,
		Email:   email,
		KeyType: keyType,
		Server:  server.String(),
		Registration: &acme.ExtendedAccount{
			Account: acme.Account{
				Status: "valid",
			},
			Location: "https://example.org/acme/acct/123456",
		},
	}

	createFakeAccountFile(t, storage, server, accountID, existingAccount)

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	require.NoError(t, err)

	err = os.WriteFile(storage.getAccountKeyPath(server, accountID), certcrypto.PEMEncode(privateKey), 0o600)
	require.NoError(t, err)

	account, err := storage.Get(server.String(), keyType, "", accountID)
	require.NoError(t, err)

	assert.Equal(t, email, account.GetEmail())
	assert.Equal(t, accountID, account.GetID())
	assert.Equal(t, keyType, account.KeyType)
	assert.Equal(t, "https://example.com/dir", account.Server)
	assert.False(t, account.NeedsRecovery)

	expectedRegistration := &acme.ExtendedAccount{
		Account: acme.Account{
			Status: "valid",
		},
		Location: "https://example.org/acme/acct/123456",
	}

	assert.Equal(t, expectedRegistration, account.GetRegistration())

	assert.NotNil(t, account.GetPrivateKey())
}

func TestAccountsStorage_Get_existingAccount_withoutPrivateKey(t *testing.T) {
	storage := NewAccountsStorage(t.TempDir())

	accountID := "test@example.com"
	email := "account@example.com"
	keyType := certcrypto.EC256

	server, err := url.Parse("https://example.com/dir")
	require.NoError(t, err)

	existingAccount := &Account{
		ID:      accountID,
		Email:   email,
		KeyType: keyType,
		Server:  server.String(),
		Registration: &acme.ExtendedAccount{
			Account: acme.Account{
				Status: "valid",
			},
			Location: "https://example.org/wrong/acct/123456",
		},
	}

	createFakeAccountFile(t, storage, server, accountID, existingAccount)

	account, err := storage.Get(server.String(), keyType, "", accountID)
	require.NoError(t, err)

	assert.Equal(t, email, account.GetEmail())
	assert.Equal(t, accountID, account.GetID())
	assert.Equal(t, keyType, account.KeyType)
	assert.Equal(t, "https://example.com/dir", account.Server)
	assert.Nil(t, account.GetRegistration())
	assert.False(t, account.NeedsRecovery)

	assert.NotNil(t, account.GetPrivateKey())
}

func TestAccountsStorage_Get_existingAccount_withoutRegistration(t *testing.T) {
	storage := NewAccountsStorage(t.TempDir())

	accountID := "test@example.com"
	email := "account@example.com"
	keyType := certcrypto.EC256

	server, err := url.Parse("https://example.com/dir")
	require.NoError(t, err)

	existingAccount := &Account{
		ID:      accountID,
		Email:   email,
		KeyType: keyType,
		Server:  server.String(),
	}

	createFakeAccountFile(t, storage, server, accountID, existingAccount)

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	require.NoError(t, err)

	err = os.WriteFile(storage.getAccountKeyPath(server, accountID), certcrypto.PEMEncode(privateKey), 0o600)
	require.NoError(t, err)

	account, err := storage.Get(server.String(), keyType, "", accountID)
	require.NoError(t, err)

	assert.Equal(t, email, account.GetEmail())
	assert.Equal(t, accountID, account.GetID())
	assert.Equal(t, keyType, account.KeyType)
	assert.Equal(t, "https://example.com/dir", account.Server)
	assert.Nil(t, account.GetRegistration())
	assert.True(t, account.NeedsRecovery)

	assert.NotNil(t, account.GetPrivateKey())
}

func createFakeAccountFile(t *testing.T, storage *AccountsStorage, server *url.URL, accountID string, existingAccount *Account) {
	t.Helper()

	accountFilePath := storage.getAccountFilePath(server, accountID)

	err := os.MkdirAll(filepath.Dir(accountFilePath), 0o700)
	require.NoError(t, err)

	file, err := os.Create(accountFilePath)
	require.NoError(t, err)

	t.Cleanup(func() { _ = file.Close() })

	err = json.NewEncoder(file).Encode(existingAccount)
	require.NoError(t, err)
}
