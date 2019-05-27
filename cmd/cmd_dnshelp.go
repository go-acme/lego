package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func createDNSHelp() cli.Command {
	return cli.Command{
		Name:   "dnshelp",
		Usage:  "Shows additional help for the '--dns' global option",
		Action: dnsHelp,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "code, c",
				Usage: fmt.Sprintf("DNS code: %s", allDNSCodes()),
			},
		},
	}
}

func dnsHelp(ctx *cli.Context) error {
	code := ctx.String("code")
	if code == "" {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintln(w, `Credentials for DNS providers must be passed through environment variables.`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `To display the documentation for a DNS providers:`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "\t$ lego dnshelp -c code")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "All DNS codes:")
		fmt.Fprintf(w, "\t%s\n", allDNSCodes())
		fmt.Fprintln(w)
		fmt.Fprintln(w, "More information: https://go-acme.github.io/lego/dns")

		return w.Flush()
	}

	displayDNSHelp(strings.ToLower(code))

	return nil
}
