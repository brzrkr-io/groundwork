# Text templates: render an nginx upstream block

Config files are mostly fixed text with a few values that change per host or
environment. Go's `text/template` package fills those values in for you. Helm
charts, Hugo and many deploy tools are built on it.

**New here:** `text/template` (`Parse`, `Execute`, dot, `range`, `if`, trim
markers, `FuncMap`), `strings.Builder`.

## Learn

### What a template is

A *template* is normal text with *actions* inside `{{ }}`. Text outside the
braces is copied as-is. Actions are replaced by data.

```text
Hello, {{.}}!      with data "ops"   =>   Hello, ops!
```

You need `import "text/template"` (not `html/template`, which is for web
pages).

### Step 1: Parse

`Parse` turns the text into a template you can run. It fails if the text is
broken, for example an action that never closes:

```go
tmpl, err := template.New("motd").Parse("Welcome to {{.}}")
if err != nil {
	// e.g. "template: motd:1: unclosed action" for "{{."
}
```

`template.New(name)` makes an empty template. The name only shows up in error
messages, like `motd` above.

> **Why is it written like this?** `template.New(...).Parse(...)` is a
> *method chain*: `New` returns a template, and you call `Parse` on it right
> away. Go allows it; it's the usual style for this package. The name
> `"motd"` is your choice.

### Step 2: Execute into a writer

`Execute(w, data)` runs the template and writes the result to `w`. `w` can be
any `io.Writer`, the same idea as the readers in *Reading input line by line*.
Printing straight to the terminal:

```go
err = tmpl.Execute(os.Stdout, "db-1") // prints "Welcome to db-1"
```

`Execute` fails when the data doesn't fit, for example `{{.Missing}}` on a
struct with no `Missing` field.

### Collecting text: strings.Builder

To get the output as a `string` instead of printing it, write into a
`strings.Builder`, then call `.String()`:

```go
var b strings.Builder
fmt.Fprint(&b, "hello ")
fmt.Fprint(&b, "world")
b.String() // "hello world"
```

> **Why is it written like this?** You pass `&b` (a pointer) so the function
> adds to *your* builder, not to a copy. Go requires the pointer here: a
> plain `b` isn't an `io.Writer`.

### Dot

Inside a template, `.` (*dot*) is the data you passed in.

| Action | Data | Output |
|---|---|---|
| `{{.}}` | `"ops"` | `ops` |
| `{{.Name}}` | `Upstream{Name: "api"}` | `api` (a struct field, must be exported) |
| `{{.host}}` | `map[string]string{"host": "db1"}` | `db1` (a map key) |

### Loops: range

`{{range .List}} ... {{end}}` repeats the middle once per element. **Inside
the loop, dot is the current element.** To reach the top-level data from
inside, use `$`, which always means "the data given to `Execute`":

```text
{{range .Users}}{{.}}@{{$.Domain}} {{end}}
Users: ["ann", "bo"], Domain: "corp"   =>   "ann@corp bo@corp "
```

An empty or nil slice runs the body zero times.

### Conditions: if / else

```text
{{if .Tags}}tagged{{else}}untagged{{end}}
```

`if` is false for an *empty value*: `false`, `0`, `""`, nil, or a slice or
map with length 0. Everything else is true.

### Trimming whitespace

Newlines outside `{{ }}` are copied too, so loops easily make blank lines. A
dash trims all spaces and newlines on that side: `{{-` trims before the
action, `-}}` trims after it. Write a space next to the dash: `{{- .X}}`.

Without trimming, with `Users: ["ann", "bo"]`:

```text
{{range .Users}}
user {{.}}
{{end}}
```

gives `"\nuser ann\n\nuser bo\n"`, full of blank lines. With `{{-` on the
action lines, the newline *in front of* each action disappears:

```text
users:
{{- range .Users}}
user {{.}}
{{- end}}
```

gives `"users:\nuser ann\nuser bo"`.

> **Why is it written like this?** The space after `{{-` is required: Go
> would read `{{-3}}` as the number -3. Putting each action on its own line
> is the author's choice; it keeps the template readable.

### Functions: FuncMap

Templates call functions by name, arguments after it, separated by spaces:
`{{upper .Name}}`, or `{{printf "%05d" .ID}}` with two arguments. `len` and
`printf` are built in. Your own functions go in a `template.FuncMap` (a map
from name to Go function), registered with `Funcs` **before** `Parse`:

