package main

//go:generate go run .

import (
	"bytes"
	_ "embed"
	"encoding/json"
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

	inputPath  = "cas.json"
	outputPath = "lego/zz_gen_ca.go"
)

//go:embed ca.go.tmpl
var srcTemplate string

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
	output, err := os.Create(filepath.Join(root, outputPath))
	if err != nil {
		return err
	}

	defer func() { _ = output.Close() }()

	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer func() { _ = input.Close() }()

	var entries []Entry

	slices.SortFunc(entries, func(a, b Entry) int {
		return strings.Compare(a.Code, b.Code)
	})

	err = json.NewDecoder(input).Decode(&entries)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}

	err = template.Must(
		template.New("ca").Funcs(map[string]any{
			"ToLower":   strings.ToLower,
			"ToVarName": toVarName,
		}).Parse(srcTemplate),
	).Execute(b, entries)
	if err != nil {
		return err
	}

	// gofmt
	source, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}

	_, err = output.Write(source)
	if err != nil {
		return err
	}

	return nil
}

func toVarName(s string) string {
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), "")
}
