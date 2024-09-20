package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

const flgCode = "code"

func createDNSHelp() *cli.Command {
	return &cli.Command{
		Name:   "dnshelp",
		Usage:  "Shows additional help for the '--dns' global option",
		Action: dnsHelp,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flgCode,
				Aliases: []string{"c"},
				Usage:   fmt.Sprintf("DNS code: %s", allDNSCodes()),
			},
		},
	}
}

func dnsHelp(ctx *cli.Context) error {
	code := ctx.String(flgCode)
	if code == "" {
		w := tabwriter.NewWriter(ctx.App.Writer, 0, 0, 2, ' ', 0)
		ew := &errWriter{w: w}

		ew.writeln(`Credentials for DNS providers must be passed through environment variables.`)
		ew.writeln()
		ew.writeln(`To display the documentation for a specific DNS provider, run:`)
		ew.writeln()
		ew.writeln("\t$ lego dnshelp -c code")
		ew.writeln()
		ew.writeln("Supported DNS providers:")
		ew.writef("\t%s\n", allDNSCodes())
		ew.writeln()
		ew.writeln("More information: https://go-acme.github.io/lego/dns")

		if ew.err != nil {
			return ew.err
		}

		return w.Flush()
	}

	return displayDNSHelp(ctx.App.Writer, strings.ToLower(code))
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) writeln(a ...interface{}) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintln(ew.w, a...)
}

func (ew *errWriter) writef(format string, a ...interface{}) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintf(ew.w, format, a...)
}
