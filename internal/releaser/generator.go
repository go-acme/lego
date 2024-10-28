package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

const (
	dnsTemplate   = "templates/dns.go.tmpl"
	dnsTargetFile = "./providers/dns/internal/useragent/useragent.go"
)

const (
	senderTemplate   = "templates/sender.go.tmpl"
	senderTargetFile = "./acme/api/internal/sender/useragent.go"
)

const (
	versionTemplate   = "templates/version.go.tmpl"
	versionTargetFile = "./cmd/lego/zz_gen_version.go"
)

//go:embed templates
var templateFS embed.FS

type Generator struct {
	templatePath string
	targetFile   string
}

func NewGenerator(templatePath string, targetFile string) *Generator {
	return &Generator{templatePath: templatePath, targetFile: targetFile}
}

func (g *Generator) Generate(version, comment string) error {
	tmpl, err := template.New(filepath.Base(g.templatePath)).ParseFS(templateFS, g.templatePath)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}

	err = tmpl.Execute(b, map[string]string{
		"version": version,
		"comment": comment,
	})
	if err != nil {
		return err
	}

	source, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(g.targetFile, source, 0o644)
}

func generate(targetVersion, comment string) error {
	generators := []*Generator{
		NewGenerator(dnsTemplate, dnsTargetFile),
		NewGenerator(senderTemplate, senderTargetFile),
		NewGenerator(versionTemplate, versionTargetFile),
	}

	for _, generator := range generators {
		err := generator.Generate(targetVersion, comment)
		if err != nil {
			return fmt.Errorf("generate file(s): %w", err)
		}
	}

	return nil
}
