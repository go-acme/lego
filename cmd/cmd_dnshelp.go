package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v3"
)

const flgCode = "code"

func createDNSHelp() *cli.Command {
	return &cli.Command{
		Name:   "dnshelp",
		Usage:  "Shows additional help for the '--dns' global option",
		Action: dnsHelp,
		Flags:  createDNSHelpFlags(),
	}
}

func createDNSHelpFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flgCode,
			Aliases: []string{"c"},
			Usage:   fmt.Sprintf("DNS code: %s", allDNSCodes()),
		},
	}
}

func dnsHelp(_ context.Context, cmd *cli.Command) error {
	code := cmd.String(flgCode)
	if code == "" {
		w := tabwriter.NewWriter(cmd.Writer, 0, 0, 2, ' ', 0)
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

	return displayDNSHelp(cmd.Writer, strings.ToLower(code))
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) writeln(a ...any) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintln(ew.w, a...)
}

func (ew *errWriter) writef(format string, a ...any) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintf(ew.w, format, a...)
}
