package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/registration"
	"github.com/urfave/cli"
)

const filePerm os.FileMode = 0600

func setup(ctx *cli.Context, accountsStorage *AccountsStorage) (*Account, *lego.Client) {
	keyType := getKeyType(ctx)
	privateKey := accountsStorage.GetPrivateKey(keyType)

	var account *Account
	if accountsStorage.ExistsAccountFilePath() {
		account = accountsStorage.LoadAccount(privateKey)
	} else {
		account = &Account{Email: accountsStorage.GetUserID(), key: privateKey}
	}

	client := newClient(ctx, account, keyType)

	return account, client
}

func newClient(ctx *cli.Context, acc registration.User, keyType certcrypto.KeyType) *lego.Client {
	config := lego.NewConfig(acc)
	config.CADirURL = ctx.GlobalString("server")

	config.Certificate = lego.CertificateConfig{
		KeyType: keyType,
		Timeout: time.Duration(ctx.GlobalInt("cert.timeout")) * time.Second,
	}
	config.UserAgent = fmt.Sprintf("lego-cli/%s", ctx.App.Version)

	if ctx.GlobalIsSet("http-timeout") {
		config.HTTPClient.Timeout = time.Duration(ctx.GlobalInt("http-timeout")) * time.Second
	}

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
	}

	if client.GetExternalAccountRequired() && !ctx.GlobalIsSet("eab") {
		log.Fatal("Server requires External Account Binding. Use --eab with --kid and --hmac.")
	}

	return client
}

// getKeyType the type from which private keys should be generated
func getKeyType(ctx *cli.Context) certcrypto.KeyType {
	keyType := ctx.GlobalString("key-type")
	switch strings.ToUpper(keyType) {
	case "RSA2048":
		return certcrypto.RSA2048
	case "RSA4096":
		return certcrypto.RSA4096
	case "RSA8192":
		return certcrypto.RSA8192
	case "EC256":
		return certcrypto.EC256
	case "EC384":
		return certcrypto.EC384
	}

	log.Fatalf("Unsupported KeyType: %s", keyType)
	return ""
}

func getEmail(ctx *cli.Context) string {
	email := ctx.GlobalString("email")
	if len(email) == 0 {
		log.Fatal("You have to pass an account (email address) to the program using --email or -m")
	}
	return email
}

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	} else if err != nil {
		return err
	}
	return nil
}

func readCSRFile(filename string) (*x509.CertificateRequest, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	raw := bytes

	// see if we can find a PEM-encoded CSR
	var p *pem.Block
	rest := bytes
	for {
		// decode a PEM block
		p, rest = pem.Decode(rest)

		// did we fail?
		if p == nil {
			break
		}

		// did we get a CSR?
		if p.Type == "CERTIFICATE REQUEST" {
			raw = p.Bytes
		}
	}

	// no PEM-encoded CSR
	// assume we were given a DER-encoded ASN.1 CSR
	// (if this assumption is wrong, parsing these bytes will fail)
	return x509.ParseCertificateRequest(raw)
}
