package root

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/cmd/internal"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/internal/dotenv"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
)

func Process(ctx context.Context, cfg *configuration.Configuration) error {
	archiver := storage.NewArchiver(cfg.Storage)

	err := archiver.Accounts(cfg)
	if err != nil {
		return err
	}

	err = archiver.Certificates(cfg.Certificates)
	if err != nil {
		return err
	}

	return process(ctx, cfg)
}

func process(ctx context.Context, cfg *configuration.Configuration) error {
	networkStack := getNetworkStack(cfg)

	store := storage.New(cfg.Storage)

	for _, accountNode := range configuration.LookupChallenges(cfg, nil) {
		account, err := store.Account.Get(accountNode.ServerConfig.URL, accountNode.KeyType, accountNode.Email, accountNode.ID)
		if err != nil {
			return err
		}

		lazyClient := sync.OnceValues(func() (*lego.Client, error) {
			return lego.NewClient(newClientConfig(accountNode.ServerConfig, account, cfg.UserAgent))
		})

		err = handleRegistration(ctx, lazyClient, accountNode.Account, store.Account, account, true)
		if err != nil {
			return fmt.Errorf("registration: %w", err)
		}

		hm := hook.NewManager(
			store.Certificate,
			withHooks(cfg.Hooks),
			hook.WithAccountMetadata(account),
		)

		for _, chlgNode := range accountNode.Children {
			// Clone the hook manager for each certificate because:
			// each certificate is different, so the metadata is different, except for the account information.
			hookManager := hm.Clone()

			err := processChallenges(ctx, lazyClient, chlgNode, store, hookManager, networkStack)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processChallenges(ctx context.Context, lazyClient lzSetUp, chlgNode *configuration.ChallengeNode, store *storage.Storage, hookManager *hook.Manager, networkStack challenge.NetworkStack) error {
	if chlgNode.DNS != nil {
		cleanUp, err := dotenv.Load(chlgNode.DNS.EnvFile)

		defer cleanUp()

		if err != nil {
			return fmt.Errorf("load environment variables: %w", err)
		}
	}

	lazySetup := sync.OnceValues(func() (*lego.Client, error) {
		client, errC := lazyClient()
		if errC != nil {
			return nil, fmt.Errorf("set up client: %w", errC)
		}

		client.Challenge.ResetSolvers()

		errC = setupChallenges(client, chlgNode.Challenge, networkStack)
		if errC != nil {
			return nil, fmt.Errorf("setup challenges: %w", errC)
		}

		return client, nil
	})

	for _, cert := range chlgNode.Certificates {
		// Renew
		if store.Certificate.ExistsFile(cert.ID, storage.ExtResource) {
			err := renew(ctx, lazySetup, cert.ID, cert, store.Certificate, hookManager)
			if err != nil {
				return err
			}

			continue
		}

		// Run
		err := obtain(ctx, lazySetup, cert.ID, cert, store.Certificate, hookManager)
		if err != nil {
			return err
		}
	}

	return nil
}

func getNetworkStack(cfg *configuration.Configuration) challenge.NetworkStack {
	switch cfg.NetworkStack {
	case "ipv4only", "ipv4":
		return challenge.IPv4Only

	case "ipv6only", "ipv6":
		return challenge.IPv6Only

	default:
		return challenge.DualStack
	}
}

func newClientConfig(serverConfig *configuration.Server, account registration.User, ua string) *lego.Config {
	config := lego.NewConfig(account)
	config.CADirURL = serverConfig.URL
	config.UserAgent = ua
	config.Certificate = lego.CertificateConfig{}

	if serverConfig.OverallRequestLimit > 0 {
		config.Certificate.OverallRequestLimit = serverConfig.OverallRequestLimit
	}

	if serverConfig.CertTimeout > 0 {
		config.Certificate.Timeout = time.Duration(serverConfig.CertTimeout) * time.Second
	}

	if serverConfig.HTTPTimeout > 0 {
		config.HTTPClient.Timeout = time.Duration(serverConfig.HTTPTimeout) * time.Second
	}

	if serverConfig.TLSSkipVerify {
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

// NOTE(ldez): this is partially a duplication with flags parsing, but the errors are slightly different.
func parseAddress(address string) (string, string, error) {
	if !strings.Contains(address, ":") {
		return "", "", fmt.Errorf("the address only accepts 'interface:port' or ':port' for its argument: '%s'",
			address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", "", fmt.Errorf("could not split address '%s': %w", address, err)
	}

	return host, port, nil
}

func withHooks(hooks *configuration.Hooks) hook.Option {
	if hooks == nil {
		return hook.Noop
	}

	return func(m *hook.Manager) {
		addHook(hooks.Pre, hook.WithPre)(m)
		addHook(hooks.Deploy, hook.WithDeploy)(m)
		addHook(hooks.Post, hook.WithPost)(m)
	}
}

func addHook(h *configuration.Hook, optFn func(cmd string, timeout time.Duration) hook.Option) hook.Option {
	if h == nil {
		return hook.Noop
	}

	return optFn(h.Cmd, h.Timeout)
}
