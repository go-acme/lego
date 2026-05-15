package internal

/*
	Inspired by https://github.com/urfave/cli/blob/v3.9.0/help.go
*/

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/go-acme/lego/v5/log"
	"github.com/urfave/cli/v3"
)

// PrintMarkdown writes markdown documentation.
func PrintMarkdown(out io.Writer, templ string, data any) {
	funcMap := template.FuncMap{
		"join":           strings.Join,
		"GetFlagNames":   GetFlagNames,
		"GetEnvVars":     GetEnvVars,
		"GetDefaultText": GetDefaultText,
		"GetUsage":       GetUsage,
	}

	t := template.Must(template.New("markdown").Funcs(funcMap).Parse(templ))

	for name, src := range map[string]string{
		"markdownUsageTemplate":           markdownUsageTemplate,
		"markdownDescriptionTemplate":     markdownDescriptionTemplate,
		"markdownVersionTemplate":         markdownVersionTemplate,
		"markdownCopyrightTemplate":       markdownCopyrightTemplate,
		"markdownAuthorsTemplate":         markdownAuthorsTemplate,
		"markdownCommandsTemplate":        markdownCommandsTemplate,
		"markdownFlagCategoriesTemplate":  markdownFlagCategoriesTemplate,
		"markdownFlagsTemplate":           markdownFlagsTemplate,
		"markdownPersistentFlagsTemplate": markdownPersistentFlagsTemplate,
	} {
		if _, err := t.New(name).Parse(src); err != nil {
			log.Error("Failed to parse template", log.ErrorAttr(err), slog.String("name", name))
		}
	}

	err := t.Execute(out, data)
	if err != nil {
		log.Error("Failed to execute template", log.ErrorAttr(err))
	}
}

func GetFlagNames(f cli.Flag) string {
	df, ok := f.(cli.DocGenerationFlag)
	if !ok {
		names := make([]string, len(f.Names()))
		for i, n := range f.Names() {
			names[i] = "`" + prefixFor(n) + n + "`"
		}

		return strings.Join(names, ", ")
	}

	placeholder := ""

	if df.TakesValue() {
		if tname := df.TypeName(); tname != "" {
			placeholder = tname
		} else {
			placeholder = "value"
		}
	}

	names := make([]string, 0, len(f.Names()))
	for _, name := range f.Names() {
		if name == "" {
			continue
		}

		pf := prefixFor(name)
		if placeholder != "" {
			names = append(names, fmt.Sprintf("`%s%s %s`", pf, name, placeholder))
		} else {
			names = append(names, fmt.Sprintf("`%s%s`", pf, name))
		}
	}

	return strings.Join(names, ", ")
}

func GetEnvVars(df cli.DocGenerationFlag) string {
	msg := new(strings.Builder)

	for i, s := range df.GetEnvVars() {
		_, _ = fmt.Fprintf(msg, "`%s`", s)
		if i < len(df.GetEnvVars())-1 {
			_, _ = fmt.Fprintf(msg, ", ")
		}
	}

	return msg.String()
}

func GetUsage(df cli.DocGenerationFlag) string {
	return strings.ReplaceAll(df.GetUsage(), "\n", "<br>")
}

func GetDefaultText(df cli.DocGenerationFlag) string {
	if df.IsDefaultVisible() {
		if s := df.GetDefaultText(); s != "" {
			return s
		} else if df.TakesValue() && df.GetValue() != "" {
			return df.GetValue()
		}
	}

	return ""
}

// prefixFor returns "-" for single-character flag names and "--" otherwise.
func prefixFor(name string) string {
	if len(name) == 1 {
		return "-"
	}

	return "--"
}
