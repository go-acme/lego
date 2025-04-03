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

	"github.com/go-acme/lego/v4/cmd"
	"github.com/urfave/cli/v3"
)

const outputFile = "../../docs/data/zz_cli_help.toml"

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

	err := generate()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("cli_help.toml updated")
}

func generate() error {
	app := createStubApp()

	outputTpl := template.Must(template.New("output").Parse(baseTemplate))

	ctx := context.Background()

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

// createStubApp Construct cli app, very similar to cmd/lego/main.go.
// Notable differences:
// - substitute "." for CWD in default config path, as the user will very likely see a different path
// - do not include version information, because we're likely running against a snapshot
// - skip DNS help and provider list, as initialization takes time, and we don't generate `lego dns --help` here.
func createStubApp() *cli.Command {
	return &cli.Command{
		Name:     "lego",
		Usage:    "Let's Encrypt client written in Go",
		Flags:    cmd.CreateFlags("./.lego"),
		Commands: cmd.CreateCommands(),
	}
}

func run(ctx context.Context, cmd *cli.Command, args []string) (h commandHelp, err error) {
	w := cmd.Writer
	defer func() { cmd.Writer = w }()

	var buf bytes.Buffer
	cmd.Writer = &buf

	if err := cmd.Run(ctx, args); err != nil {
		return h, err
	}

	return commandHelp{
		Title: strings.Join(args, " "),
		Help:  strings.TrimSpace(buf.String()),
	}, nil
}
