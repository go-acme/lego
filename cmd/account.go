package cmd

import (
	"crypto"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/registration"
)

// Account represents a users local saved credentials
type Account struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	key          crypto.PrivateKey
}

// NewAccount creates a new account for an email address
func NewAccount(c *cli.Context, email string) *Account {
	privateKey := getPrivateKey(c, email)

	accountFile := filepath.Join(getAccountsRootPath(c, email), accountFileName)
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privateKey}
	}

	return loadAccountFromFile(c, email, accountFile, privateKey)
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
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}

/** End **/

// Save the account to disk
func (a *Account) Save(c *cli.Context) error {
	jsonBytes, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}

	accountPath := filepath.Join(getAccountsRootPath(c, a.Email), accountFileName)
	return ioutil.WriteFile(accountPath, jsonBytes, filePerm)
}

// GetAccountPath returns the OS dependent path to a particular account
func (a *Account) GetAccountPath(c *cli.Context) string {
	return getAccountsRootPath(c, a.Email)
}
