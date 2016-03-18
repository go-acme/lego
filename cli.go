// Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/xenolf/lego/acme"
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

var gittag string

func main() {
	app := cli.NewApp()
	app.Name = "lego"
	app.Usage = "Let's encrypt client to go!"

	version := "0.2.0"
	if strings.HasPrefix(gittag, "v") {
		version = gittag
	}

	app.Version = version

	acme.UserAgent = "lego/" + app.Version

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
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-bundle",
					Usage: "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
				},
			},
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
					Name:  "days",
					Value: 0,
					Usage: "The number of days left on a certificate to renew it.",
				},
				cli.BoolFlag{
					Name:  "reuse-key",
					Usage: "Used to indicate you want to reuse your current private key for the new certificate.",
				},
				cli.BoolFlag{
					Name:  "no-bundle",
					Usage: "Do not create a certificate bundle by adding the issuers certificate to the new certificate.",
				},
			},
		},
		{
			Name:   "httphelp",
			Usage:  "Shows additional help for the --http global option",
			Action: httphelp,
		},
		{
			Name:   "dnshelp",
			Usage:  "Shows additional help for the --dns global option",
			Action: dnshelp,
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
		cli.BoolFlag{
			Name:  "accept-tos, a",
			Usage: "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.",
		},
		cli.StringFlag{
			Name:  "key-type, k",
			Value: "rsa2048",
			Usage: "Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384",
		},
		cli.StringFlag{
			Name:  "path",
			Usage: "Directory to use for storing the data",
			Value: defaultPath,
		},
		cli.StringSliceFlag{
			Name:  "exclude, x",
			Usage: "Explicitly disallow solvers by name from being used. Solvers: \"http-01\", \"tls-sni-01\".",
		},
		cli.StringFlag{
			Name:  "http-address",
			Usage: "Set the port and interface to use for HTTP based challenges to listen on. Supported: interface:port or :port",
		},
		cli.StringFlag{
			Name:  "tls-address",
			Usage: "Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port",
		},
		cli.StringFlag{
			Name:  "http",
			Usage: "Solve an HTTP challenge using the specified provider. Disables all other challenges. Run 'lego httphelp' for help on usage.",
		},
		cli.StringFlag{
			Name:  "dns",
			Usage: "Solve a DNS challenge using the specified provider. Disables all other challenges. Run 'lego dnshelp' for help on usage.",
		},
	}

	app.Run(os.Args)
}

func httphelp(c *cli.Context) {
	fmt.Printf(
		`Additional information needed for HTTP providers must be passed through
environment variables.

Here is an example bash command using the Webroot provider:

  $ WEBROOT_PATH=/path/to/webroot \
    lego --http webroot --domains www.example.com --email me@bar.com run

`)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Valid providers and their associated environment variables:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "\twebroot:\tWEBROOT_PATH")
	w.Flush()

	fmt.Println(`
For a more detailed explanation of an HTTP provider's variables,
please consult their online documentation.`)
}

func dnshelp(c *cli.Context) {
	fmt.Printf(
		`Credentials for DNS providers must be passed through environment variables.

Here is an example bash command using the CloudFlare DNS provider:

  $ CLOUDFLARE_EMAIL=foo@bar.com \
    CLOUDFLARE_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
    lego --dns cloudflare --domains www.example.com --email me@bar.com run

`)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Valid providers and their associated credential environment variables:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "\tcloudflare:\tCLOUDFLARE_EMAIL, CLOUDFLARE_API_KEY")
	fmt.Fprintln(w, "\tdigitalocean:\tDO_AUTH_TOKEN")
	fmt.Fprintln(w, "\tdnsimple:\tDNSIMPLE_EMAIL, DNSIMPLE_API_KEY")
	fmt.Fprintln(w, "\tgandi:\tGANDI_API_KEY")
	fmt.Fprintln(w, "\tmanual:\tnone")
	fmt.Fprintln(w, "\tnamecheap:\tNAMECHEAP_API_USER, NAMECHEAP_API_KEY")
	fmt.Fprintln(w, "\trfc2136:\tRFC2136_TSIG_KEY, RFC2136_TSIG_SECRET,\n\t\tRFC2136_TSIG_ALGORITHM, RFC2136_NAMESERVER")
	fmt.Fprintln(w, "\troute53:\tAWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION")
	w.Flush()

	fmt.Println(`
For a more detailed explanation of a DNS provider's credential variables,
please consult their online documentation.`)
}
