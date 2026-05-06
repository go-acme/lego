package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func createArchives() *cli.Command {
	return &cli.Command{
		Name:  "archives",
		Usage: "Archives management.",
		Commands: []*cli.Command{
			createArchivesList(),
			createArchivesRestore(),
		},
	}
}

func parseArchiveDate(filename string) (time.Time, error) {
	lastIndex := strings.LastIndex(filename, "_")

	unixRaw, err := strconv.ParseInt(strings.TrimSuffix(filename[lastIndex+1:], ".zip"), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixRaw, 0), nil
}
