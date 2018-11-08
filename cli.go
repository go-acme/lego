// Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package main

import (
	"os"
	"path/filepath"

	"github.com/xenolf/lego/cmd"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/log"
)

var (
	version = "dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "lego"
	app.Usage = "Let's Encrypt client written in Go"
	app.Version = version

	defaultPath := ""
	cwd, err := os.Getwd()
	if err == nil {
		defaultPath = filepath.Join(cwd, ".lego")
	}
	app.Flags = cmd.CreateFlags(defaultPath)

	app.Before = func(c *cli.Context) error {
		if c.GlobalString("path") == "" {
			log.Fatal("Could not determine current working directory. Please pass --path.")
		}
		return nil
	}

	app.Commands = cmd.CreateCommands()

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
