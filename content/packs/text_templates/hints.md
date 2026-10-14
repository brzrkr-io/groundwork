## Task 1: Render a template

This is the "Parse, then execute" example from the README. There are two
steps, and each can fail, so check the error after both. `Summary` further
down the file does the same job.

### Closer

`template.New(name).Parse(text)` gives you the template. Then make a
`var b strings.Builder` and call `tmpl.Execute(&b, data)`. On success,
return `b.String()`.

### Answer

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

## Task 2: Write the upstream block

Use a `range` over `.Servers`. Inside it, dot is one address, so how do you
reach the `Port` of the whole `Upstream`? Then think about where the newlines
around each action end up.

### Closer

Inside `{{range .Servers}}`, `{{.}}` is the address and `{{$.Port}}` is the
port (`$` is the top-level data). Put `{{- range .Servers}}` and `{{- end}}`
each on their own line, so the dash eats the newline in front of them.

### Answer

```go
const upstreamTmpl = `upstream {{.Name}} {
{{- range .Servers}}
    server {{.}}:{{$.Port}};
{{- end}}
}
`
```

Each trimmed action removes its own line's newline, so only the `server` lines
and the closing `}` remain.

## Task 3: Write a one-line summary

Register `join` first, then build the template from three parts: the
upper-cased name, `: `, and an `if`/`else` that picks between the joined list
and `no servers`.

### Closer

Add `"join": strings.Join,` to `summaryFuncs`. Functions are called with their
arguments after them: `{{upper .Name}}`, `{{join .Servers ", "}}`. An empty or
nil slice counts as false in `{{if .Servers}}`.

### Answer

```go
"join":  strings.Join,
```

```go
const summaryTmpl = `{{upper .Name}}: {{if .Servers}}{{join .Servers ", "}}{{else}}no servers{{end}}`
```
