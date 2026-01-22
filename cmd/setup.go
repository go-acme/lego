package cmd

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/urfave/cli/v3"
)

// setupClient creates a new client with challenge settings.
func setupClient(cmd *cli.Command, account *Account, keyType certcrypto.KeyType) *lego.Client {
	client := newClient(cmd, account, keyType)

	setupChallenges(cmd, client)

	return client
}

func setupAccount(ctx context.Context, cmd *cli.Command, accountsStorage *AccountsStorage) (*Account, certcrypto.KeyType) {
	keyType := getKeyType(cmd)
	privateKey := accountsStorage.GetPrivateKey(keyType)

	var account *Account
	if accountsStorage.ExistsAccountFilePath() {
		account = accountsStorage.LoadAccount(ctx, privateKey)
	} else {
		account = &Account{Email: accountsStorage.GetEmail(), key: privateKey}
	}

	return account, keyType
}

func newClient(cmd *cli.Command, acc registration.User, keyType certcrypto.KeyType) *lego.Client {
	config := lego.NewConfig(acc)
	config.CADirURL = cmd.String(flgServer)

	config.Certificate = lego.CertificateConfig{
		KeyType:             keyType,
		Timeout:             time.Duration(cmd.Int(flgCertTimeout)) * time.Second,
		OverallRequestLimit: cmd.Int(flgOverallRequestLimit),
		EnableCommonName:    !cmd.Bool(flgDisableCommonName),
	}
	config.UserAgent = getUserAgent(cmd)

	if cmd.IsSet(flgHTTPTimeout) {
		config.HTTPClient.Timeout = time.Duration(cmd.Int(flgHTTPTimeout)) * time.Second
	}

	if cmd.Bool(flgTLSSkipVerify) {
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
	retryClient.CheckRetry = checkRetry
	retryClient.Logger = nil

	if _, v := os.LookupEnv("LEGO_DEBUG_ACME_HTTP_CLIENT"); v {
		retryClient.Logger = log.Default()
	}

	config.HTTPClient = retryClient.StandardClient()

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal("Could not create client.", log.ErrorAttr(err))
	}

	if client.GetExternalAccountRequired() && !cmd.IsSet(flgEAB) {
		log.Fatal(fmt.Sprintf("Server requires External Account Binding. Use --%s with --%s and --%s.", flgEAB, flgKID, flgHMAC))
	}

	return client
}

// getKeyType the type from which private keys should be generated.
func getKeyType(cmd *cli.Command) certcrypto.KeyType {
	keyType := cmd.String(flgKeyType)
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

	log.Fatal("Unsupported KeyType.", slog.String("keyType", keyType))

	return ""
}

func getUserAgent(cmd *cli.Command) string {
	return strings.TrimSpace(fmt.Sprintf("%s lego-cli/%s", cmd.String(flgUserAgent), cmd.Version))
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

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	rt, err := retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)
	if err != nil {
		return rt, err
	}

	if resp == nil {
		return rt, nil
	}

	if resp.StatusCode/100 == 2 {
		return rt, nil
	}

	all, err := io.ReadAll(resp.Body)
	if err == nil {
		var errorDetails *acme.ProblemDetails

		err = json.Unmarshal(all, &errorDetails)
		if err != nil {
			return rt, fmt.Errorf("%s %s: %s", resp.Request.Method, resp.Request.URL.Redacted(), string(all))
		}

		switch errorDetails.Type {
		case acme.BadNonceErr:
			return false, &acme.NonceError{
				ProblemDetails: errorDetails,
			}

		case acme.AlreadyReplacedErr:
			if errorDetails.HTTPStatus == http.StatusConflict {
				return false, &acme.AlreadyReplacedError{
					ProblemDetails: errorDetails,
				}
			}

		default:
			log.Warnf(log.LazySprintf("retry: %v", errorDetails))

			return rt, errorDetails
		}
	}

	return rt, nil
}
