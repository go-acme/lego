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
	app.Version = "0.1.0"

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
			Name:   "revoke",
			Usage:  "Revoke a certificate",
			Action: revoke,
		},
		{
			Name:   "renew",
			Usage:  "Renew a certificate",
			Action: renew,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name: "days",
					Value: 0,
					Usage: "The number of days left on a certificate to renew it.",
				},
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "domains, d",
			Usage: "Add domains to the process",
		},
		cli.StringFlag{
			Name:  "server, s",
			Value: "https://acme-v01.api.letsencrypt.org/directory",
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
	}

	app.Run(os.Args)
}