```go
funcs := template.FuncMap{"lower": strings.ToLower}
tmpl, err := template.New("x").Funcs(funcs).Parse("{{lower .}}")
```

If the text uses a function the FuncMap doesn't have, `Parse` fails with
`function "lower" not defined`.

> **Why is it written like this?** No parentheses or commas around the
> arguments: that's template syntax, not Go syntax. The template language
> requires it.

### template.Must

`template.Must(...)` panics if `Parse` fails. People use it for templates
parsed once at package level. Here the templates are parsed inside functions
instead, so a mistake fails one test, not all of them.

## Your task

Your team generates nginx config for its load balancers. Finish the renderer.

### 1. Render a template

**Goal:** `Render(name, text, data)` parses `text`, runs it with `data`, and
returns the output as a string.

- Call the template `name` (the parameter, not a fixed word).
- If parsing **or** executing fails, return `""` and the error.

```go
Render("test", "Hello, {{.}}!", "ops")
// => "Hello, ops!", nil

Render("test", "{{.Name}} listens on {{.Port}}", Upstream{Name: "api", Port: 8080})
// => "api listens on 8080", nil

Render("test", "{{.host}}:{{.port}}", map[string]string{"host": "db1", "port": "5432"})
// => "db1:5432", nil

Render("test", "{{.Name", Upstream{Name: "api"})
// => "", error: template: test:1: unclosed action

Render("test", "{{range .Servers}}x", Upstream{})
// => "", error (range has no {{end}})

Render("test", "{{.Missing}}", Upstream{Name: "api"})
// => "", error: ... can't evaluate field Missing in type render.Upstream
```

**Think it through:**
- Two steps can fail. Where do you check each error?
- `Summary`, further down the file, already does this job. What's different
  about where it gets its name and text?

### 2. Write the upstream block

**Goal:** change the `upstreamTmpl` constant so it writes one `server` line
per address in `Servers`.

`RenderUpstream` is already written; it calls your `Render` with
`upstreamTmpl`. Each `server` line:

- starts with exactly 4 spaces, then `server `,
- then the address, `:`, and the `Port` of the whole `Upstream`,
- ends with `;`.

Every line, including the last `}`, ends with a newline. No blank lines.

```go
RenderUpstream(Upstream{Name: "api", Port: 8080, Servers: []string{"10.0.0.1", "10.0.0.2"}})
// => "upstream api {\n    server 10.0.0.1:8080;\n    server 10.0.0.2:8080;\n}\n"
```

Printed, that is:

```text
upstream api {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
}
```

With no servers: `"upstream empty {\n}\n"`.

**Think it through:**
- Inside the loop, dot is one address. How do you reach `Port`?
- Which newlines do you need to trim so there are no blank lines? Look at the
  `users:` example.

### 3. Write a one-line summary

**Goal:** make `Summary` return `NAME: server, server`, or `NAME: no
servers` when there are none.

You change two things; the parsing code is already written.

1. Add a `"join"` entry for `strings.Join` to `summaryFuncs`. Do this
   **first**, or every case fails with `function "join" not defined`.
2. Change `summaryTmpl` to output:
   - the name in upper case (the `upper` function is already registered),
   - fixed text `: ` (colon, space),
   - the servers joined by `", "` (comma, space), or the words `no servers`
     when `Servers` is nil or empty.

No newline at the end.

```go
Summary(Upstream{Name: "api", Servers: []string{"10.0.0.1", "10.0.0.2"}})
// => "API: 10.0.0.1, 10.0.0.2"

Summary(Upstream{Name: "web", Servers: []string{"web-1.internal"}})
// => "WEB: web-1.internal"

Summary(Upstream{Name: "empty"})                       // nil Servers
// => "EMPTY: no servers"

Summary(Upstream{Name: "drained", Servers: []string{}}) // empty Servers
// => "DRAINED: no servers"
```

**Think it through:**
- `strings.Join` takes two arguments. In what order, and how do you pass
  them in template syntax?
- Which action picks between two outputs?

## Learn more

- [text/template package](https://pkg.go.dev/text/template)
- [Go by Example: Text Templates](https://gobyexample.com/text-templates)
- [nginx: the upstream module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
