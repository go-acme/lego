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

const userIDPlaceholder = "noemail@example.com"

const (
	baseAccountsRootFolderName = "accounts"
	baseKeysFolderName         = "keys"
	accountFileName            = "account.json"
)

type AccountsStorageConfig struct {
	Email    string
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
	userID          string
	email           string
	rootPath        string
	keysPath        string
	accountFilePath string

	server    *url.URL
	userAgent string
}

// NewAccountsStorage Creates a new AccountsStorage.
func NewAccountsStorage(config AccountsStorageConfig) (*AccountsStorage, error) {
	email := config.Email

	userID := email
	if userID == "" {
		userID = userIDPlaceholder
	}

	serverURL, err := url.Parse(config.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q: %w", config.Server, err)
	}

	rootPath := filepath.Join(config.BasePath, baseAccountsRootFolderName)
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(serverURL.Host)
	accountsPath := filepath.Join(rootPath, serverPath)
	rootUserPath := filepath.Join(accountsPath, userID)

	return &AccountsStorage{
		userID:          userID,
		email:           email,
		rootPath:        rootPath,
		keysPath:        filepath.Join(rootUserPath, baseKeysFolderName),
		accountFilePath: filepath.Join(rootUserPath, accountFileName),
		server:          serverURL,
		userAgent:       config.UserAgent,
	}, nil
}

func (s *AccountsStorage) ExistsAccountFilePath() bool {
	if _, err := os.Stat(s.accountFilePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatal("Could not read the account file.",
			slog.String("filepath", s.accountFilePath),
			log.ErrorAttr(err),
		)
	}

	return true
}

func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

func (s *AccountsStorage) GetUserID() string {
	return s.userID
}

func (s *AccountsStorage) GetEmail() string {
	return s.email
}

func (s *AccountsStorage) Save(account *Account) error {
	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(s.accountFilePath, jsonBytes, filePerm)
}

func (s *AccountsStorage) LoadAccount(ctx context.Context, privateKey crypto.PrivateKey) (*Account, error) {
	fileBytes, err := os.ReadFile(s.accountFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read the account file (userID: %s): %w", s.GetUserID(), err)
	}

	var account Account

	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		return nil, fmt.Errorf("could not parse the account file (userID: %s): %w", s.GetUserID(), err)
	}

	account.key = privateKey

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := s.tryRecoverRegistration(ctx, privateKey)
		if err != nil {
			return nil, fmt.Errorf("could not load the account file, registration is nil (userID: %s): %w", s.GetUserID(), err)
		}

		account.Registration = reg

		err = s.Save(&account)
		if err != nil {
			return nil, fmt.Errorf("could not save the account file, registration is nil (userID: %s): %w", s.GetUserID(), err)
		}
	}

	return &account, nil
}

func (s *AccountsStorage) GetPrivateKey(keyType certcrypto.KeyType) (crypto.PrivateKey, error) {
	accKeyPath := filepath.Join(s.keysPath, s.GetUserID()+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		// TODO(ldez): debug level?
		log.Info("No key found for the account. Generating a new private key.",
			slog.String("userID", s.GetUserID()),
			slog.Any("keyType", keyType),
		)

		err := CreateNonExistingFolder(s.keysPath)
		if err != nil {
			return nil, fmt.Errorf("could not check/create the directory %q for the account (userID: %s): %w", s.keysPath, s.GetUserID(), err)
		}

		privateKey, err := generatePrivateKey(accKeyPath, keyType)
		if err != nil {
			return nil, fmt.Errorf("could not generate the private account key (userID: %s): %w", s.GetUserID(), err)
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
