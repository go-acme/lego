package cmd

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/urfave/cli/v3"
)

func newClient(cmd *cli.Command, account registration.User, keyType certcrypto.KeyType) (*lego.Client, error) {
	client, err := lego.NewClient(newClientConfig(cmd, account, keyType))
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	if client.GetExternalAccountRequired() && !cmd.IsSet(flgEAB) { // TODO(ldez): handle this flag.
		return nil, errors.New("server requires External Account Binding (EAB)")
	}

	return client, nil
}

func newClientConfig(cmd *cli.Command, account registration.User, keyType certcrypto.KeyType) *lego.Config {
	config := lego.NewConfig(account)
	config.CADirURL = cmd.String(flgServer)
	config.UserAgent = getUserAgent(cmd)

	config.Certificate = lego.CertificateConfig{
		KeyType:             keyType,
		Timeout:             time.Duration(cmd.Int(flgCertTimeout)) * time.Second,
		OverallRequestLimit: cmd.Int(flgOverallRequestLimit),
	}

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

	return config
}

func getUserAgent(cmd *cli.Command) string {
	return strings.TrimSpace(fmt.Sprintf("%s lego-cli/%s", cmd.String(flgUserAgent), cmd.Version))
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

func newObtainRequest(cmd *cli.Command, domains []string) certificate.ObtainRequest {
	return certificate.ObtainRequest{
		Domains:                        domains,
		MustStaple:                     cmd.Bool(flgMustStaple),
		NotBefore:                      cmd.Timestamp(flgNotBefore),
		NotAfter:                       cmd.Timestamp(flgNotAfter),
		Bundle:                         !cmd.Bool(flgNoBundle),
		PreferredChain:                 cmd.String(flgPreferredChain),
		EnableCommonName:               cmd.Bool(flgEnableCommonName),
		Profile:                        cmd.String(flgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flgAlwaysDeactivateAuthorizations),
	}
}

func newObtainForCSRRequest(cmd *cli.Command, csr *x509.CertificateRequest) certificate.ObtainForCSRRequest {
	return certificate.ObtainForCSRRequest{
		CSR:                            csr,
		NotBefore:                      cmd.Timestamp(flgNotBefore),
		NotAfter:                       cmd.Timestamp(flgNotAfter),
		Bundle:                         !cmd.Bool(flgNoBundle),
		PreferredChain:                 cmd.String(flgPreferredChain),
		EnableCommonName:               cmd.Bool(flgEnableCommonName),
		Profile:                        cmd.String(flgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flgAlwaysDeactivateAuthorizations),
	}
}

func validateNetworkStack(cmd *cli.Command) error {
	if cmd.Bool(flgIPv4Only) && cmd.Bool(flgIPv6Only) {
		return fmt.Errorf("cannot specify both --%s and --%s", flgIPv4Only, flgIPv6Only)
	}

	return nil
}

func newAccountsStorageConfig(cmd *cli.Command) storage.AccountsStorageConfig {
	return storage.AccountsStorageConfig{
		BasePath:  cmd.String(flgPath),
		Server:    cmd.String(flgServer),
		UserAgent: getUserAgent(cmd),
	}
}

func newSaveOptions(cmd *cli.Command) *storage.SaveOptions {
	return &storage.SaveOptions{
		PEM:         cmd.Bool(flgPEM),
		PFX:         cmd.Bool(flgPFX),
		PFXFormat:   cmd.String(flgPFXPass),
		PFXPassword: cmd.String(flgPFXFormat),
	}
}
