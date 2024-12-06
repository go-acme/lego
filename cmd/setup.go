package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/urfave/cli/v2"
)

const filePerm os.FileMode = 0o600

// setupClient creates a new client with challenge settings.
func setupClient(ctx *cli.Context, account *Account, keyType certcrypto.KeyType) *lego.Client {
	client := newClient(ctx, account, keyType)

	setupChallenges(ctx, client)

	return client
}

func setupAccount(ctx *cli.Context, accountsStorage *AccountsStorage) (*Account, certcrypto.KeyType) {
	keyType := getKeyType(ctx)
	privateKey := accountsStorage.GetPrivateKey(keyType)

	var account *Account
	if accountsStorage.ExistsAccountFilePath() {
		account = accountsStorage.LoadAccount(privateKey)
	} else {
		account = &Account{Email: accountsStorage.GetUserID(), key: privateKey}
	}

	return account, keyType
}

func newClient(ctx *cli.Context, acc registration.User, keyType certcrypto.KeyType) *lego.Client {
	config := lego.NewConfig(acc)
	config.CADirURL = ctx.String(flgServer)

	config.Certificate = lego.CertificateConfig{
		KeyType:             keyType,
		Timeout:             time.Duration(ctx.Int(flgCertTimeout)) * time.Second,
		OverallRequestLimit: ctx.Int(flgOverallRequestLimit),
	}
	config.UserAgent = getUserAgent(ctx)

	if ctx.IsSet(flgHTTPTimeout) {
		config.HTTPClient.Timeout = time.Duration(ctx.Int(flgHTTPTimeout)) * time.Second
	}

	if ctx.Bool(flgTLSSkipVerify) {
		defaultTransport, ok := config.HTTPClient.Transport.(*http.Transport)
		if ok { // This is always true because the default client used by the CLI defined the transport.
			tr := defaultTransport.Clone()
			tr.TLSClientConfig.InsecureSkipVerify = true
			config.HTTPClient.Transport = tr
		}
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient = config.HTTPClient

	config.HTTPClient = retryClient.StandardClient()

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
	}

	if client.GetExternalAccountRequired() && !ctx.IsSet(flgEAB) {
		log.Fatalf("Server requires External Account Binding. Use --%s with --%s and --%s.", flgEAB, flgKID, flgHMAC)
	}

	return client
}

// getKeyType the type from which private keys should be generated.
func getKeyType(ctx *cli.Context) certcrypto.KeyType {
	keyType := ctx.String(flgKeyType)
	switch strings.ToUpper(keyType) {
	case "RSA2048":
		return certcrypto.RSA2048
	case "RSA3072":
		return certcrypto.RSA3072
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
	email := ctx.String(flgEmail)
	if email == "" {
		log.Fatalf("You have to pass an account (email address) to the program using --%s or -m", flgEmail)
	}
	return email
}

func getUserAgent(ctx *cli.Context) string {
	return strings.TrimSpace(fmt.Sprintf("%s lego-cli/%s", ctx.String(flgUserAgent), ctx.App.Version))
}

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o700)
	} else if err != nil {
		return err
	}
	return nil
}

func readCSRFile(filename string) (*x509.CertificateRequest, error) {
	bytes, err := os.ReadFile(filename)
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
		if p.Type == "CERTIFICATE REQUEST" || p.Type == "NEW CERTIFICATE REQUEST" {
			raw = p.Bytes
		}
	}

	// no PEM-encoded CSR
	// assume we were given a DER-encoded ASN.1 CSR
	// (if this assumption is wrong, parsing these bytes will fail)
	return x509.ParseCertificateRequest(raw)
}
