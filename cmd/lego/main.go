// Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-acme/lego/v4/cmd"
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "lego"
	app.HelpName = "lego"
	app.Usage = "Let's Encrypt client written in Go"
	app.EnableBashCompletion = true

	app.Version = getVersion()
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("lego version %s %s/%s\n", c.App.Version, runtime.GOOS, runtime.GOARCH)
	}

	var defaultPath string
	cwd, err := os.Getwd()
	if err == nil {
		defaultPath = filepath.Join(cwd, ".lego")
	}

	app.Flags = cmd.CreateFlags(defaultPath)

	app.Before = cmd.Before

	app.Commands = cmd.CreateCommands()

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
