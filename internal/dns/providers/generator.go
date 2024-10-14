package main

//go:generate go run .

import (
	"bytes"
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

	srcTemplate = "internal/dns/providers/dns_providers.go.tmpl"
	outputPath  = "providers/dns/zz_gen_dns_providers.go"
)

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

	tmplFile := filepath.Join(root, srcTemplate)

	tlt := template.New(filepath.Base(tmplFile)).Funcs(map[string]interface{}{
		"cleanName": func(src string) string {
			return strings.ReplaceAll(src, "-", "")
		},
	})

	b := &bytes.Buffer{}
	err = template.Must(tlt.ParseFiles(tmplFile)).Execute(b, info)
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
