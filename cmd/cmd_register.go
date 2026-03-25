package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-acme/lego/v5/registration/zerossl"
	"github.com/urfave/cli/v3"
)

func createRegister() *cli.Command {
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

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	if account.Registration == nil {
		client, err := newClient(cmd, account)
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}

		reg, err := registerAccount(ctx, cmd, client)
		if err != nil {
			return fmt.Errorf("could not complete registration: %w", err)
		}

		account.Registration = reg
		if err = accountsStorage.Save(keyType, account); err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}

		log.Warnf(log.LazySprintf(storage.RootPathWarningMessage, accountsStorage.GetRootPath()))
	} else {
		log.Info("Account already registered, skipping.")
	}

	return nil
}

func registerAccount(ctx context.Context, cmd *cli.Command, client *lego.Client) (*acme.ExtendedAccount, error) {
	accepted := handleTOS(cmd, client)
	if !accepted {
		log.Fatal("You did not accept the TOS. Unable to proceed.")
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

	reader := bufio.NewReader(os.Stdin)

	log.Warn("Please review the TOS", slog.String("url", urlTOS))

	for {
		fmt.Println("Do you accept the TOS? Y/n")

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read from the console", log.ErrorAttr(err))
		}

		text = strings.Trim(text, "\r\n")
		switch text {
		case "", "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("Your input was invalid. Please answer with one of Y/y, n/N or by pressing enter.")
		}
	}
}
