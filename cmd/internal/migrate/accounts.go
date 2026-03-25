package migrate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

type OldAccount struct {
	Email        string       `json:"email"`
	Registration *OldResource `json:"registration"`
}

type OldResource struct {
	Body acme.Account `json:"body"`
	URI  string       `json:"uri,omitempty"`
}

// Accounts migrates the old accounts directory structure to the new one.
func Accounts(root string) error {
	matches, err := zglob.Glob(filepath.Join(root, "**", "account.json"))
	if err != nil {
		return err
	}

	for _, srcAccountFilePath := range matches {
		log.Info("Migrating an account file.", slog.String("filepath", srcAccountFilePath))

		data, err := os.ReadFile(srcAccountFilePath)
		if err != nil {
			return fmt.Errorf("could not read the account file %q: %w", srcAccountFilePath, err)
		}

		var oldAccount OldAccount

		err = json.Unmarshal(data, &oldAccount)
		if err != nil {
			return fmt.Errorf("could not parse the account file %q: %w", srcAccountFilePath, err)
		}

		account := storage.Account{
			ID:    oldAccount.Email,
			Email: oldAccount.Email,
			Registration: &acme.ExtendedAccount{
				Account:  oldAccount.Registration.Body,
				Location: oldAccount.Registration.URI,
			},
		}

		accountDir := filepath.Dir(srcAccountFilePath)

		if account.ID == "" {
			account.ID = filepath.Base(accountDir)
		}

		srcKeyPath := filepath.Join(accountDir, "keys", account.GetID()+storage.ExtKey)

		account.KeyType, err = getKeyType(srcKeyPath)
		if err != nil {
			return fmt.Errorf("could not guess the account key type: %w", err)
		}

		// Move the private key file.

		dstKeyPath := filepath.Join(accountDir, account.GetID()+storage.ExtKey)

		err = os.Rename(srcKeyPath, dstKeyPath)
		if err != nil {
			return fmt.Errorf("could not rename the private key file %q to %q: %w", srcKeyPath, dstKeyPath, err)
		}

		err = os.RemoveAll(filepath.Join(accountDir, "keys"))
		if err != nil {
			return fmt.Errorf("could not remove the old keys directory: %w", err)
		}

		// Create the new account file.

		newAccountFile, err := os.Create(filepath.Join(accountDir, "account.json"))
		if err != nil {
			return fmt.Errorf("could not create the new account file: %w", err)
		}

		encoder := json.NewEncoder(newAccountFile)
		encoder.SetIndent("", "  ")

		err = encoder.Encode(account)
		if err != nil {
			return fmt.Errorf("could not encode the new account file: %w", err)
		}
	}

	return nil
}

func getKeyType(srcKeyPath string) (certcrypto.KeyType, error) {
	pk, err := storage.ReadPrivateKeyFile(srcKeyPath)
	if err != nil {
		return "", fmt.Errorf("could not read the private key file %q: %w", srcKeyPath, err)
	}

	kt, err := certcrypto.GetPrivateKeyType(pk)
	if err != nil {
		return "", fmt.Errorf("could not get the private key type: %w", err)
	}

	return kt, nil
}
