// Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/go-acme/lego/v5/cmd"
	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:                  "lego",
		Usage:                 "Let's Encrypt client written in Go",
		Version:               getVersion(),
		EnableShellCompletion: true,
		Flags:                 cmd.CreateFlags(""),
		Before:                cmd.Before,
		Commands:              cmd.CreateCommands(),
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("lego version %s %s/%s\n", cmd.Version, runtime.GOOS, runtime.GOARCH)
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal("Error", log.ErrorAttr(err))
	}
}
