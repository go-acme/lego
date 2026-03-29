package cmd

import (
	"context"
	"crypto"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createKeyRollover() *cli.Command {
	return &cli.Command{
		Name:   "keyrollover",
		Usage:  "Update the account private key.",
		Action: accountKeyRollover,
		Flags:  flags.CreateKeyRolloverFlags(),
	}
}

func accountKeyRollover(ctx context.Context, cmd *cli.Command) error {
	log.Info("Account private key rollover.",
		slog.String("account", storage.GetEffectiveAccountID(cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))),
		slog.String("email", cmd.String(flags.FlgEmail)),
		slog.String("server", cmd.String(flags.FlgServer)),
		slog.String("keyType", cmd.String(flags.FlgKeyType)),
	)

	if !confirm("Do you want to proceed?") {
		log.Info("Aborting.")
		return nil
	}

	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return err
	}

	accountsStorage := storage.NewAccountsStorage(cmd.String(flags.FlgPath))

	account, err := accountsStorage.Get(cmd.String(flags.FlgServer), keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	var privateKey crypto.Signer

	privateKey, keyType, err = getPrivateKey(cmd, keyType)
	if err != nil {
		return fmt.Errorf("get private key: %w", err)
	}

	client, err := newClient(cmd, account)
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	err = client.Registration.KeyRollover(ctx, privateKey)
	if err != nil {
		return fmt.Errorf("could not complete key rollover: %w", err)
	}

	account.SetPrivateKey(privateKey)

	err = accountsStorage.SavePrivateKey(account)
	if err != nil {
		return fmt.Errorf("could not save the private key file: %w", err)
	}

	if keyType != "" && account.GetKeyType() != keyType {
		account.KeyType = keyType

		err = accountsStorage.Save(account)
		if err != nil {
			return fmt.Errorf("could not save the account file: %w", err)
		}
	}

	return nil
}

func getPrivateKey(cmd *cli.Command, keyType certcrypto.KeyType) (crypto.Signer, certcrypto.KeyType, error) {
	if cmd.IsSet(flags.FlgPrivateKey) {
		privateKey, err := storage.ReadPrivateKeyFile(cmd.String(flags.FlgPrivateKey))
		if err != nil {
			return nil, "", fmt.Errorf("load private key: %w", err)
		}

		kt, err := certcrypto.GetPrivateKeyType(privateKey)
		if err != nil {
			return nil, "", fmt.Errorf("get private key type: %w", err)
		}

		return privateKey, kt, nil
	}

	log.Debug("Generating a new private key.")

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, "", fmt.Errorf("generate a new private key: %w", err)
	}

	return privateKey, "", nil
}
