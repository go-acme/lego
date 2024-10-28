package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"

	hcversion "github.com/hashicorp/go-version"
	"github.com/urfave/cli/v2"
)

const flgMode = "mode"

const (
	modePatch = "patch"
	modeMinor = "minor"
	modeMajor = "major"
)

const versionSourceFile = "./cmd/lego/zz_gen_version.go"

const (
	commentRelease = "release"
	commentDetach  = "detach"
)

func main() {
	app := cli.NewApp()
	app.Name = "lego-releaser"
	app.Usage = "Lego releaser"
	app.HelpName = "releaser"
	app.Commands = []*cli.Command{
		{
			Name:   "release",
			Usage:  "Update file for a release",
			Action: release,
			Before: func(ctx *cli.Context) error {
				mode := ctx.String("mode")
				switch mode {
				case modePatch, modeMinor, modeMajor:
					return nil
				default:
					return fmt.Errorf("invalid mode: %s", mode)
				}
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flgMode,
					Aliases: []string{"m"},
					Value:   modePatch,
					Usage:   fmt.Sprintf("The release mode: %s|%s|%s", modePatch, modeMinor, modeMajor),
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
	mode := ctx.String(flgMode)

	currentVersion, err := readCurrentVersion(versionSourceFile)
	if err != nil {
		return fmt.Errorf("read current version: %w", err)
	}

	nextVersion, err := bumpVersion(mode, currentVersion)
	if err != nil {
		return fmt.Errorf("bump version: %w", err)
	}

	err = generate(nextVersion, commentRelease)
	if err != nil {
		return err
	}

	return nil
}

func detach(_ *cli.Context) error {
	currentVersion, err := readCurrentVersion(versionSourceFile)
	if err != nil {
		return fmt.Errorf("read current version: %w", err)
	}

	v := currentVersion.Core().String()

	err = generate(v, commentDetach)
	if err != nil {
		return err
	}

	return nil
}

func readCurrentVersion(filename string) (*hcversion.Version, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	v := visitor{data: make(map[string]string)}
	ast.Walk(v, file)

	current, err := hcversion.NewSemver(v.data["defaultVersion"])
	if err != nil {
		return nil, err
	}

	return current, nil
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

func bumpVersion(mode string, v *hcversion.Version) (string, error) {
	segments := v.Segments()

	switch mode {
	case modePatch:
		return fmt.Sprintf("%d.%d.%d", segments[0], segments[1], segments[2]+1), nil
	case modeMinor:
		return fmt.Sprintf("%d.%d.0", segments[0], segments[1]+1), nil
	case modeMajor:
		return fmt.Sprintf("%d.0.0", segments[0]+1), nil
	default:
		return "", fmt.Errorf("invalid mode: %s", mode)
	}
}
