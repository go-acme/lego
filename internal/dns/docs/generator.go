package main

//go:generate go run .

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/format"
	html "html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v4/internal/dns/descriptors"
)

const (
	root = "../../../"

	mdTemplate     = root + "internal/dns/docs/dns.md.tmpl"
	cliTemplate    = root + "internal/dns/docs/dns.go.tmpl"
	cliOutput      = root + "cmd/zz_gen_cmd_dnshelp.go"
	docOutput      = root + "docs/content/dns"
	readmeTemplate = root + "internal/dns/docs/readme.md.tmpl"
	readmePath     = root + "README.md"
)

const (
	startLine = "<!-- START DNS PROVIDERS LIST -->"
	endLine   = "<!-- END DNS PROVIDERS LIST -->"
)

func main() {
	models, err := descriptors.GetProviderInformation(root)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range models.Providers {
		// generate documentation
		err = generateDocumentation(m)
		if err != nil {
			log.Fatal(err)
		}
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

func generateDocumentation(m descriptors.Provider) error {
	filename := filepath.Join(docOutput, "zz_gen_"+m.Code+".md")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	return template.Must(template.ParseFiles(mdTemplate)).Execute(file, m)
}

func generateCLIHelp(models *descriptors.Providers) error {
	filename := filepath.Clean(cliOutput)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

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

func generateReadMe(models *descriptors.Providers) error {
	tpl := html.Must(html.New(filepath.Base(readmeTemplate)).ParseFiles(readmeTemplate))
	providers := orderProviders(models)

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
			if err = tpl.Execute(buffer, providers); err != nil {
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

func orderProviders(models *descriptors.Providers) [][]descriptors.Provider {
	providers := append(models.Providers, descriptors.Provider{
		Name: "Manual",
		Code: "manual",
	})

	const nbCol = 4

	slices.SortFunc(providers, func(a, b descriptors.Provider) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	var matrix [][]descriptors.Provider
	var row []descriptors.Provider

	for i, p := range providers {
		switch {
		case len(row) == nbCol:
			matrix = append(matrix, row)
			row = []descriptors.Provider{p}

		case i == len(providers)-1:
			row = append(row, p)
			for j := len(row); j < nbCol; j++ {
				row = append(row, descriptors.Provider{})
			}
			matrix = append(matrix, row)

		default:
			row = append(row, p)
		}
	}

	if len(row) < nbCol {
		for j := len(row); j < nbCol; j++ {
			row = append(row, descriptors.Provider{})
		}
		matrix = append(matrix, row)
	}

	return matrix
}
