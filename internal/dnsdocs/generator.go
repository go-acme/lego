package main

//go:generate go run .

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

const (
	root        = "../../"
	dnsPackage  = root + "providers/dns"
	mdTemplate  = root + "internal/dnsdocs/dns.md.tmpl"
	cliTemplate = root + "internal/dnsdocs/dns.go.tmpl"
	cliOutput   = root + "cmd/zz_gen_cmd_dnshelp.go"
	docOutput   = root + "docs/content/dns"
)

type Model struct {
	Name          string         // Real name of the DNS provider
	Code          string         // DNS code
	Since         string         // First lego version
	URL           string         // DNS provider URL
	Description   string         // Provider summary
	Example       string         // CLI example
	Configuration *Configuration // Environment variables
	Links         *Links         // Links
	Additional    string         // Extra documentation
	GeneratedFrom string         // Source file
}

type Configuration struct {
	Credentials map[string]string
	Additional  map[string]string
}

type Links struct {
	API      string
	GoClient string
}

type Providers struct {
	Providers []Model
}

func main() {
	models := &Providers{}

	err := filepath.Walk(dnsPackage, walker(models))
	if err != nil {
		log.Fatal(err)
	}

	// generate CLI help
	err = generateCLIHelp(models)
	if err != nil {
		log.Fatal(err)
	}
}

func walker(prs *Providers) func(string, os.FileInfo, error) error {
	return func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".toml" {
			m := Model{}

			m.GeneratedFrom, err = filepath.Rel(root, path)
			if err != nil {
				return err
			}

			_, err := toml.DecodeFile(path, &m)
			if err != nil {
				return err
			}

			prs.Providers = append(prs.Providers, m)

			// generate documentation
			return generateDocumentation(m)
		}

		return nil
	}
}

func generateDocumentation(m Model) error {
	filename := filepath.Join(docOutput, "zz_gen_"+m.Code+".md")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	return template.Must(template.ParseFiles(mdTemplate)).Execute(file, m)
}

func generateCLIHelp(models *Providers) error {
	filename := filepath.Join(cliOutput)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	tlt := template.New(filepath.Base(cliTemplate)).Funcs(map[string]interface{}{
		"safe": func(src string) string {
			return strings.ReplaceAll(src, "`", "'")
		},
	})

	b := &bytes.Buffer{}
	err = template.Must(tlt.ParseFiles(cliTemplate)).Execute(b, models)
	if err != nil {
		return err
	}

	// gofmt
	source, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(source)
	return err
}
