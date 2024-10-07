package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type Generator struct {
	baseUserAgent string
	template      string
	sourcePath    string
}

func NewGenerator(baseUserAgent string, tmpl string, sourcePath string) *Generator {
	return &Generator{baseUserAgent: baseUserAgent, template: tmpl, sourcePath: sourcePath}
}

func (g *Generator) Release(mode string) error {
	// Read file
	data, err := readUserAgentFile(g.sourcePath)
	if err != nil {
		return err
	}

	// Bump version
	newVersion, err := g.bumpVersion(data["ourUserAgent"], mode)
	if err != nil {
		return err
	}

	// Write file
	comment := "release" // detach|release

	return g.writeUserAgentFile(g.sourcePath, newVersion, comment)
}

func (g *Generator) Detach() error {
	// Read file
	data, err := readUserAgentFile(g.sourcePath)
	if err != nil {
		return err
	}

	// Write file
	version := strings.TrimPrefix(data["ourUserAgent"], g.baseUserAgent)
	comment := "detach"

	return g.writeUserAgentFile(g.sourcePath, version, comment)
}

func (g *Generator) writeUserAgentFile(filename, version, comment string) error {
	tmpl, err := template.New("ua").Parse(g.template)
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

	return os.WriteFile(filename, source, 0o644)
}

func (g *Generator) bumpVersion(userAgent, mode string) (string, error) {
	prevVersion := strings.TrimPrefix(userAgent, g.baseUserAgent)

	allString := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`).FindStringSubmatch(prevVersion)

	if len(allString) != 4 {
		return "", fmt.Errorf("invalid version format: %s", prevVersion)
	}

	switch mode {
	case "patch":
		patch, err := strconv.Atoi(allString[3])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%s.%d", allString[1], allString[2], patch+1), nil
	case "minor":
		minor, err := strconv.Atoi(allString[2])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%d.0", allString[1], minor+1), nil
	case "major":
		major, err := strconv.Atoi(allString[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d.0.0", major+1), nil
	default:
		return "", fmt.Errorf("invalid mode: %s", mode)
	}
}

func readUserAgentFile(filename string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	v := visitor{data: make(map[string]string)}
	ast.Walk(v, file)

	return v.data, nil
}

type visitor struct {
	data map[string]string
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch d := n.(type) {
	case *ast.GenDecl:
		if d.Tok == token.CONST {
			for _, spec := range d.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
					continue
				}

				va, ok := valueSpec.Values[0].(*ast.BasicLit)
				if !ok {
					continue
				}
				if va.Kind != token.STRING {
					continue
				}

				s, err := strconv.Unquote(va.Value)
				if err != nil {
					continue
				}

				v.data[valueSpec.Names[0].String()] = s
			}
		}
	default:
		// noop
	}
	return v
}
