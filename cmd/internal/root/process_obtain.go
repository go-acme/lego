package root

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/cmd/internal"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
)

func obtain(ctx context.Context, cfg *configuration.Configuration) error {
	networkStack := getNetworkStack(cfg)

	store := storage.New(cfg.Storage)

	for accountID, challengesInfo := range createCertificatesMapping(cfg) {
		accountConfig := cfg.Accounts[accountID]

		keyType, err := certcrypto.ToKeyType(accountConfig.KeyType)
		if err != nil {
			return err
		}

		serverConfig := configuration.GetServerConfig(cfg, accountID)

		account, err := store.Account.Get(serverConfig.URL, keyType, accountConfig.Email, accountID)
		if err != nil {
			return err
		}

		lazyClient := sync.OnceValues(func() (*lego.Client, error) {
			client, errC := lego.NewClient(newClientConfig(serverConfig, account, cfg.UserAgent))
			if errC != nil {
				return nil, errC
			}

			if client.GetServerMetadata().ExternalAccountRequired && accountConfig.ExternalAccountBinding == nil {
				return nil, errors.New("server requires External Account Binding (EAB)")
			}

			return client, nil
		})

		err = handleRegistration(ctx, lazyClient, accountConfig, store.Account, account)
		if err != nil {
			return fmt.Errorf("registration: %w", err)
		}

		for challengeID, certIDs := range challengesInfo {
			chlgConfig := cfg.Challenges[challengeID]

			lazySetup := sync.OnceValues(func() (*lego.Client, error) {
				client, errC := lazyClient()
				if errC != nil {
					return nil, fmt.Errorf("set up client: %w", errC)
				}

				client.Challenge.RemoveAll()

				errC = setupChallenges(client, chlgConfig, networkStack)
				if errC != nil {
					return nil, fmt.Errorf("setup challenges: %w", errC)
				}

				return client, nil
			})

			for _, certID := range certIDs {
				certConfig := cfg.Certificates[certID]

				// Renew
				if store.Certificate.ExistsFile(certID, storage.ExtResource) {
					err = renewCertificate(ctx, lazyClient, certID, certConfig, store.Certificate)
					if err != nil {
						return err
					}

					continue
				}

				// Run
				err := runCertificate(ctx, lazySetup, certConfig, store.Certificate)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createCertificatesMapping creates a mapping of account -> challenge -> certificate IDs.
func createCertificatesMapping(cfg *configuration.Configuration) map[string]map[string][]string {
	// Accounts -> Challenges -> Certificates
	certsMappings := make(map[string]map[string][]string)

	for certID, certDesc := range cfg.Certificates {
		if _, ok := certsMappings[certDesc.Account]; !ok {
			certsMappings[certDesc.Account] = make(map[string][]string)
		}

		certsMappings[certDesc.Account][certDesc.Challenge] = append(certsMappings[certDesc.Account][certDesc.Challenge], certID)
	}

	return certsMappings
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
