package root

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/prompt"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-acme/lego/v5/registration/zerossl"
)

func handleRegistration(ctx context.Context, lazyClient lzSetUp, accountConfig *configuration.Account, accountsStorage *storage.AccountsStorage, account *storage.Account, allowRegister bool) error {
	err := updateAccountOrigin(accountsStorage, account)
	if err != nil {
		return err
	}

	if account.NeedsRecovery {
		client, err := lazyClient()
		if err != nil {
			return fmt.Errorf("set up client: %w", err)
		}

		reg, err := client.Registration.ResolveAccountByKey(ctx)
		if err != nil {
			return fmt.Errorf("resolve account by key: %w", err)
		}

		account.Registration = reg

		err = accountsStorage.Save(account)
		if err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}

		return nil
	}

	if allowRegister {
		if account.Registration == nil {
			client, err := lazyClient()
			if err != nil {
				return fmt.Errorf("set up client: %w", err)
			}

			reg, err := registerAccount(ctx, client, accountConfig)
			if err != nil {
				return fmt.Errorf("could not complete registration: %w", err)
			}

			account.Registration = reg

			if err = accountsStorage.Save(account); err != nil {
				return fmt.Errorf("could not save the account file: %w", err)
			}

			log.Warnf(log.LazySprintf(storage.RootPathWarningMessage, accountsStorage.GetRootPath()))
		} else {
			log.Debug("Account already registered, skipping.", slog.String("account", account.GetID()))
		}
	} else if account.Registration == nil {
		return fmt.Errorf("the account %s is not registered", account.GetID())
	}

	return nil
}

func registerAccount(ctx context.Context, client *lego.Client, accountConfig *configuration.Account) (*acme.ExtendedAccount, error) {
	accepted := handleTOS(client, accountConfig)
	if !accepted {
		return nil, errors.New("you did not accept the TOS: unable to proceed")
	}

	if client.GetServerMetadata().ExternalAccountRequired && accountConfig.ExternalAccountBinding == nil {
		return nil, errors.New("server requires External Account Binding (EAB)")
	}

	if accountConfig.ExternalAccountBinding != nil {
		return client.Registration.RegisterWithExternalAccountBinding(ctx, registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  accountConfig.ExternalAccountBinding.KID,
			HmacEncoded:          accountConfig.ExternalAccountBinding.HmacKey,
		})
	} else if zerossl.IsZeroSSL(accountConfig.Server) {
		return registration.RegisterWithZeroSSL(ctx, client.Registration, accountConfig.Email)
	}

	return client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
}

func handleTOS(client *lego.Client, accountConfig *configuration.Account) bool {
	// metadata items are optional, and termsOfService too.
	urlTOS := client.GetServerMetadata().TermsOfService
	if urlTOS == "" {
		return true
	}

	// Check for a global acceptance override
	if accountConfig.AcceptsTermsOfService {
		return true
	}

	log.Warn("Please review the TOS", slog.String("url", urlTOS))

	return prompt.Confirm("Do you accept the TOS?")
}

func updateAccountOrigin(accountsStorage *storage.AccountsStorage, account *storage.Account) error {
	if account.Origin != storage.OriginConfiguration {
		account.Origin = storage.OriginConfiguration

		err := accountsStorage.Save(account)
		if err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}
	}

	return nil
}
