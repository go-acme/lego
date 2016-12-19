package main

import (
	"crypto"
	"encoding/json"
	"os"
	"path"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/storage"
)

// Account represents a users local saved credentials
type Account struct {
	Email        string `json:"email"`
	key          crypto.PrivateKey
	Registration *acme.RegistrationResource `json:"registration"`

	conf *Configuration
}

// NewAccount creates a new account for an email address
func NewAccount(email string, conf *Configuration, s storage.StorageProvider) *Account {
	accKeysPath := conf.AccountKeysPath(email)
	// TODO: move to function in configuration?
	accKeyPath := accKeysPath + string(os.PathSeparator) + email + ".key"
	if err := s.CheckPath(accKeysPath); err != nil {
		logger().Fatalf("Could not check/create directory for account %s: %v", email, err)
	}

	var privKey crypto.PrivateKey
	if _, err := s.ReadPath(accKeyPath); err != nil {
		logger().Printf("%v", err.Error())
		logger().Printf("No key found for account %s. Generating a curve P384 EC key.", email)
		privKey, err = generatePrivateKey(accKeyPath, s)
		if err != nil {
			logger().Fatalf("Could not generate RSA private account key for account %s: %v", email, err)
		}

		logger().Printf("Saved key to %s", accKeyPath)
	} else {
		privKey, err = loadPrivateKey(accKeyPath, s)
		if err != nil {
			logger().Fatalf("Could not load RSA private key from file %s: %v", accKeyPath, err)
		}
	}

	accountFile := path.Join(conf.AccountPath(email), "account.json")
	fileBytes, err := s.ReadPath(accountFile)
	if err != nil {
		return &Account{Email: email, key: privKey, conf: conf}
	}

	var acc Account
	err = json.Unmarshal(fileBytes, &acc)
	if err != nil {
		logger().Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	acc.key = privKey
	acc.conf = conf

	if acc.Registration == nil {
		logger().Fatalf("Could not load account for %s. Registration is nil.", email)
	}

	if acc.conf == nil {
		logger().Fatalf("Could not load account for %s. Configuration is nil.", email)
	}

	return &acc
}

/** Implementation of the acme.User interface **/

// GetEmail returns the email address for the account
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

/** End **/

// Save the account to disk
func (a *Account) Save(s storage.StorageProvider) error {
	jsonBytes, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}
	return s.WritePath(
		path.Join(a.conf.AccountPath(a.Email), "account.json"),
		jsonBytes)
}
