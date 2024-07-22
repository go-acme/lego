package cmd

import (
	"fmt"

	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
)

func createRegister() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register an account",
		Before: func(ctx *cli.Context) error {
			return nil
		},
		Action: registerAction,
		Flags:  []cli.Flag{},
	}
}

func registerAction(ctx *cli.Context) error {
	accountsStorage := NewAccountsStorage(ctx)

	account, client := setup(ctx, accountsStorage)
	setupChallenges(ctx, client)

	if account.Registration == nil {
		reg, err := register(ctx, client)
		if err != nil {
			log.Fatalf("Could not complete registration\n\t%v", err)
		}

		account.Registration = reg
		if err = accountsStorage.Save(account); err != nil {
			log.Fatal(err)
		}

		fmt.Printf(rootPathWarningMessage, accountsStorage.GetRootPath())
		return err
	}

	return nil
}
