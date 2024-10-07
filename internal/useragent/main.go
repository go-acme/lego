package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "lego-releaser"
	app.Usage = "Lego releaser"
	app.HelpName = "releaser"
	app.Commands = []*cli.Command{
		{
			Name:   "release",
			Usage:  "Update file for a release",
			Action: release,
			Before: func(ctx *cli.Context) error {
				mode := ctx.String("mode")
				switch mode {
				case "patch", "minor", "major":
					return nil
				default:
					return fmt.Errorf("invalid mode: %s", mode)
				}
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "mode",
					Aliases: []string{"m"},
					Value:   "patch",
					Usage:   "The release mode: patch|minor|major",
				},
			},
		},
		{
			Name:   "detach",
			Usage:  "Update file post release",
			Action: detach,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func release(ctx *cli.Context) error {
	mode := ctx.String("mode")

	generators := []*Generator{
		NewGenerator(senderBaseUserAgent, senderTemplate, senderSourceFile),
		NewGenerator(dnsBaseUserAgent, dnsTemplate, dnsSourceFile),
	}

	for _, generator := range generators {
		err := generator.Release(mode)
		if err != nil {
			return err
		}
	}

	return nil
}

func detach(_ *cli.Context) error {
	generators := []*Generator{
		NewGenerator(senderBaseUserAgent, senderTemplate, senderSourceFile),
		NewGenerator(dnsBaseUserAgent, dnsTemplate, dnsSourceFile),
	}

	for _, generator := range generators {
		err := generator.Detach()
		if err != nil {
			return err
		}
	}

	return nil
}
