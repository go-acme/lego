package cmd

import (
	"crypto"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
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
	accKeyPath := getAccountKeyPath(c, email)

	var privKey crypto.PrivateKey
	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {

		log.Printf("No key found for account %s. Generating a curve P384 EC key.", email)
		privKey, err = generatePrivateKey(accKeyPath)
		if err != nil {
			log.Fatalf("Could not generate RSA private account key for account %s: %v", email, err)
		}

		log.Printf("Saved key to %s", accKeyPath)
	} else {
		privKey, err = loadPrivateKey(accKeyPath)
		if err != nil {
			log.Fatalf("Could not load RSA private key from file %s: %v", accKeyPath, err)
		}
	}

	accountFile := filepath.Join(getAccountPath(c, email), "account.json")
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privKey}
	}

	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		log.Fatalf("Could not load file for account %s -> %v", email, err)
	}

	var account Account
	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		log.Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	account.key = privKey

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := tryRecoverAccount(privKey, c)
		if err != nil {
			log.Fatalf("Could not load account for %s. Registration is nil -> %#v", email, err)
		}

		account.Registration = reg
		err = account.Save(c)
		if err != nil {
			log.Fatalf("Could not save account for %s. Registration is nil -> %#v", email, err)
		}
	}

	return &account
}

func tryRecoverAccount(privKey crypto.PrivateKey, c *cli.Context) (*registration.Resource, error) {
	// couldn't load account but got a key. Try to look the account up.
	config := acme.NewDefaultConfig(&Account{key: privKey}).
		WithCADirURL(c.GlobalString("server"))

	client, err := acme.NewClient(config)
	if err != nil {
		return nil, err
	}

	reg, err := client.Registration.ResolveAccountByKey()
	if err != nil {
		return nil, err
	}
	return reg, nil
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

	return ioutil.WriteFile(
		filepath.Join(getAccountPath(c, a.Email), "account.json"),
		jsonBytes,
		0600,
	)
}

// GetAccountPath returns the OS dependent path to a particular account
func (a *Account) GetAccountPath(c *cli.Context) string {
	return getAccountPath(c, a.Email)
}

func getAccountKeyPath(c *cli.Context, email string) string {
	accKeysPath := filepath.Join(getAccountPath(c, email), "keys")
	if err := checkFolder(accKeysPath); err != nil {
		log.Fatalf("Could not check/create directory for account %s: %v", email, err)
	}
	return filepath.Join(accKeysPath, email+".key")
}

// getAccountPath returns the OS dependent path to a particular account
func getAccountPath(c *cli.Context, acc string) string {
	srv, _ := url.Parse(c.GlobalString("server"))
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(srv.Host)

	accountsPath := filepath.Join(c.GlobalString("path"), "accounts", serverPath)

	return filepath.Join(accountsPath, acc)
}
