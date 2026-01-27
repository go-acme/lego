package storage

import (
	"context"
	"crypto"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
)

const (
	baseAccountsRootFolderName = "accounts"
	baseKeysFolderName         = "keys"
	accountFileName            = "account.json"
)

type AccountsStorageConfig struct {
	BasePath string

	Server    string
	UserAgent string
}

// AccountsStorage A storage for account data.
//
// rootPath:
//
//	./.lego/accounts/
//	     │      └── root accounts directory
//	     └── "path" option
//
// rootUserPath:
//
//	./.lego/accounts/localhost_14000/foo@example.com/
//	     │      │             │             └── userID ("email" option)
//	     │      │             └── CA server ("server" option)
//	     │      └── root accounts directory
//	     └── "path" option
//
// keysPath:
//
//	./.lego/accounts/localhost_14000/foo@example.com/keys/
//	     │      │             │             │           └── root keys directory
//	     │      │             │             └── userID ("email" option)
//	     │      │             └── CA server ("server" option)
//	     │      └── root accounts directory
//	     └── "path" option
//
// accountFilePath:
//
//	./.lego/accounts/localhost_14000/foo@example.com/account.json
//	     │      │             │             │             └── account file
//	     │      │             │             └── userID ("email" option)
//	     │      │             └── CA server ("server" option)
//	     │      └── root accounts directory
//	     └── "path" option
type AccountsStorage struct {
	rootPath string

	server    *url.URL
	userAgent string
}

// NewAccountsStorage Creates a new AccountsStorage.
func NewAccountsStorage(config AccountsStorageConfig) (*AccountsStorage, error) {
	serverURL, err := url.Parse(config.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q: %w", config.Server, err)
	}

	return &AccountsStorage{
		rootPath:  filepath.Join(config.BasePath, baseAccountsRootFolderName),
		server:    serverURL,
		userAgent: config.UserAgent,
	}, nil
}

func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

func (s *AccountsStorage) Save(account *Account) error {
	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(s.getAccountFilePath(account.GetID()), jsonBytes, filePerm)
}

func (s *AccountsStorage) Get(ctx context.Context, keyType certcrypto.KeyType, email, accountID string) (*Account, error) {
	effectiveAccountID := getEffectiveAccountID(email, accountID)

	if !s.existsAccountFile(effectiveAccountID) {
		return s.newAccount(keyType, email, accountID)
	}

	return s.readAccount(ctx, keyType, effectiveAccountID)
}

func (s *AccountsStorage) newAccount(keyType certcrypto.KeyType, email, accountID string) (*Account, error) {
	privateKey, err := s.getPrivateKey(keyType, getEffectiveAccountID(email, accountID))
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}

	return NewAccount(email, accountID, privateKey), nil
}

func (s *AccountsStorage) readAccount(ctx context.Context, keyType certcrypto.KeyType, effectiveAccountID string) (*Account, error) {
	accountFilePath := s.getAccountFilePath(effectiveAccountID)

	fileBytes, err := os.ReadFile(accountFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read the account file %q : %w", accountFilePath, err)
	}

	account := new(Account)

	err = json.Unmarshal(fileBytes, account)
	if err != nil {
		return nil, fmt.Errorf("could not parse the account file %s: %w", accountFilePath, err)
	}

	account.key, err = s.getPrivateKey(keyType, effectiveAccountID)
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := s.tryRecoverRegistration(ctx, account.key)
		if err != nil {
			return nil, fmt.Errorf("could not load the account file, registration is nil (accountID: %s): %w", effectiveAccountID, err)
		}

		account.Registration = reg

		err = s.Save(account)
		if err != nil {
			return nil, fmt.Errorf("could not save the account file, registration is nil (accountID: %s): %w", effectiveAccountID, err)
		}
	}

	return account, nil
}

func (s *AccountsStorage) getPrivateKey(keyType certcrypto.KeyType, accountID string) (crypto.PrivateKey, error) {
	keysPath := s.getKeysPath(accountID)

	accKeyPath := filepath.Join(keysPath, accountID+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		// TODO(ldez): debug level?
		log.Info("No key found for the account. Generating a new private key.",
			slog.String("accountID", accountID),
			slog.Any("keyType", keyType),
		)

		err := CreateNonExistingFolder(keysPath)
		if err != nil {
			return nil, fmt.Errorf("could not check/create the directory %q for the account (accountID: %s): %w", keysPath, accountID, err)
		}

		privateKey, err := generatePrivateKey(accKeyPath, keyType)
		if err != nil {
			return nil, fmt.Errorf("could not generate the private account key (accountID: %s): %w", accountID, err)
		}

		// TODO(ldez): debug level?
		log.Info("Saved key.", slog.String("filepath", accKeyPath))

		return privateKey, nil
	}

	privateKey, err := LoadPrivateKey(accKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not load the private key from the file %q: %w", accKeyPath, err)
	}

	return privateKey, nil
}

func (s *AccountsStorage) existsAccountFile(effectiveAccountID string) bool {
	accountFilePath := s.getAccountFilePath(effectiveAccountID)

	if _, err := os.Stat(accountFilePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatal("Could not read the account file.",
			slog.String("filepath", accountFilePath),
			log.ErrorAttr(err),
		)
	}

	return true
}

func (s *AccountsStorage) getAccountFilePath(accountID string) string {
	return filepath.Join(s.getRootUserPath(accountID), accountFileName)
}

func (s *AccountsStorage) getKeysPath(accountID string) string {
	return filepath.Join(s.getRootUserPath(accountID), baseKeysFolderName)
}

func (s *AccountsStorage) getRootUserPath(accountID string) string {
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(s.server.Host)

	return filepath.Join(s.rootPath, serverPath, accountID)
}

func (s *AccountsStorage) tryRecoverRegistration(ctx context.Context, privateKey crypto.PrivateKey) (*registration.Resource, error) {
	// couldn't load account but got a key. Try to look the account up.
	config := lego.NewConfig(&Account{key: privateKey})
	config.CADirURL = s.server.String()
	config.UserAgent = s.userAgent

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	reg, err := client.Registration.ResolveAccountByKey(ctx)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func LoadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func generatePrivateKey(file string, keyType certcrypto.KeyType) (crypto.PrivateKey, error) {
	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, err
	}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer certOut.Close()

	pemKey := certcrypto.PEMBlock(privateKey)

	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
