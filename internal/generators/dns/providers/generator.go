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

	"github.com/go-acme/lego/v5/internal/generators/dns/descriptors"
)

const (
	root = "../../../../"

	outputPath = "providers/dns/zz_gen_dns_providers.go"
)

//go:embed dns_providers.go.tmpl
var srcTemplate string

//go:embed custom_dns_providers.go.tmpl
var srcCustomTemplate string

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

	err = generateMapping(info)
	if err != nil {
		return err
	}

	err = generatedRegister(info)
	if err != nil {
		return err
	}

	return nil
}

func generateMapping(info *descriptors.Providers) error {
	file, err := os.Create(filepath.Join(root, outputPath))
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	err = generateFile(info, srcTemplate, file)
	if err != nil {
		return err
	}

	fmt.Printf("Switch mapping for %d DNS providers has been generated.\n", len(info.Providers)+1)

	return nil
}

func generatedRegister(info *descriptors.Providers) error {
	dir := filepath.Join(root, "providers", "dns", "internal", "dnslight")

	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	noop, err := os.Create(filepath.Join(dir, "zz_gen_noop.go"))
	if err != nil {
		return err
	}

	defer func() { _ = noop.Close() }()

	_, _ = noop.WriteString("package dnslight\n")

	for _, provider := range info.Providers {
		err = generateRegisterProvider(dir, provider)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateRegisterProvider(dir string, provider descriptors.Provider) error {
	file, err := os.Create(filepath.Join(dir, fmt.Sprintf("zz_gen_%s.go", cleanName(provider.Code))))
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	err = generateFile(provider, srcCustomTemplate, file)
	if err != nil {
		return err
	}

	return nil
}

func cleanName(src string) string {
	return strings.ReplaceAll(src, "-", "")
}

func generateFile(data any, tmpl string, file *os.File) error {
	b := &bytes.Buffer{}

	err := template.Must(
		template.New("").
			Funcs(map[string]any{
				"cleanName": cleanName,
			}).Parse(tmpl),
	).Execute(b, data)
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

	return nil
}
