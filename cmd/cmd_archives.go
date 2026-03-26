package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func createArchives() *cli.Command {
	return &cli.Command{
		Name:  "archives",
		Usage: "Archives management.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all archives.",
				Action: listArchives,
				Flags: []cli.Flag{
					flags.CreatePathFlag(false),
				},
			},
			{
				Name:   "restore",
				Usage:  "Restore an archive.",
				Action: restoreArchive,
				Flags: []cli.Flag{
					flags.CreatePathFlag(false),
				},
			},
		},
	}
}

func restoreArchive(_ context.Context, cmd *cli.Command) error {
	archiver := storage.NewArchiver(cmd.String(flags.FlgPath))

	accountPaths, err := archiver.ListArchivedAccounts()
	if err != nil {
		return err
	}

	certificatePaths, err := archiver.ListArchivedCertificates()
	if err != nil {
		return err
	}

	all := slices.Concat(accountPaths, certificatePaths)

	if len(all) == 0 {
		log.Info("No archives were found.")
		return nil
	}

	var options []string

	for i, filename := range all {
		date, errP := parseArchiveDate(filename)
		if errP != nil {
			return errP
		}

		fmt.Printf("%d: %s - %s - %s\n", i+1, filepath.Base(filepath.Dir(filename)), date, filepath.Base(filename))

		options = append(options, strconv.Itoa(i+1))
	}

	fmt.Println()

	choice := choose("Choose the archive to restore:", options)

	index, err := strconv.Atoi(choice)
	if err != nil {
		return err
	}

	err = archiver.Restore(all[index-1])
	if err != nil {
		return fmt.Errorf("restore archive: %w", err)
	}

	log.Info("Account restored.")

	return nil
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

func parseArchiveDate(filename string) (time.Time, error) {
	lastIndex := strings.LastIndex(filename, "_")

	unixRaw, err := strconv.ParseInt(strings.TrimSuffix(filename[lastIndex+1:], ".zip"), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixRaw, 0), nil
}
