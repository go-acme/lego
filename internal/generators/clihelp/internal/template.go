package internal

/*
	Inspired by https://github.com/urfave/cli/blob/v3.9.0/template.go
*/

const (
	markdownUsageTemplate = `{{if .UsageText}}{{.UsageText}}{{else}}{{.FullName}}` +
		`{{if .VisibleFlags}} [options]{{end}}` +
		`{{if .VisibleCommands}} [command [command options]]{{end}}` +
		`{{if .ArgsUsage}} {{.ArgsUsage}}{{end}}{{end}}`

	markdownDescriptionTemplate = `{{if .Description}}
### Description

{{.Description}}
{{end}}`

	markdownVersionTemplate = `{{if .Version}}{{if not .HideVersion}}
**Version:** {{.Version}}
{{end}}{{end}}`

	markdownCopyrightTemplate = `{{if .Copyright}}
### Copyright

{{.Copyright}}
{{end}}`

	markdownAuthorsTemplate = `{{with $length := len .Authors}}{{if ne 1 $length}}
### Authors

{{range $.Authors}}  - {{.}}
{{end}}{{else}}
### Author

{{range $.Authors}}  - {{.}}
{{end}}{{end}}{{end}}`

	markdownCommandsTemplate = `{{if .VisibleCommands}}
### Commands

| Name | Usage |
|------|-------|
{{range .VisibleCommands}}| ` + "`" + `{{join .Names ", "}}` + "`" + ` | {{.Usage}} |
{{end}}{{end}}`

	markdownFlagCategoriesTemplate = `{{range .VisibleFlagCategories}}{{if .Name}}#### {{.Name}}

{{end}}| Flag | Env Var | Usage |
|------|-------|-------|
{{range .Flags}}| {{ GetFlagNames .}} | {{ GetEnvVars . }} | {{ GetUsage . }} {{ $def := GetDefaultText . }}{{if $def }}<br> (Default: {{ $def }}){{end}} |
{{end}}
{{end}}`

	markdownFlagsTemplate = `| Flag | Env Var | Usage |
|------|-------|-------|
{{range .VisibleFlags}}| {{ GetFlagNames .}} | {{ GetEnvVars . }} | {{ GetUsage . }} {{ $def := GetDefaultText . }}{{if $def }}<br> (Default: {{ $def }}){{end}} |
{{end}}`

	markdownPersistentFlagsTemplate = `| Flag | Env Var | Usage |
|------|-------|-------|
{{range .VisiblePersistentFlags}}| {{ GetFlagNames .}} | {{ GetEnvVars . }} | {{ GetUsage . }} {{ $def := GetDefaultText . }}{{if $def }}<br> (Default: {{ $def }}){{end}} |
{{end}}`
)

const RootCommandHelpTemplate = `## ` + "`" + `{{.FullName}}` + "`" + `{{if .Usage}}

> {{.Usage}}
{{end}}{{template "markdownVersionTemplate" .}}
### Usage

` + "```" + `
{{template "markdownUsageTemplate" .}}
` + "```" + `
{{template "markdownDescriptionTemplate" .}}{{- if .Authors}}
{{template "markdownAuthorsTemplate" .}}{{end}}{{template "markdownCommandsTemplate" .}}{{if .VisibleFlagCategories}}
### Global Options

{{template "markdownFlagCategoriesTemplate" .}}{{else if .VisibleFlags}}
### Global Options

{{template "markdownFlagsTemplate" .}}{{end}}{{template "markdownCopyrightTemplate" .}}`

const CommandHelpTemplate = `## ` + "`" + `{{.FullName}}` + "`" + `{{if .Usage}}

> {{.Usage}}
{{end}}
### Usage

` + "```" + `
{{template "markdownUsageTemplate" .}}
` + "```" + `
{{if .Category}}**Category:** {{.Category}}

{{end}}{{template "markdownDescriptionTemplate" .}}{{if .VisibleFlagCategories}}
### Options

{{template "markdownFlagCategoriesTemplate" .}}{{else if .VisibleFlags}}
### Options

{{template "markdownFlagsTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}
### Global Options

{{template "markdownPersistentFlagsTemplate" .}}{{end}}`

const SubcommandHelpTemplate = `## ` + "`" + `{{.FullName}}` + "`" + `{{if .Usage}}

> {{.Usage}}
{{end}}
### Usage

` + "```" + `
{{if .UsageText}}{{.UsageText}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{end}}{{end}}
` + "```" + `
{{if .Category}}**Category:** {{.Category}}

{{end}}{{template "markdownDescriptionTemplate" .}}{{template "markdownCommandsTemplate" .}}{{if .VisibleFlagCategories}}
### Options

{{template "markdownFlagCategoriesTemplate" .}}{{else if .VisibleFlags}}
### Options

{{template "markdownFlagsTemplate" .}}{{end}}`
