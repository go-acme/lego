package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
	"github.com/urfave/cli/v3"
)

type ListAccount struct {
	*storage.Account

	Path string `json:"path,omitempty"`
}

func createAccountsList() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Display information about accounts.",
		Action: listAccounts,
		Flags:  flags.CreateListFlags(),
	}
}

func listAccounts(_ context.Context, cmd *cli.Command) error {
	basePath := cmd.String(flags.FlgPath)

	cfg, err := loadConfiguration(cmd)
	if err == nil {
		log.Debug("Configuration loaded from a file.", slog.String("cmd", "accounts list"))

		basePath = cfg.Storage
	}

	if cmd.Bool(flags.FlgFormatJSON) {
		return listAccountsJSON(basePath)
	}

	return listAccountsText(basePath)
}

func listAccountsText(basePath string) error {
	accounts, err := readAccounts(basePath)
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts were found.")
		return nil
	}

	fmt.Println("Found the following accounts:")

	for _, account := range accounts {
		fmt.Println(account.GetID())
		fmt.Println("├── Email:", account.Email)
		fmt.Println("├── Server:", account.Server)
		fmt.Println("├── Key Type:", account.KeyType)
		fmt.Println("└── Path:", account.Path)
		fmt.Println()
	}

	return nil
}

func listAccountsJSON(basePath string) error {
	accounts, err := readAccounts(basePath)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(accounts)
}

func readAccounts(basePath string) ([]ListAccount, error) {
	accountsStorage := storage.NewAccountsStorage(basePath)

	matches, err := zglob.Glob(filepath.Join(accountsStorage.GetRootPath(), "**", "account.json"))
	if err != nil {
		return nil, err
	}

	var accounts []ListAccount

	for _, filename := range matches {
		account, err := storage.ReadJSONFile[storage.Account](filename)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, ListAccount{
			Account: account,
			Path:    filename,
		})
	}

	return accounts, nil
}
