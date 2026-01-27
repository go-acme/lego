package storage

import (
	"crypto"
	"crypto/rsa"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/registration"
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

	account := &Account{
		Email: "account@example.com",
		ID:    accountID,
		Registration: &registration.Resource{
			Body: acme.Account{
				Status:                 "valid",
				Contact:                []string{"contact@example.com"},
				TermsOfServiceAgreed:   true,
				Orders:                 "https://ame.example.com/orders/123456",
				OnlyReturnExisting:     true,
				ExternalAccountBinding: []byte(`"EAB"`),
			},
			URI: "https://ame.example.com",
		},
		key: crypto.PrivateKey(""),
	}

	accountFilePath := storage.getAccountFilePath(accountID)

	err = os.MkdirAll(filepath.Dir(accountFilePath), 0o755)
	require.NoError(t, err)

	err = storage.Save(account)
	require.NoError(t, err)

	require.FileExists(t, accountFilePath)

	relativeAccountFilePath, err := filepath.Rel(basePath, accountFilePath)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(baseAccountsRootFolderName, accountID, accountFileName), relativeAccountFilePath)

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

	account, err := storage.Get(t.Context(), certcrypto.RSA4096, email, "")
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", account.GetEmail())
	assert.Equal(t, "test@example.com", account.GetID())
	assert.Nil(t, account.GetRegistration())
	assert.NotNil(t, account.GetPrivateKey())
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

	expectedRegistration := &registration.Resource{
		Body: acme.Account{
			Status: "valid",
		},
		URI: "https://example.org/acme/acct/123456",
	}

	assert.Equal(t, expectedRegistration, account.GetRegistration())

	assert.NotNil(t, account.GetPrivateKey())
}

func TestAccountsStorage_getPrivateKey(t *testing.T) {
	testCases := []struct {
		desc     string
		basePath string
	}{
		{
			desc: "create a new private key",
		},
		{
			desc:     "existing private key",
			basePath: "testdata",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			if test.basePath == "" {
				test.basePath = t.TempDir()
			}

			storage, err := NewAccountsStorage(AccountsStorageConfig{
				BasePath: test.basePath,
			})
			require.NoError(t, err)

			accountID := "test@example.com"

			expectedPath := filepath.Join(test.basePath, baseAccountsRootFolderName, accountID, baseKeysFolderName, "test@example.com.key")

			privateKey, err := storage.getPrivateKey(certcrypto.RSA4096, accountID)
			require.NoError(t, err)

			assert.FileExists(t, expectedPath)

			assert.IsType(t, &rsa.PrivateKey{}, privateKey)
		})
	}
}
