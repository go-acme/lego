package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/mattn/go-zglob"
	"github.com/urfave/cli/v3"
)

type ListAccount struct {
	storage.Account

	Path string `json:"path,omitempty"`
}

func createListAccounts() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Display information about accounts.",
		Action: listAccounts,
		Flags:  flags.CreateListFlags(),
	}
}

func listAccounts(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool(flags.FlgFormatJSON) {
		return listAccountsJSON(ctx, cmd)
	}

	return listAccountsText(ctx, cmd)
}

func listAccountsText(_ context.Context, cmd *cli.Command) error {
	accounts, err := readAccounts(cmd)
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

func listAccountsJSON(_ context.Context, cmd *cli.Command) error {
	accounts, err := readAccounts(cmd)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(accounts)
}

func readAccounts(cmd *cli.Command) ([]ListAccount, error) {
	accountsStorage := storage.NewAccountsStorage(cmd.String(flags.FlgPath))

	matches, err := zglob.Glob(filepath.Join(accountsStorage.GetRootPath(), "**", "account.json"))
	if err != nil {
		return nil, err
	}

	var accounts []ListAccount

	for _, filename := range matches {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var account storage.Account

		err = json.Unmarshal(data, &account)
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
