package main

//go:generate go run .

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"unicode"
)

const (
	root = "../../../"

	inputPath     = "cas.json"
	outputPath    = "lego/zz_gen_ca.go"
	outputDocPath = "docs/content/advanced/zz_gen_caservers.md"
)

//go:embed templates/ca.go.tmpl
var srcTemplate string

//go:embed templates/caservers.md.tmpl
var docTemplate string

type Entry struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Code             string `json:"code"`
	DirectoryURL     string `json:"directoryURL"`
	DocumentationURL string `json:"documentationURL"`
}

func main() {
	err := generate()
	if err != nil {
		log.Fatal(err)
	}
}

func generate() error {
	entries, err := getCADescription()
	if err != nil {
		return err
	}

	err = generateGoFile(entries)
	if err != nil {
		return fmt.Errorf("generate go file: %w", err)
	}

	err = generateDocFile(entries)
	if err != nil {
		return fmt.Errorf("generate doc file: %w", err)
	}

	return nil
}

func generateGoFile(entries []Entry) error {
	b, err := executeTemplate(entries, srcTemplate)
	if err != nil {
		return err
	}

	// gofmt
	source, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}

	output, err := os.Create(filepath.Join(root, outputPath))
	if err != nil {
		return err
	}

	defer func() { _ = output.Close() }()

	_, err = output.Write(source)
	if err != nil {
		return err
	}

	return nil
}

func generateDocFile(entries []Entry) error {
	b, err := executeTemplate(entries, docTemplate)
	if err != nil {
		return err
	}

	output, err := os.Create(filepath.Join(root, outputDocPath))
	if err != nil {
		return err
	}

	defer func() { _ = output.Close() }()

	_, err = output.Write(b.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func getCADescription() ([]Entry, error) {
	input, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	defer func() { _ = input.Close() }()

	var entries []Entry

	slices.SortFunc(entries, func(a, b Entry) int {
		return strings.Compare(a.Code, b.Code)
	})

	err = json.NewDecoder(input).Decode(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func executeTemplate(entries []Entry, src string) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}

	err := template.Must(
		template.New("ca").Funcs(map[string]any{
			"ToLower":   strings.ToLower,
			"ToVarName": toVarName,
		}).Parse(src),
	).Execute(b, entries)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func toVarName(s string) string {
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), "")
}
