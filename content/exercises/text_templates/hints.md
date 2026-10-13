## Task 1: two steps, two errors

`Render` is the "Parse, then execute" example from the README. Check the error
after `Parse` and again after `Execute`, and return `""` with it each time.
Look at how `Summary` does the same job further down the file.

## Task 1: almost the answer

```go
tmpl, err := template.New(name).Parse(text)
if err != nil {
	return "", err
}
var b strings.Builder
if err := tmpl.Execute(&b, data); err != nil {
	return "", err
}
return b.String(), nil
```

## Task 2: range with trim markers

Inside `{{range .Servers}}`, `{{.}}` is one address and `{{$.Port}}` is the
port. Put `{{- range .Servers}}` and `{{- end}}` each on their own line, so
the dash eats the newline in front of them:

```go
const upstreamTmpl = `upstream {{.Name}} {
{{- range .Servers}}
    server {{.}}:{{$.Port}};
{{- end}}
}
`
```

## Task 3: if, else and join

Add `"join": strings.Join,` to `summaryFuncs` first. Then:

```go
`{{upper .Name}}: {{if .Servers}}{{join .Servers ", "}}{{else}}no servers{{end}}`
```
