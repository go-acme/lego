package storage

import (
	"context"
	"crypto"
	"encoding/json"
	"encoding/pem"
	"errors"
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
	accountFileName            = "account.json"
)

type PrivateKeyNotFound struct {
	AccountID string
}

func (e *PrivateKeyNotFound) Error() string {
	return fmt.Sprintf("no private key found for account %q", e.AccountID)
}

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
//	     │      │             │             └── accountID
//	     │      │             └── CA server ("server" option)
//	     │      └── root accounts directory
//	     └── "path" option
//
// keysPath:
//
//	./.lego/accounts/localhost_14000/foo@example.com/RSA4096/
//	     │      │             │             │          └── per key type directory
//	     │      │             │             └── accountID
//	     │      │             └── CA server ("server" option)
//	     │      └── root accounts directory
//	     └── "path" option
//
// accountFilePath:
//
//	./.lego/accounts/localhost_14000/foo@example.com/RSA4096/account.json
//	     │      │             │             │          │       └── account file
//	     │      │             │             │          └── per key type directory
//	     │      │             │             └── accountID
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

// GetRootPath returns the root path of the storage of the accounts.
func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

// Save saves the account to a file (only the JSON file).
func (s *AccountsStorage) Save(keyType certcrypto.KeyType, account *Account) error {
	if account.ID == "" {
		account.ID = account.GetID()
	}

	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(s.getAccountFilePath(keyType, account.GetID()), jsonBytes, filePerm)
}

// Get gets an account from a file or creates a new one (the files are saved).
func (s *AccountsStorage) Get(ctx context.Context, keyType certcrypto.KeyType, email, accountID string) (*Account, error) {
	effectiveAccountID := getEffectiveAccountID(email, accountID)

	if !s.existsAccountFile(keyType, effectiveAccountID) {
		account, err := s.createAccount(keyType, email, accountID)
		if err != nil {
			return nil, err
		}

		err = s.Save(keyType, account)
		if err != nil {
			return nil, err
		}

		return account, nil
	}

	return s.getAccount(ctx, keyType, effectiveAccountID)
}

// createAccount creates a new account.
func (s *AccountsStorage) createAccount(keyType certcrypto.KeyType, email, accountID string) (*Account, error) {
	effectiveAccountID := getEffectiveAccountID(email, accountID)

	privateKey, err := s.createPrivateKey(keyType, effectiveAccountID)
	if err != nil {
		return nil, err
	}

	return NewAccount(email, effectiveAccountID, keyType, privateKey), nil
}

// getAccount gets the account from a file.
// It will also try to recover the registration if it's missing (and save the account file).
// And it will also create a new private key if it doesn't exist (and save the private key file).
func (s *AccountsStorage) getAccount(ctx context.Context, keyType certcrypto.KeyType, effectiveAccountID string) (*Account, error) {
	accountFilePath := s.getAccountFilePath(keyType, effectiveAccountID)

	fileBytes, err := os.ReadFile(accountFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read the account file %q: %w", accountFilePath, err)
	}

	account := new(Account)

	err = json.Unmarshal(fileBytes, account)
	if err != nil {
		return nil, fmt.Errorf("could not parse the account file %s: %w", accountFilePath, err)
	}

	account.key, err = s.readPrivateKey(keyType, effectiveAccountID)
	if err == nil {
		if account.Registration != nil && account.Registration.Body.Status != "" {
			return account, nil
		}

		account.Registration, err = s.tryRecoverRegistration(ctx, account.key)
		if err != nil {
			return nil, fmt.Errorf("could not load the account file, registration is nil (accountID: %s): %w", effectiveAccountID, err)
		}

		err = s.Save(keyType, account)
		if err != nil {
			return nil, fmt.Errorf("could not save the account file, registration is nil (accountID: %s): %w", effectiveAccountID, err)
		}

		return account, nil
	}

	var privateKeyNotFound *PrivateKeyNotFound

	if !errors.As(err, &privateKeyNotFound) {
		return nil, err
	}

	// TODO(ldez): debug level?
	log.Info("No key found for the account. Generating a new private key.",
		slog.String("accountID", effectiveAccountID),
		slog.Any("keyType", keyType),
	)

	account.key, err = s.createPrivateKey(keyType, effectiveAccountID)
	if err != nil {
		return nil, fmt.Errorf("new private key creation: %w", err)
	}

	return account, nil
}

// createPrivateKey generates a new private key and saves it to a file.
func (s *AccountsStorage) createPrivateKey(keyType certcrypto.KeyType, effectiveAccountID string) (crypto.PrivateKey, error) {
	keysPath := s.getKeyPath(keyType, effectiveAccountID)

	accKeyPath := filepath.Join(keysPath, effectiveAccountID+".key")

	err := CreateNonExistingFolder(keysPath)
	if err != nil {
		return nil, fmt.Errorf("could not check/create the directory %q for the account (accountID: %s): %w", keysPath, effectiveAccountID, err)
	}

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, fmt.Errorf("private key generation (accountID: %s): %w", effectiveAccountID, err)
	}

	certOut, err := os.Create(accKeyPath)
	if err != nil {
		return nil, fmt.Errorf("private key file creation: (accountID: %s): %w", effectiveAccountID, err)
	}

	defer func() {
		_ = certOut.Close()

		// TODO(ldez): debug level?
		log.Info("Private key saved.", slog.String("filepath", accKeyPath))
	}()

	pemKey := certcrypto.PEMBlock(privateKey)

	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return nil, fmt.Errorf("private key PEM encoding: (accountID: %s): %w", effectiveAccountID, err)
	}

	return privateKey, nil
}

// readPrivateKey reads the private key from a file.
func (s *AccountsStorage) readPrivateKey(keyType certcrypto.KeyType, effectiveAccountID string) (crypto.PrivateKey, error) {
	keysPath := s.getKeyPath(keyType, effectiveAccountID)

	accKeyPath := filepath.Join(keysPath, effectiveAccountID+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		return nil, &PrivateKeyNotFound{AccountID: effectiveAccountID}
	} else if err != nil {
		return nil, err
	}

	privateKey, err := ReadPrivateKeyFile(accKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading the private key file %q: %w", accKeyPath, err)
	}

	return privateKey, nil
}

// existsAccountFile checks if the account file exists.
func (s *AccountsStorage) existsAccountFile(keyType certcrypto.KeyType, effectiveAccountID string) bool {
	accountFilePath := s.getAccountFilePath(keyType, effectiveAccountID)

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

// getAccountFilePath returns the account file path.
func (s *AccountsStorage) getAccountFilePath(keyType certcrypto.KeyType, effectiveAccountID string) string {
	return filepath.Join(s.getKeyPath(keyType, effectiveAccountID), accountFileName)
}

// getKeyPath returns the path to the folder that contains the private key for an account.
func (s *AccountsStorage) getKeyPath(keyType certcrypto.KeyType, effectiveAccountID string) string {
	return filepath.Join(s.getRootUserPath(effectiveAccountID), string(keyType))
}

// getRootUserPath returns the path to the root folder for an account.
func (s *AccountsStorage) getRootUserPath(effectiveAccountID string) string {
	return filepath.Join(s.rootPath, sanitizeHost(s.server), effectiveAccountID)
}

// tryRecoverRegistration tries to recover the registration from the private key.
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

// ReadPrivateKeyFile reads a private key file.
func ReadPrivateKeyFile(filename string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading the private key: %w", err)
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing the private key: %w", err)
	}

	return privateKey, nil
}

func sanitizeHost(uri *url.URL) string {
	return strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(uri.Host)
}
