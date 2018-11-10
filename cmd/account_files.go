package cmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/log"
)

const (
	baseAccountFolderName = "accounts"
	baseKeysFolderName    = "keys"
	accountFileName       = "account.json"
)

// getAccountPath returns the OS dependent path to a particular account
func getAccountPath(c *cli.Context, acc string) string {
	srv, _ := url.Parse(c.GlobalString("server"))
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(srv.Host)

	accountsPath := filepath.Join(c.GlobalString("path"), baseAccountFolderName, serverPath)

	return filepath.Join(accountsPath, acc)
}

func getOrCreateAccountKeysFolder(c *cli.Context, email string) string {
	accKeysPath := filepath.Join(getAccountPath(c, email), baseKeysFolderName)
	if err := createNonExistingFolder(accKeysPath); err != nil {
		log.Fatalf("Could not check/create directory for account %s: %v", email, err)
	}
	return accKeysPath
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
