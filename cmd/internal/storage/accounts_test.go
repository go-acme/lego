package storage

import (
	"crypto"
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

	storage, err := NewAccountsStorage(AccountsStorageConfig{
		BasePath: basePath,
	})
	require.NoError(t, err)

	assert.Truef(t, strings.HasPrefix(storage.GetRootPath(), basePath),
		"%s is not prefixed with %s", storage.GetRootPath(), basePath)

	rootPath, err := filepath.Rel(basePath, storage.GetRootPath())
	require.NoError(t, err)

	assert.Equal(t, baseAccountsRootFolderName, rootPath)
}

func TestAccountsStorage_Save(t *testing.T) {
	basePath := t.TempDir()

	storage, err := NewAccountsStorage(AccountsStorageConfig{
		BasePath: basePath,
	})
	require.NoError(t, err)

	accountID := "test@example.com"
	keyType := certcrypto.RSA4096

	account := &Account{
		Email:   "account@example.com",
		ID:      accountID,
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
		key: crypto.PrivateKey(""),
	}

	accountFilePath := storage.getAccountFilePath(keyType, accountID)

	err = os.MkdirAll(filepath.Dir(accountFilePath), 0o755)
	require.NoError(t, err)

	err = storage.Save(keyType, account)
	require.NoError(t, err)

	require.FileExists(t, accountFilePath)
	assert.NoFileExists(t, filepath.Join(storage.getKeyPath(keyType, accountID), account.GetID()+ExtKey))

	file, err := os.ReadFile(accountFilePath)
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", accountFileName))
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(file))
}

func TestAccountsStorage_Get_newAccount(t *testing.T) {
	storage, err := NewAccountsStorage(AccountsStorageConfig{
		BasePath: t.TempDir(),
	})
	require.NoError(t, err)

	email := "test@example.com"
	keyType := certcrypto.RSA4096

	account, err := storage.Get(t.Context(), keyType, email, "")
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", account.GetEmail())
	assert.Equal(t, "test@example.com", account.GetID())
	assert.Nil(t, account.GetRegistration())
	assert.NotNil(t, account.GetPrivateKey())

	assert.FileExists(t, storage.getAccountFilePath(keyType, email))
	assert.FileExists(t, filepath.Join(storage.getKeyPath(keyType, email), account.GetID()+ExtKey))
}

func TestAccountsStorage_Get_existingAccount(t *testing.T) {
	storage, err := NewAccountsStorage(AccountsStorageConfig{
		BasePath: "testdata",
	})
	require.NoError(t, err)

	accountID := "test@example.com"

	account, err := storage.Get(t.Context(), certcrypto.RSA4096, "", accountID)
	require.NoError(t, err)

	assert.Equal(t, "account@example.com", account.GetEmail())
	assert.Equal(t, "test@example.com", account.GetID())
	assert.Equal(t, certcrypto.RSA4096, account.KeyType)

	expectedRegistration := &acme.ExtendedAccount{
		Account: acme.Account{
			Status: "valid",
		},
		Location: "https://example.org/acme/acct/123456",
	}

	assert.Equal(t, expectedRegistration, account.GetRegistration())

	assert.NotNil(t, account.GetPrivateKey())
}
