package main

//go:generate go run .

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v5/cmd"
	"github.com/go-acme/lego/v5/internal/generators/clihelp/internal"
	"github.com/urfave/cli/v3"
)

const outputFile = "../../../docs/data/zz_cli_help.toml"

const baseTemplate = `# THIS FILE IS AUTO-GENERATED. PLEASE DO NOT EDIT.

{{ range .}}
[[command]]
title   = "{{.Title}}"
content = """
{{.Help}}
"""
{{end -}}
`

type commandHelp struct {
	Title string
	Help  string
}

func main() {
	log.SetFlags(0)

	err := generate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("cli_help.toml updated")
}

func generate(ctx context.Context) error {
	app := createStubApp()

	outputTpl := template.Must(template.New("output").Parse(baseTemplate))

	// collect output of various help pages
	var help []commandHelp

	for _, args := range [][]string{
		{"lego", "-h"},
		{"lego", "run", "-h"},
		{"lego", "accounts", "register", "-h"},
		{"lego", "accounts", "recover", "-h"},
		{"lego", "accounts", "keyrollover", "-h"},
		{"lego", "accounts", "list", "-h"},
		{"lego", "certificates", "revoke", "-h"},
		{"lego", "certificates", "list", "-h"},
		{"lego", "archives", "restore", "-h"},
		{"lego", "archives", "list", "-h"},
		{"lego", "dnshelp", "-h"},
		{"lego", "migrate", "-h"},
	} {
		content, err := run(ctx, app, args)
		if err != nil {
			return fmt.Errorf("running %s failed: %w", args, err)
		}

		help = append(help, content)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("cannot open cli_help.toml: %w", err)
	}

	err = outputTpl.Execute(f, help)

	defer func() { _ = f.Close() }()

	if err != nil {
		return fmt.Errorf("failed to write cli_help.toml: %w", err)
	}

	return nil
}

// createStubApp Construct cli app, very similar to main.go.
// Notable differences:
// - do not include version information, because we're likely running against a snapshot
// - skip DNS help and provider list, as initialization takes time, and we don't generate `lego dns --help` here.
func createStubApp() *cli.Command {
	cli.RootCommandHelpTemplate = internal.RootCommandHelpTemplate
	cli.CommandHelpTemplate = internal.CommandHelpTemplate
	cli.SubcommandHelpTemplate = internal.SubcommandHelpTemplate
	cli.HelpPrinter = internal.PrintMarkdown

	root := cmd.CreateRootCommand()
	root.EnableShellCompletion = false
	root.Usage = "Get or renew a certificate with a configuration file"

	// Hides the subcommands for a better render of the root command.
	for _, command := range root.Commands {
		command.Hidden = true
	}

	root.Before = nil

	return root
}

func run(ctx context.Context, app *cli.Command, args []string) (h commandHelp, err error) {
	w := app.Writer

	defer func() { app.Writer = w }()

	var buf bytes.Buffer

	app.Writer = &buf

	if app.Command(args[1]) != nil {
		app.Command(args[1]).Writer = app.Writer
	}

	if err := app.Run(ctx, args); err != nil {
		return h, err
	}

	return commandHelp{
		Title: strings.Join(args, " "),
		Help:  strings.TrimSpace(buf.String()),
	}, nil
}
