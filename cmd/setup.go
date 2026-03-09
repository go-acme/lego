package cmd

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/urfave/cli/v3"
)

func newClient(cmd *cli.Command, account registration.User, keyType certcrypto.KeyType) (*lego.Client, error) {
	client, err := lego.NewClient(newClientConfig(cmd, account, keyType))
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	if client.GetServerMetadata().ExternalAccountRequired && !cmd.IsSet(flgEAB) {
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

	config.HTTPClient = internal.NewRetryableClient(config.HTTPClient)

	return config
}

func getUserAgent(cmd *cli.Command) string {
	return strings.TrimSpace(fmt.Sprintf("%s lego-cli/%s", cmd.String(flgUserAgent), cmd.Version))
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

func newHookManager(cmd *cli.Command, certsStorage *storage.CertificatesStorage, account *storage.Account) *hook.Manager {
	return hook.NewManager(
		certsStorage,
		hook.WithPre(cmd.String(flgPreHook), cmd.Duration(flgPreHookTimeout)),
		hook.WithDeploy(cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout)),
		hook.WithPost(cmd.String(flgPostHook), cmd.Duration(flgPostHookTimeout)),
		hook.WithAccountMetadata(account),
	)
}

func parseAddress(cmd *cli.Command, flgName string) (string, string, error) {
	address := cmd.String(flgName)

	if !strings.Contains(address, ":") {
		return "", "", fmt.Errorf("the flag '--%s' only accepts 'interface:port' or ':port' for its argument: '%s'",
			flgName, address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", "", fmt.Errorf("could not split address '%s': %w", address, err)
	}

	return host, port, nil
}
