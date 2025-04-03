// Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-acme/lego/v4/cmd"
	"github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v3"
)

func main() {
	var defaultPath string
	cwd, err := os.Getwd()
	if err == nil {
		defaultPath = filepath.Join(cwd, ".lego")
	}

	app := &cli.Command{
		Name:                  "lego",
		Usage:                 "Let's Encrypt client written in Go",
		Version:               getVersion(),
		EnableShellCompletion: true,
		Flags:                 cmd.CreateFlags(defaultPath),
		Before:                cmd.Before,
		Commands:              cmd.CreateCommands(),
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("lego version %s %s/%s\n", cmd.Version, runtime.GOOS, runtime.GOARCH)
	}

	app.Commands = cmd.CreateCommands()

	ctx := context.Background()

	err = app.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
