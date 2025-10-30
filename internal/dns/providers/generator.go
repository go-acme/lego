package main

//go:generate go run .

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v4/internal/dns/descriptors"
)

const (
	root = "../../../"

	outputPath = "providers/dns/zz_gen_dns_providers.go"
)

//go:embed dns_providers.go.tmpl
var srcTemplate string

func main() {
	err := generate()
	if err != nil {
		log.Fatal(err)
	}
}

func generate() error {
	info, err := descriptors.GetProviderInformation(root)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(root, outputPath))
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	b := &bytes.Buffer{}

	err = template.Must(
		template.New("").Funcs(map[string]any{
			"cleanName": func(src string) string {
				return strings.ReplaceAll(src, "-", "")
			},
		}).Parse(srcTemplate),
	).Execute(b, info)
	if err != nil {
		return err
	}

	// gofmt
	source, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(source)
	if err != nil {
		return err
	}

	fmt.Printf("Switch mapping for %d DNS providers has been generated.\n", len(info.Providers)+1)

	return nil
}
