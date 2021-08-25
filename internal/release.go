package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/urfave/cli"
)

const sourceFile = "./acme/api/internal/sender/useragent.go"

const uaTemplate = `package sender

// CODE GENERATED AUTOMATICALLY
// THIS FILE MUST NOT BE EDITED BY HAND

const (
	// ourUserAgent is the User-Agent of this underlying library package.
	ourUserAgent = "xenolf-acme/{{ .version }}"

	// ourUserAgentComment is part of the UA comment linked to the version status of this underlying library package.
	// values: detach|release
	// NOTE: Update this with each tagged release.
	ourUserAgentComment = "{{ .comment }}"
)

`

func main() {
	app := cli.NewApp()
	app.Name = "lego-releaser"
	app.Usage = "Lego releaser"
	app.HelpName = "releaser"
	app.Commands = []cli.Command{
		{
			Name:   "release",
			Usage:  "Update file for a release",
			Action: release,
			Before: func(ctx *cli.Context) error {
				mode := ctx.String("mode")
				switch mode {
				case "patch", "minor", "major":
					return nil
				default:
					return fmt.Errorf("invalid mode: %s", mode)
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "mode, m",
					Value: "patch",
					Usage: "The release mode: patch|minor|major",
				},
			},
		},
		{
			Name:   "detach",
			Usage:  "Update file post release",
			Action: detach,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func release(ctx *cli.Context) error {
	mode := ctx.String("mode")

	// Read file
	data, err := readUserAgentFile(sourceFile)
	if err != nil {
		return err
	}

	// Bump version
	newVersion, err := bumpVersion(data["ourUserAgent"], mode)
	if err != nil {
		return err
	}

	// Write file
	comment := "release" // detach|release
	return writeUserAgentFile(sourceFile, newVersion, comment)
}

func detach(_ *cli.Context) error {
	// Read file
	data, err := readUserAgentFile(sourceFile)
	if err != nil {
		return err
	}

	// Write file
	version := strings.TrimPrefix(data["ourUserAgent"], "xenolf-acme/")
	comment := "detach"
	return writeUserAgentFile(sourceFile, version, comment)
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

func writeUserAgentFile(filename, version, comment string) error {
	tmpl, err := template.New("ua").Parse(uaTemplate)
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

func bumpVersion(userAgent, mode string) (string, error) {
	prevVersion := strings.TrimPrefix(userAgent, "xenolf-acme/")

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
