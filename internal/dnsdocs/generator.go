package main

//go:generate go run .

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	readmePath  = root + "README.md"
)

const (
	startLine = "<!-- START DNS PROVIDERS LIST -->"
	endLine   = "<!-- END DNS PROVIDERS LIST -->"
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

	// generate README.md
	err = generateReadMe(models)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Documentation for %d DNS providers has been generated.\n", len(models.Providers)+1)
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

func generateReadMe(models *Providers) error {
	max, lines := extractTableData(models)

	file, err := os.Open(readmePath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	var skip bool

	buffer := bytes.NewBufferString("")

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		text := fileScanner.Text()

		if text == startLine {
			_, _ = fmt.Fprintln(buffer, text)
			err = writeDNSTable(buffer, lines, max)
			if err != nil {
				return err
			}
			skip = true
		}

		if text == endLine {
			skip = false
		}

		if skip {
			continue
		}

		_, _ = fmt.Fprintln(buffer, text)
	}

	if fileScanner.Err() != nil {
		return fileScanner.Err()
	}

	if skip {
		return errors.New("missing end tag")
	}

	return os.WriteFile(readmePath, buffer.Bytes(), 0o666)
}

func extractTableData(models *Providers) (int, [][]string) {
	readmePattern := "[%s](https://go-acme.github.io/lego/dns/%s/)"

	items := []string{fmt.Sprintf(readmePattern, "Manual", "manual")}

	var max int

	for _, pvd := range models.Providers {
		item := fmt.Sprintf(readmePattern, strings.ReplaceAll(pvd.Name, "|", "/"), pvd.Code)
		items = append(items, item)

		if max < len(item) {
			max = len(item)
		}
	}

	const nbCol = 4

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i]) < strings.ToLower(items[j])
	})

	var lines [][]string
	var line []string

	for i, item := range items {
		switch {
		case len(line) == nbCol:
			lines = append(lines, line)
			line = []string{item}

		case i == len(items)-1:
			line = append(line, item)
			for j := len(line); j < nbCol; j++ {
				line = append(line, "")
			}
			lines = append(lines, line)

		default:
			line = append(line, item)
		}
	}

	if len(line) < nbCol {
		for j := len(line); j < nbCol; j++ {
			line = append(line, "")
		}
		lines = append(lines, line)
	}

	return max, lines
}

func writeDNSTable(w io.Writer, lines [][]string, size int) error {
	_, err := fmt.Fprintf(w, "\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "|%[1]s|%[1]s|%[1]s|%[1]s|\n", strings.Repeat(" ", size+2))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "|%[1]s|%[1]s|%[1]s|%[1]s|\n", strings.Repeat("-", size+2))
	if err != nil {
		return err
	}

	linePattern := fmt.Sprintf("| %%-%[1]ds | %%-%[1]ds | %%-%[1]ds | %%-%[1]ds |\n", size)
	for _, line := range lines {
		_, err = fmt.Fprintf(w, linePattern, line[0], line[1], line[2], line[3])
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(w, "\n")
	return err
}
