package main

//go:generate go run .

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	html "html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v5/internal/generators/dns/descriptors"
)

//go:embed templates
var templateFS embed.FS

const (
	root = "../../../../"

	cliOutput    = root + "cmd/zz_gen_cmd_dnshelp.go"
	docOutput    = root + "docs/content/dns"
	readmePath   = root + "README.md"
	workflowPath = root + ".github/workflows/tests-lint.yml"
)

const (
	mdTemplate       = "templates/dns.md.tmpl"
	cliTemplate      = "templates/dns.go.tmpl"
	readmeTemplate   = "templates/readme.md.tmpl"
	workflowTemplate = "templates/tests-lint.yml.tmpl"
)

const (
	readmeStartLine = "<!-- START DNS PROVIDERS LIST -->"
	readmeEndLine   = "<!-- END DNS PROVIDERS LIST -->"
)

const (
	workflowStartLine = "# START GENERATED"
	workflowEndLine   = "# END GENERATED"
)

type templateExecutor interface {
	Execute(wr io.Writer, data any) error
}

func main() {
	models, err := descriptors.GetProviderInformation(root)
	if err != nil {
		log.Fatal(err)
	}

	err = cleanDocumentation()
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

	// generate Workflow
	err = generateWorkflow(models)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Documentation for %d DNS providers has been generated.\n", len(models.Providers)+1)
}

func cleanDocumentation() error {
	paths, err := filepath.Glob(filepath.Join(docOutput, "zz_gen_*.md"))
	if err != nil {
		return err
	}

	for _, p := range paths {
		err = os.RemoveAll(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateDocumentation(m descriptors.Provider) error {
	filename := filepath.Join(docOutput, "zz_gen_"+m.Code+".md")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	return template.Must(template.ParseFS(templateFS, mdTemplate)).Execute(file, m)
}

func generateCLIHelp(models *descriptors.Providers) error {
	filename := filepath.Clean(cliOutput)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	b := &bytes.Buffer{}

	err = template.Must(
		template.New(filepath.Base(cliTemplate)).Funcs(map[string]any{
			"safe": func(src string) string {
				return strings.ReplaceAll(src, "`", "'")
			},
		}).ParseFS(templateFS, cliTemplate),
	).Execute(b, models)
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
	tpl := html.Must(html.New(filepath.Base(readmeTemplate)).ParseFS(templateFS, readmeTemplate))

	providers := orderProviders(models)

	return generateWithPlaceholder(providers, tpl, readmePath, readmeStartLine, readmeEndLine)
}

func generateWorkflow(models *descriptors.Providers) error {
	tpl := template.Must(template.New(filepath.Base(workflowTemplate)).ParseFS(templateFS, workflowTemplate))

	slices.SortFunc(models.Providers, func(a, b descriptors.Provider) int {
		return strings.Compare(strings.ToLower(a.Code), strings.ToLower(b.Code))
	})

	return generateWithPlaceholder(models.Providers, tpl, workflowPath, workflowStartLine, workflowEndLine)
}

func generateWithPlaceholder(data any, tpl templateExecutor, outputPath, start, end string) error {
	file, err := os.Open(outputPath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	var skip bool

	buffer := bytes.NewBufferString("")

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		text := fileScanner.Text()

		if text == start {
			_, _ = fmt.Fprintln(buffer, text)
			if err = tpl.Execute(buffer, data); err != nil {
				return err
			}

			skip = true
		}

		if text == end {
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

	return os.WriteFile(outputPath, buffer.Bytes(), 0o666)
}

func orderProviders(models *descriptors.Providers) [][]descriptors.Provider {
	const nbCol = 4

	slices.SortFunc(models.Providers, func(a, b descriptors.Provider) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	var (
		matrix [][]descriptors.Provider
		row    []descriptors.Provider
	)

	for i, p := range models.Providers {
		switch {
		case len(row) == nbCol:
			matrix = append(matrix, row)
			row = []descriptors.Provider{p}

		case i == len(models.Providers)-1:
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
