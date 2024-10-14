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

	"github.com/go-acme/lego/v4/internal/dns/descriptors"
)

const (
	root = "../../../"

	mdTemplate  = root + "internal/dns/docs/dns.md.tmpl"
	cliTemplate = root + "internal/dns/docs/dns.go.tmpl"
	cliOutput   = root + "cmd/zz_gen_cmd_dnshelp.go"
	docOutput   = root + "docs/content/dns"
	readmePath  = root + "README.md"
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
	maximum, lines := extractTableData(models)

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
			err = writeDNSTable(buffer, lines, maximum)
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

func extractTableData(models *descriptors.Providers) (int, [][]string) {
	readmePattern := "[%s](https://go-acme.github.io/lego/dns/%s/)"

	items := []string{fmt.Sprintf(readmePattern, "Manual", "manual")}

	var maximum int

	for _, pvd := range models.Providers {
		item := fmt.Sprintf(readmePattern, strings.ReplaceAll(pvd.Name, "|", "/"), pvd.Code)
		items = append(items, item)

		if maximum < len(item) {
			maximum = len(item)
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

	return maximum, lines
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
