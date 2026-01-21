package cmd

import (
	"context"
	"crypto"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/urfave/cli/v3"
)

const userIDPlaceholder = "noemail@example.com"

const (
	baseAccountsRootFolderName = "accounts"
	baseKeysFolderName         = "keys"
	accountFileName            = "account.json"
)

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
	rootUserPath    string
	keysPath        string
	accountFilePath string
	cmd             *cli.Command
}

// NewAccountsStorage Creates a new AccountsStorage.
func NewAccountsStorage(cmd *cli.Command) *AccountsStorage {
	// TODO: move to account struct?
	email := cmd.String(flgEmail)

	userID := email
	if userID == "" {
		userID = userIDPlaceholder
	}

	serverURL, err := url.Parse(cmd.String(flgServer))
	if err != nil {
		log.Fatal("URL parsing", "flag", flgServer, "serverURL", cmd.String(flgServer), "error", err)
	}

	rootPath := filepath.Join(cmd.String(flgPath), baseAccountsRootFolderName)
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(serverURL.Host)
	accountsPath := filepath.Join(rootPath, serverPath)
	rootUserPath := filepath.Join(accountsPath, userID)

	return &AccountsStorage{
		userID:          userID,
		email:           email,
		rootPath:        rootPath,
		rootUserPath:    rootUserPath,
		keysPath:        filepath.Join(rootUserPath, baseKeysFolderName),
		accountFilePath: filepath.Join(rootUserPath, accountFileName),
		cmd:             cmd,
	}
}

func (s *AccountsStorage) ExistsAccountFilePath() bool {
	accountFile := filepath.Join(s.rootUserPath, accountFileName)
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatal("Could not read the account file.", "filepath", accountFile, "error", err)
	}

	return true
}

func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

func (s *AccountsStorage) GetRootUserPath() string {
	return s.rootUserPath
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

func (s *AccountsStorage) LoadAccount(ctx context.Context, privateKey crypto.PrivateKey) *Account {
	fileBytes, err := os.ReadFile(s.accountFilePath)
	if err != nil {
		log.Fatal("Could not load the account file.", "userID", s.GetUserID(), "error", err)
	}

	var account Account

	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		log.Fatal("Could not parse the account file.", "userID", s.GetUserID(), "error", err)
	}

	account.key = privateKey

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := tryRecoverRegistration(ctx, s.cmd, privateKey)
		if err != nil {
			log.Fatal("Could not load the account file. Registration is nil.", "userID", s.GetUserID(), "error", err)
		}

		account.Registration = reg

		err = s.Save(&account)
		if err != nil {
			log.Fatal("Could not save the account file. Registration is nil.", "userID", s.GetUserID(), "error", err)
		}
	}

	return &account
}

func (s *AccountsStorage) GetPrivateKey(keyType certcrypto.KeyType) crypto.PrivateKey {
	accKeyPath := filepath.Join(s.keysPath, s.GetUserID()+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		log.Info("No key found for the account. Generating a new private key.", "userID", s.GetUserID(), "keyType", keyType)
		s.createKeysFolder()

		privateKey, err := generatePrivateKey(accKeyPath, keyType)
		if err != nil {
			log.Fatal("Could not generate the RSA private account key.", "userID", s.GetUserID(), "error", err)
		}

		log.Info("Saved key.", "filepath", accKeyPath)

		return privateKey
	}

	privateKey, err := loadPrivateKey(accKeyPath)
	if err != nil {
		log.Fatal("Could not load an RSA private key from the file.", "filepath", accKeyPath, "error", err)
	}

	return privateKey
}

func (s *AccountsStorage) createKeysFolder() {
	if err := createNonExistingFolder(s.keysPath); err != nil {
		log.Fatal("Could not check/create the directory for the account.", "userID", s.GetUserID(), "error", err)
	}
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

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
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

func tryRecoverRegistration(ctx context.Context, cmd *cli.Command, privateKey crypto.PrivateKey) (*registration.Resource, error) {
	// couldn't load account but got a key. Try to look the account up.
	config := lego.NewConfig(&Account{key: privateKey})
	config.CADirURL = cmd.String(flgServer)
	config.UserAgent = getUserAgent(cmd)

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
