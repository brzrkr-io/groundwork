# Text templates: render an nginx upstream block

Config files are mostly fixed text with a few values that change per host,
service or environment. Go's `text/template` package fills those values in for
you. Helm charts, Hugo and many internal deploy tools are built on it.

## Learn

### Parse, then execute

A *template* is text with *actions* inside `{{ }}`. You turn the text into a
template once with `Parse`, then run it with `Execute`, passing the data to
fill in:

```go
tmpl, err := template.New("greeting").Parse("Hello, {{.}}!")
if err != nil {
	return "", err // the template text itself is broken
}
var b strings.Builder
err = tmpl.Execute(&b, "ops") // b now holds "Hello, ops!"
```

- `template.New(name)` creates an empty template. The name shows up in error
  messages.
- `Execute(w, data)` writes its output to any `io.Writer`, just like the
  readers from *Reading input line by line* read from any `io.Reader`. A
  `*strings.Builder` is a writer that collects text in memory; call
  `b.String()` to get it. Pass `&b`, a pointer, so `Execute` can add to it.
- Both steps can fail. `Parse` fails when the text is malformed, for example
  `{{.Name` (`template: test:1: unclosed action`). `Execute` fails when the
  data doesn't fit, for example `{{.Missing}}` on a struct with no `Missing`
  field.

You need `import "text/template"` (not `html/template`, which is for web
pages).

### Dot

Inside a template, `.` (*dot*) is the current data.

| Action | Data | Output |
|---|---|---|
| `{{.}}` | `"ops"` | `ops` |
| `{{.Name}}` | `Upstream{Name: "api"}` | `api` (an exported struct field) |
| `{{.host}}` | `map[string]string{"host": "db1"}` | `db1` (a map key) |

### Loops: range

`{{range .Servers}} ... {{end}}` repeats its body once per element. **Inside
the range, dot is the current element**, so `{{.}}` is one server. To reach
the top-level data from inside the loop, use `$`, which always means "the data
passed to `Execute`":

```go
{{range .Servers}}{{.}}:{{$.Port}} {{end}}
// Servers: ["a", "b"], Port: 80  =>  "a:80 b:80 "
```

An empty or nil slice runs the body zero times.

### Conditions: if / else

```go
{{if .Servers}}has servers{{else}}no servers{{end}}
```

`if` is false for the *empty value* of a type: `false`, `0`, `""`, a nil
pointer, and a nil or zero-length slice or map. Everything else is true.

### Trimming whitespace

Every character outside `{{ }}` is copied to the output, including newlines.
That makes loops produce blank lines. A dash next to the braces trims all
whitespace (spaces and newlines) on that side: `{{-` trims before the action,
`-}}` trims after it. The dash needs a space next to it: `{{- .X}}`.

Without trimming, with `Servers: ["a", "b"]`:

```text
{{range .Servers}}
server {{.}};
{{end}}
```

gives `"\nserver a;\n\nserver b;\n"`: blank lines everywhere. Trimming the
newline *before* each action fixes it:

```text
servers:
{{- range .Servers}}
server {{.}};
{{- end}}
```

gives `"servers:\nserver a;\nserver b;"`. Each loop pass now outputs exactly
one `"\n"` plus one line.

### Functions: FuncMap

Templates can call functions: `{{upper .Name}}` calls `upper` with `.Name`
as its argument, and `{{join .Servers ", "}}` passes two arguments. Built-in
functions such as `len` and `printf` always work. Your own functions go in a
`template.FuncMap`, a map from name to Go function, registered with `Funcs`
**before** `Parse`:

```go
funcs := template.FuncMap{"upper": strings.ToUpper}
tmpl, err := template.New("x").Funcs(funcs).Parse("{{upper .}}")
```

If the text calls a function the FuncMap doesn't have, `Parse` fails:

```text
template: summary:1: function "join" not defined
```

### template.Must

When the template text is a constant in your program, you often parse it
once at package level with `template.Must`, which panics if `Parse` returned
an error:

```go
var motd = template.Must(template.New("motd").Parse("Welcome to {{.}}\n"))
```

That is fine for finished code, but a panic at package level crashes every
test at once. In this exercise the templates are parsed inside functions
instead, so a mistake shows up as an error in one test.

## Your task

Your team generates nginx config for its load balancers. Finish the
renderer.

### 1. Render a template

Implement `Render(name, text string, data any) (string, error)`. Parse
`text` into a template called `name`, execute it with `data` into a
`strings.Builder`, and return the output. If parsing or executing fails,
return `""` and the error.

```go
Render("test", "Hello, {{.}}!", "ops")
// => "Hello, ops!", nil

Render("test", "{{.Name}} listens on {{.Port}}", Upstream{Name: "api", Port: 8080})
// => "api listens on 8080", nil

Render("test", "{{.Name", Upstream{Name: "api"})
// => "", error: template: test:1: unclosed action

Render("test", "{{.Missing}}", Upstream{Name: "api"})
// => "", error: ... can't evaluate field Missing in type render.Upstream
```

### 2. Write the upstream block

`RenderUpstream` is already written: it calls your `Render` with the
`upstreamTmpl` constant. Change `upstreamTmpl` so that it writes one
`server` line per entry in `Servers`, with `:Port` after each address. Each
`server` line starts with exactly 4 spaces and ends with `;`. Every line,
including the last `}`, ends with a newline, and there are no blank lines.

```go
RenderUpstream(Upstream{Name: "api", Port: 8080, Servers: []string{"10.0.0.1", "10.0.0.2"}})
```

returns this text (with a newline after the `}`):

```text
upstream api {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
}
```

As a Go string: `"upstream api {\n    server 10.0.0.1:8080;\n    server 10.0.0.2:8080;\n}\n"`.

With no servers, the result is `"upstream empty {\n}\n"`:

```text
upstream empty {
}
```

### 3. Write a one-line summary

`Summary` renders `summaryTmpl` for a status page. The parsing code is
written; you change the FuncMap and the template.

1. First add `"join": strings.Join` to `summaryFuncs`. Do this **before**
   using `join` in the template, or every case fails with
   `function "join" not defined`.
2. Then change `summaryTmpl` to output the name in upper case (use the
   `upper` function that is already registered), a colon and a space, and
   then either the servers joined by `", "` or, when `Servers` is empty or
   nil, the words `no servers`. There is no newline at the end.

```go
Summary(Upstream{Name: "api", Servers: []string{"10.0.0.1", "10.0.0.2"}})
// => "API: 10.0.0.1, 10.0.0.2"

Summary(Upstream{Name: "web", Servers: []string{"web-1.internal"}})
// => "WEB: web-1.internal"

Summary(Upstream{Name: "empty"})
// => "EMPTY: no servers"
```

## Learn more

- [text/template package](https://pkg.go.dev/text/template)
- [Go by Example: Text Templates](https://gobyexample.com/text-and-templates)
- [nginx: the upstream module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
