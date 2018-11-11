package cmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
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

const (
	baseAccountFolderName = "accounts"
	baseKeysFolderName    = "keys"
	accountFileName       = "account.json"
)

func getOrCreateAccountKeysFolder(c *cli.Context, email string) string {
	accKeysPath := filepath.Join(getAccountPath(c, email), baseKeysFolderName)
	if err := createNonExistingFolder(accKeysPath); err != nil {
		log.Fatalf("Could not check/create directory for account %s: %v", email, err)
	}
	return accKeysPath
}

func loadAccountFromFile(c *cli.Context, email, accountFile string, privateKey crypto.PrivateKey) *Account {
	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		log.Fatalf("Could not load file for account %s -> %v", email, err)
	}

	var account Account
	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		log.Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	account.key = privateKey

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := tryRecoverAccount(privateKey, c)
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

func tryRecoverAccount(privateKey crypto.PrivateKey, c *cli.Context) (*registration.Resource, error) {
	// couldn't load account but got a key. Try to look the account up.
	config := acme.NewDefaultConfig(&Account{key: privateKey}).
		WithCADirURL(c.GlobalString("server")).
		WithUserAgent(fmt.Sprintf("lego-cli/%s", c.App.Version))

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

func getPrivateKey(c *cli.Context, email string) crypto.PrivateKey {
	accKeysPath := getOrCreateAccountKeysFolder(c, email)
	accKeyPath := filepath.Join(accKeysPath, email+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		log.Printf("No key found for account %s. Generating a curve P384 EC key.", email)

		privKey, err := generatePrivateKey(accKeyPath)
		if err != nil {
			log.Fatalf("Could not generate RSA private account key for account %s: %v", email, err)
		}

		log.Printf("Saved key to %s", accKeyPath)
		return privKey
	}

	privKey, err := loadPrivateKey(accKeyPath)
	if err != nil {
		log.Fatalf("Could not load RSA private key from file %s: %v", accKeyPath, err)
	}
	return privKey
}

func generatePrivateKey(file string) (crypto.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	pemKey := pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pemKey)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private key type")
}

// getAccountPath returns the OS dependent path to a particular account
func getAccountPath(c *cli.Context, acc string) string {
	srv, _ := url.Parse(c.GlobalString("server"))
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(srv.Host)

	accountsPath := filepath.Join(c.GlobalString("path"), baseAccountFolderName, serverPath)

	return filepath.Join(accountsPath, acc)
}
