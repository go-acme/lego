package main

import (
	"log"
	"os"
	"path"

	"github.com/codegangsta/cli"
)

// Logger is used to log errors; if nil, the default log.Logger is used.
var Logger *log.Logger

// logger is an helper function to retrieve the available logger
func logger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return Logger
}

func main() {

	app := cli.NewApp()
	app.Name = "lego"
	app.Usage = "Let's encrypt client to go!"
	app.Version = "0.0.2"

	cwd, err := os.Getwd()
	if err != nil {
		logger().Fatal("Could not determine current working directory. Please pass --path.")
	}
	defaultPath := path.Join(cwd, ".lego")

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Register an account, then create and install a certificate",
			Action: run,
		},
		{
			Name:  "auth",
			Usage: "Create a certificate - must already have an account",
			Action: func(c *cli.Context) {
				logger().Fatal("Not implemented")
			},
		},
		{
			Name:   "revoke",
			Usage:  "Revoke a certificate",
			Action: revoke,
		},
	}

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "domains, d",
			Usage: "Add domains to the process",
		},
		cli.StringFlag{
			Name:  "server, s",
			Value: "https://acme-staging.api.letsencrypt.org/",
			Usage: "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.",
		},
		cli.StringFlag{
			Name:  "email, m",
			Usage: "Email used for registration and recovery contact.",
		},
		cli.IntFlag{
			Name:  "rsa-key-size, B",
			Value: 2048,
			Usage: "Size of the RSA key.",
		},
		cli.StringFlag{
			Name:  "path",
			Usage: "Directory to use for storing the data",
			Value: defaultPath,
		},
		cli.StringFlag{
			Name:  "port",
			Usage: "Challenges will use this port to listen on. Please make sure to forward port 443 to this port on your machine. Otherwise use setcap on the binary",
		},
		cli.BoolFlag{
			Name:  "devMode",
			Usage: "If set to true, all client side challenge pre-tests are skipped.",
		},
	}

	app.Run(os.Args)
}
