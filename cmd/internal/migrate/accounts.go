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

		if account.ID == "" {
			account.ID = filepath.Base(filepath.Dir(srcAccountFilePath))
		}

		accountsDir := filepath.Dir(srcAccountFilePath)

		srcKeyPath := filepath.Join(accountsDir, "keys", account.GetID()+storage.ExtKey)

		account.KeyType, err = getKeyType(srcKeyPath)
		if err != nil {
			return fmt.Errorf("could not guess the account key type: %w", err)
		}

		newAccountDir := filepath.Join(accountsDir, string(account.GetKeyType()))

		err = os.MkdirAll(newAccountDir, 0o700)
		if err != nil {
			return fmt.Errorf("could not create the directory %q: %w", newAccountDir, err)
		}

		// Rename the private key file.

		dstKeyPath := filepath.Join(newAccountDir, account.GetID()+storage.ExtKey)

		err = os.Rename(srcKeyPath, dstKeyPath)
		if err != nil {
			return fmt.Errorf("could not rename the private key file %q to %q: %w", srcKeyPath, dstKeyPath, err)
		}

		// Create the new account file.

		newAccountFile, err := os.Create(filepath.Join(newAccountDir, "account.json"))
		if err != nil {
			return fmt.Errorf("could not create the new account file: %w", err)
		}

		encoder := json.NewEncoder(newAccountFile)
		encoder.SetIndent("", "  ")

		err = encoder.Encode(account)
		if err != nil {
			return fmt.Errorf("could not encode the new account file: %w", err)
		}

		// Clean up.

		err = accountsCleanUp(srcAccountFilePath)
		if err != nil {
			return fmt.Errorf("could not clean up: %w", err)
		}
	}

	return nil
}

func accountsCleanUp(srcAccountPath string) error {
	err := os.Remove(srcAccountPath)
	if err != nil {
		return fmt.Errorf("could not remove the account file %q: %w", srcAccountPath, err)
	}

	err = os.Remove(filepath.Join(filepath.Dir(srcAccountPath), "keys"))
	if err != nil {
		return fmt.Errorf("could not remove the keys directory: %w", err)
	}

	return nil
}

func getKeyType(srcKeyPath string) (certcrypto.KeyType, error) {
	pk, err := storage.ReadPrivateKeyFile(srcKeyPath)
	if err != nil {
		return "", fmt.Errorf("could not read the private key file %q: %w", srcKeyPath, err)
	}

	kt, err := guessPrivateKeyType(pk)
	if err != nil {
		return "", fmt.Errorf("could not guess the private key type: %w", err)
	}

	return kt, nil
}
