package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/prompt"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-acme/lego/v5/registration/zerossl"
	"github.com/urfave/cli/v3"
)

func createAccountRegister() *cli.Command {
	return &cli.Command{
		Name:   "register",
		Usage:  "Register an account.",
		Action: register,
		Flags:  flags.CreateRegisterFlags(),
	}
}

func register(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return err
	}

	accountsStorage := storage.NewAccountsStorage(cmd.String(flags.FlgPath))

	account, err := accountsStorage.Get(cmd.String(flags.FlgServer), keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	lazyClient := sync.OnceValues(func() (*lego.Client, error) {
		return newClient(cmd, account)
	})

	err = handleRegistration(ctx, cmd, lazyClient, accountsStorage, account, true)
	if err != nil {
		return fmt.Errorf("registration: %w", err)
	}

	return nil
}

func handleRegistration(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, accountsStorage *storage.AccountsStorage, account *storage.Account, allowRegister bool) error {
	updateAccountOrigin(account)

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

			reg, err := registerAccount(ctx, cmd, client)
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

func registerAccount(ctx context.Context, cmd *cli.Command, client *lego.Client) (*acme.ExtendedAccount, error) {
	accepted := handleTOS(cmd, client)
	if !accepted {
		log.Fatal("You did not accept the TOS. Unable to proceed.")
	}

	if client.GetServerMetadata().ExternalAccountRequired && !cmd.IsSet(flags.FlgEAB) {
		return nil, errors.New("server requires External Account Binding (EAB)")
	}

	if cmd.Bool(flags.FlgEAB) {
		kid := cmd.String(flags.FlgEABKID)
		hmacEncoded := cmd.String(flags.FlgEABHMAC)

		if kid == "" || hmacEncoded == "" {
			log.Fatal(fmt.Sprintf("Requires arguments --%s and --%s.", flags.FlgEABKID, flags.FlgEABHMAC))
		}

		return client.Registration.RegisterWithExternalAccountBinding(ctx, registration.RegisterEABOptions{
			TermsOfServiceAgreed: accepted,
			Kid:                  kid,
			HmacEncoded:          hmacEncoded,
		})
	} else if zerossl.IsZeroSSL(cmd.String(flags.FlgServer)) {
		return registration.RegisterWithZeroSSL(ctx, client.Registration, cmd.String(flags.FlgEmail))
	}

	return client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
}

func handleTOS(cmd *cli.Command, client *lego.Client) bool {
	// metadata items are optional, and termsOfService too.
	urlTOS := client.GetServerMetadata().TermsOfService
	if urlTOS == "" {
		return true
	}

	// Check for a global accept override
	if cmd.Bool(flags.FlgAcceptTOS) {
		return true
	}

	log.Warn("Please review the TOS.", slog.String("url", urlTOS))

	return prompt.Confirm("Do you accept the TOS?")
}

func updateAccountOrigin(account *storage.Account) {
	if account.Origin == "" || account.Origin == storage.OriginMigration {
		account.Origin = storage.OriginCommand
	}
}
