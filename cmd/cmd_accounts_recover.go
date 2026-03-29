package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createRecover() *cli.Command {
	return &cli.Command{
		Name:   "recover",
		Usage:  "Recover/import an account from the private key.",
		Action: accountRecover,
		Flags:  flags.CreateRecoverFlags(),
	}
}

func accountRecover(ctx context.Context, cmd *cli.Command) error {
	log.Info("Recovering account.",
		slog.String("account", storage.GetEffectiveAccountID(cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))),
		slog.String("email", cmd.String(flags.FlgEmail)),
		slog.String("server", cmd.String(flags.FlgServer)),
	)

	if !confirm("Do you want to proceed?") {
		log.Info("Aborting.")
		return nil
	}

	accountsStorage := storage.NewAccountsStorage(cmd.String(flags.FlgPath))

	privateKey, err := storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	account, err := storage.NewRawAccount(cmd.String(flags.FlgAccountID), cmd.String(flags.FlgEmail), privateKey)
	if err != nil {
		return fmt.Errorf("raw account: %w", err)
	}

	account.Server = cmd.String(flags.FlgServer)
	account.NeedsRecovery = true

	client, err := newClient(cmd, account)
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	reg, err := client.Registration.ResolveAccountByKey(ctx)
	if err != nil {
		return fmt.Errorf("resolve account by key: %w", err)
	}

	account.Registration = reg

	err = accountsStorage.SavePrivateKey(account)
	if err != nil {
		return fmt.Errorf("could not save the private key file: %w", err)
	}

	err = accountsStorage.Save(account)
	if err != nil {
		return fmt.Errorf("could not save the account file: %w", err)
	}

	return nil
}
