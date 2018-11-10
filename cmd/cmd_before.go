package cmd

import (
	"os"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/log"
)

func Before(c *cli.Context) error {
	if c.GlobalString("path") == "" {
		log.Fatal("Could not determine current working directory. Please pass --path.")
	}

	err := createNonExistingFolder(c.GlobalString("path"))
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	return nil
}

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}
