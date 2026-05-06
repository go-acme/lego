package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createArchivesList() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List all archives.",
		Action: listArchives,
		Flags: []cli.Flag{
			flags.CreatePathFlag(false),
		},
	}
}

func listArchives(_ context.Context, cmd *cli.Command) error {
	archiver := storage.NewArchiver(cmd.String(flags.FlgPath))

	accountPaths, err := archiver.ListArchivedAccounts()
	if err != nil {
		return err
	}

	displayArchivePaths("Account", accountPaths)

	certificatePaths, err := archiver.ListArchivedCertificates()
	if err != nil {
		return err
	}

	displayArchivePaths("Certificate", certificatePaths)

	if len(accountPaths) == 0 && len(certificatePaths) == 0 {
		log.Info("No archives were found.")
	}

	return nil
}

func displayArchivePaths(kind string, paths []string) {
	if len(paths) == 0 {
		return
	}

	slices.Sort(paths)

	fmt.Println(kind + " archives:")

	prefix := "├──"

	for i, filename := range paths {
		if i == len(paths)-1 {
			prefix = "└──"
		}

		date, _ := parseArchiveDate(filename)

		fmt.Println(prefix, filepath.Base(filename), "("+date.String()+")")
	}

	fmt.Println()
}
