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

func TestAccountsStorage_GetUserID(t *testing.T) {
	testCases := []struct {
		desc     string
		email    string
		expected string
	}{
		{
			desc:     "with email",
			email:    "test@example.com",
			expected: "test@example.com",
		},
		{
			desc:     "without email",
			expected: "noemail@example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			storage, err := NewAccountsStorage(AccountsStorageConfig{
				Email:    test.email,
				BasePath: t.TempDir(),
			})
			require.NoError(t, err)

			assert.Equal(t, test.email, storage.GetEmail())
			assert.Equal(t, test.expected, storage.GetUserID())
		})
	}
}

func TestAccountsStorage_ExistsAccountFilePath(t *testing.T) {
	testCases := []struct {
		desc   string
		setup  func(t *testing.T, storage *AccountsStorage)
		assert assert.BoolAssertionFunc
	}{
		{
			desc: "an account file exists",
			setup: func(t *testing.T, storage *AccountsStorage) {
				t.Helper()

				err := os.MkdirAll(filepath.Dir(storage.accountFilePath), 0o755)
				require.NoError(t, err)

				err = os.WriteFile(storage.accountFilePath, []byte("test"), 0o644)
				require.NoError(t, err)
			},
			assert: assert.True,
		},
		{
			desc:   "no account file",
			assert: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			storage, err := NewAccountsStorage(AccountsStorageConfig{
				BasePath: t.TempDir(),
			})
			require.NoError(t, err)

			if test.setup != nil {
				test.setup(t, storage)
			}

			test.assert(t, storage.ExistsAccountFilePath())
		})
	}
}

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
		Email:    "test@example.com",
		BasePath: basePath,
	})
	require.NoError(t, err)

	account := &Account{
		Email: "account@example.com",
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

	err = os.MkdirAll(filepath.Dir(storage.accountFilePath), 0o755)
	require.NoError(t, err)

	err = storage.Save(account)
	require.NoError(t, err)

	require.FileExists(t, storage.accountFilePath)

	accountFilePath, err := filepath.Rel(basePath, storage.accountFilePath)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(baseAccountsRootFolderName, "test@example.com", accountFileName), accountFilePath)

	file, err := os.ReadFile(storage.accountFilePath)
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join("testdata", accountFileName))
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(file))
}

func TestAccountsStorage_LoadAccount(t *testing.T) {
	storage, err := NewAccountsStorage(AccountsStorageConfig{
		Email:    "test@example.com",
		BasePath: t.TempDir(),
	})
	require.NoError(t, err)

	storage.accountFilePath = filepath.Join("testdata", accountFileName)

	account := storage.LoadAccount(t.Context(), "")

	expected := &Account{
		Email: "account@example.com",
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

	assert.Equal(t, expected, account)
}

func TestAccountsStorage_GetPrivateKey(t *testing.T) {
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
				Email:    "test@example.com",
				BasePath: test.basePath,
			})
			require.NoError(t, err)

			expectedPath := filepath.Join(test.basePath, baseAccountsRootFolderName, "test@example.com", baseKeysFolderName, "test@example.com.key")

			privateKey := storage.GetPrivateKey(certcrypto.RSA4096)

			assert.FileExists(t, expectedPath)

			assert.IsType(t, &rsa.PrivateKey{}, privateKey)
		})
	}
}
