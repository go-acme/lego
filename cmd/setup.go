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
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/urfave/cli/v3"
)

type lzSetUp func() (*lego.Client, error)

func newClient(cmd *cli.Command, account registration.User) (*lego.Client, error) {
	client, err := lego.NewClient(newClientConfig(cmd, account))
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	if client.GetServerMetadata().ExternalAccountRequired && !cmd.IsSet(flags.FlgEAB) {
		return nil, errors.New("server requires External Account Binding (EAB)")
	}

	return client, nil
}

func newClientConfig(cmd *cli.Command, account registration.User) *lego.Config {
	config := lego.NewConfig(account)
	config.CADirURL = cmd.String(flags.FlgServer)
	config.UserAgent = getUserAgentFromFlag(cmd)

	config.Certificate = lego.CertificateConfig{
		Timeout:             time.Duration(cmd.Int(flags.FlgCertTimeout)) * time.Second,
		OverallRequestLimit: cmd.Int(flags.FlgOverallRequestLimit),
	}

	if cmd.IsSet(flags.FlgHTTPTimeout) {
		config.HTTPClient.Timeout = time.Duration(cmd.Int(flags.FlgHTTPTimeout)) * time.Second
	}

	if cmd.Bool(flags.FlgTLSSkipVerify) {
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

func getUserAgentFromFlag(cmd *cli.Command) string {
	return getUserAgent(cmd, cmd.String(flags.FlgUserAgent))
}

func getUserAgent(cmd *cli.Command, ua string) string {
	return strings.TrimSpace(fmt.Sprintf("%s lego-cli/%s", ua, cmd.Version))
}

func newObtainRequest(cmd *cli.Command, domains []string) (certificate.ObtainRequest, error) {
	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return certificate.ObtainRequest{}, err
	}

	return certificate.ObtainRequest{
		Domains:                        domains,
		MustStaple:                     cmd.Bool(flags.FlgMustStaple),
		KeyType:                        keyType,
		NotBefore:                      cmd.Timestamp(flags.FlgNotBefore),
		NotAfter:                       cmd.Timestamp(flags.FlgNotAfter),
		Bundle:                         !cmd.Bool(flags.FlgNoBundle),
		PreferredChain:                 cmd.String(flags.FlgPreferredChain),
		EnableCommonName:               cmd.Bool(flags.FlgEnableCommonName),
		Profile:                        cmd.String(flags.FlgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flags.FlgAlwaysDeactivateAuthorizations),
	}, nil
}

func newObtainForCSRRequest(cmd *cli.Command, csr *x509.CertificateRequest) certificate.ObtainForCSRRequest {
	return certificate.ObtainForCSRRequest{
		CSR:                            csr,
		NotBefore:                      cmd.Timestamp(flags.FlgNotBefore),
		NotAfter:                       cmd.Timestamp(flags.FlgNotAfter),
		Bundle:                         !cmd.Bool(flags.FlgNoBundle),
		PreferredChain:                 cmd.String(flags.FlgPreferredChain),
		EnableCommonName:               cmd.Bool(flags.FlgEnableCommonName),
		Profile:                        cmd.String(flags.FlgProfile),
		AlwaysDeactivateAuthorizations: cmd.Bool(flags.FlgAlwaysDeactivateAuthorizations),
	}
}

func newSaveOptions(cmd *cli.Command) *storage.SaveOptions {
	return &storage.SaveOptions{
		PEM: cmd.Bool(flags.FlgPEM),

		PFX:         cmd.Bool(flags.FlgPFX),
		PFXPassword: cmd.String(flags.FlgPFXPass),
		PFXFormat:   cmd.String(flags.FlgPFXFormat),
	}
}

func newHookManager(cmd *cli.Command, certsStorage *storage.CertificatesStorage, account *storage.Account) *hook.Manager {
	return hook.NewManager(
		certsStorage,
		hook.WithPre(cmd.String(flags.FlgPreHook), cmd.Duration(flags.FlgPreHookTimeout)),
		hook.WithDeploy(cmd.String(flags.FlgDeployHook), cmd.Duration(flags.FlgDeployHookTimeout)),
		hook.WithPost(cmd.String(flags.FlgPostHook), cmd.Duration(flags.FlgPostHookTimeout)),
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
