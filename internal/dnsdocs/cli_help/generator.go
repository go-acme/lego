package main

//go:generate go run .

import (
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v4/cmd"
	"github.com/urfave/cli/v2"
)

const outputFile = "../../../docs/data/zz_cli_help.toml"

var outputTpl = template.Must(template.New("output").Parse(`# THIS FILE IS AUTO-GENERATED. PLEASE DO NOT EDIT.

{{ range .}}
[[command]]
title   = "{{.Title}}"
content = """
{{.Help}}
"""
{{end -}}
`))

type commandHelp struct{ Title, Help string }

func main() {
	log.SetFlags(0)

	// collect output of various help pages
	var help []commandHelp
	for _, args := range [][]string{
		{"lego", "help"},
		{"lego", "help", "run"},
		{"lego", "help", "renew"},
		{"lego", "help", "revoke"},
		{"lego", "help", "list"},
		{"lego", "dnshelp"},
	} {
		help = append(help, run(args))
	}

	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatalf("cannot open cli_help.toml: %v", err)
	}
	defer f.Close()

	if err = outputTpl.Execute(f, help); err != nil {
		log.Fatalf("failed to write cli_help.toml: %v", err)
	}

	log.Println("cli_help.toml updated")
}

var lego = func() *cli.App {
	// Construct cli app, very similar to cmd/lego/main.go.
	// Notable differences:
	// - substitute "." for CWD in default config path, as the user
	//   will very likely see a different path
	// - do not include version information, because we're likely
	//   running against a snapshot
	// - skip DNS help and provider list, as initialization takes time,
	//   and we don't generate `lego dns --help` here
	app := cli.NewApp()
	app.Name = "lego"
	app.HelpName = "lego"
	app.Usage = "Let's Encrypt client written in Go"
	app.Flags = cmd.CreateFlags("./.lego")
	app.Commands = cmd.CreateCommands()
	return app
}()

func run(args []string) commandHelp {
	var w = lego.Writer
	defer func() { lego.Writer = w }()

	var buf bytes.Buffer
	lego.Writer = &buf

	if err := lego.Run(args); err != nil {
		log.Fatalf("running %s failed: %v", args, err)
	}
	return commandHelp{
		Title: strings.Join(args, " "),
		Help:  strings.TrimSpace(buf.String()),
	}
}

func init() {

}
