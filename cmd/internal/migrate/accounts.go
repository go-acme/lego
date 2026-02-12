package migrate

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

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

		var account storage.Account

		err = json.Unmarshal(data, &account)
		if err != nil {
			return fmt.Errorf("could not parse the account file %q: %w", srcAccountFilePath, err)
		}

		account.ID = account.GetID()

		accountsDir := filepath.Dir(srcAccountFilePath)

		srcKeyPath := filepath.Join(accountsDir, "keys", account.ID+storage.ExtKey)

		account.KeyType, err = getKeyType(srcKeyPath)
		if err != nil {
			return fmt.Errorf("could not guess the account key type: %w", err)
		}

		newAccountDir := filepath.Join(accountsDir, string(account.KeyType))

		err = os.MkdirAll(newAccountDir, 0o700)
		if err != nil {
			return fmt.Errorf("could not create the directory %q: %w", newAccountDir, err)
		}

		// Rename the private key file.

		dstKeyPath := filepath.Join(newAccountDir, account.ID+storage.ExtKey)

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

	kt, err := guessKeyType(pk)
	if err != nil {
		return "", fmt.Errorf("could not guess the private key type: %w", err)
	}

	return kt, nil
}

func guessKeyType(key crypto.PrivateKey) (certcrypto.KeyType, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		switch k.Size() {
		case 256:
			return certcrypto.RSA2048, nil
		case 384:
			return certcrypto.RSA3072, nil
		case 512:
			return certcrypto.RSA4096, nil
		case 1024:
			return certcrypto.RSA8192, nil
		default:
			return "", fmt.Errorf("unsupported RSA key size: %d", k.Size())
		}

	case *ecdsa.PrivateKey:
		switch k.Curve.Params().BitSize {
		case 256:
			return certcrypto.EC256, nil
		case 384:
			return certcrypto.EC384, nil
		default:
			return "", fmt.Errorf("unsupported ECDSA key size: %d", k.Curve.Params().BitSize)
		}

	default:
		return "", fmt.Errorf("unsupported key type: %T", key)
	}
}
