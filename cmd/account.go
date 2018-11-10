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
	accKeysPath := AccountKeysPath(c, email)
	// TODO: move to function in configuration?
	accKeyPath := filepath.Join(accKeysPath, email+".key")
	if err := checkFolder(accKeysPath); err != nil {
		log.Fatalf("Could not check/create directory for account %s: %v", email, err)
	}

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

	accountFile := filepath.Join(AccountPath(c, email), "account.json")
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privKey}
	}

	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		log.Fatalf("Could not load file for account %s -> %v", email, err)
	}

	var acc Account
	err = json.Unmarshal(fileBytes, &acc)
	if err != nil {
		log.Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	acc.key = privKey

	if acc.Registration == nil || acc.Registration.Body.Status == "" {
		reg, err := tryRecoverAccount(privKey, c)
		if err != nil {
			log.Fatalf("Could not load account for %s. Registration is nil -> %#v", email, err)
		}

		acc.Registration = reg
		err = acc.Save(c)
		if err != nil {
			log.Fatalf("Could not save account for %s. Registration is nil -> %#v", email, err)
		}
	}

	return &acc
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
		filepath.Join(AccountPath(c, a.Email), "account.json"),
		jsonBytes,
		0600,
	)
}

// AccountKeysPath returns the OS dependent path to the keys of a particular account
func AccountKeysPath(c *cli.Context, acc string) string {
	return filepath.Join(AccountPath(c, acc), "keys")
}

// AccountPath returns the OS dependent path to a particular account
func AccountPath(c *cli.Context, acc string) string {
	return filepath.Join(accountsPath(c), acc)
}

// accountsPath returns the OS dependent path to the local accounts for a specific CA
func accountsPath(c *cli.Context) string {
	return filepath.Join(c.GlobalString("path"), "accounts", serverPath(c))
}

// serverPath returns the OS dependent path to the data for a specific CA
func serverPath(c *cli.Context) string {
	srv, _ := url.Parse(c.GlobalString("server"))
	return strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(srv.Host)
}
