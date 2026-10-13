# Text templates: systemd unit files

Every long-running service on a Linux host is described by a systemd *unit
file*. Here you fix and finish the templates a provisioning tool uses to write
them.

**Heads-up:** the first test fails with a template parse error until you fix
task 1. That is part of the exercise.

## Learn

### Recap

`template.New(name).Parse(text)` turns text into a template, and
`Execute(&b, data)` writes it into a `strings.Builder`. `{{.Field}}` prints a
field, `{{range}}...{{end}}` loops, `{{if}}...{{else}}...{{end}}` branches on
empty vs non-empty values, and `{{-` / `-}}` trim whitespace on that side of
an action. The `render` helper in `exercise.go` does the parse-and-execute
part for you and returns any error.

### Reading a parse error

```text
template: unit:5: unexpected EOF
```

This means: in the template named `unit`, at line 5, the text ended
(*EOF*, end of file) while the parser was still waiting for something.
Every `range`, `if` and `with` needs its own `{{end}}`; a missing one gives
exactly this error.

### with

`{{with .User}}...{{end}}` runs its body only when `.User` is non-empty, and
**inside the body, dot is set to `.User`**. It is a shorter `if` for "print
this value if there is one":

```go
{{with .User}}User={{.}}{{end}}
// User: "svc-payments" => "User=svc-payments"
// User: ""             => ""
```

`with` can have an `{{else}}` too, which runs when the value is empty. That
gives you a default:

```go
{{with .Restart}}{{.}}{{else}}on-failure{{end}}
// Restart: "always" => "always"
// Restart: ""       => "on-failure"
```

### Ranging over a map

`range` can assign each element to *variables*. Template variables start
with `$`. Over a map you get the key and the value:

```go
{{range $key, $value := .}}{{$key}}={{$value}};{{end}}
// map[string]string{"b": "2", "a": "1"} => "a=1;b=2;"
```

Go maps have no order, but a template's `range` over a map **always visits
the keys in sorted order**. So the output is the same every time, which is
exactly what you want in a config file.

## Your task

A unit file for a payments service looks like this:

```text
[Unit]
Description=Payments API
After=network.target
After=postgres.service

[Service]
ExecStart=/usr/local/bin/payments --port=8080
User=svc-payments
Restart=always
Environment="DB_HOST=db.internal"
Environment="LOG_LEVEL=info"
```

Each task below fixes one template constant. Every expected output is shown
exactly; each line ends with a newline (`\n`), including the last one, and
there are never blank lines.

### 1. Fix the [Unit] template

`unitTmpl` has two bugs. First, its `range` has no `{{end}}`, so parsing
fails with `template: unit:5: unexpected EOF`. Once you add the `{{end}}`,
you'll see the second bug: blank lines between the `After=` lines. Fix both
so there is one `After=` line per entry in `After`.

```go
UnitSection(Unit{Description: "Payments API", After: []string{"network.target", "postgres.service"}})
```

returns:

```text
[Unit]
Description=Payments API
After=network.target
After=postgres.service
```

As a Go string: `"[Unit]\nDescription=Payments API\nAfter=network.target\nAfter=postgres.service\n"`.
With no `After` entries, `Unit{Description: "Cache warmer"}` gives
`"[Unit]\nDescription=Cache warmer\n"`.

### 2. Make User optional and give Restart a default

Change `serviceTmpl`:

- Write the `User=` line only when `User` is non-empty. When it is empty,
  leave the whole line out (no blank line either).
- When `Restart` is empty, write `Restart=on-failure`.

```go
ServiceSection(Service{ExecStart: "/usr/local/bin/payments --port=8080", User: "svc-payments", Restart: "always"})
// => "[Service]\nExecStart=/usr/local/bin/payments --port=8080\nUser=svc-payments\nRestart=always\n"

ServiceSection(Service{ExecStart: "/usr/bin/true"})
// => "[Service]\nExecStart=/usr/bin/true\nRestart=on-failure\n"
```

The second one, as text:

```text
[Service]
ExecStart=/usr/bin/true
Restart=on-failure
```

### 3. Write the Environment lines

`envTmpl` is empty. The data passed to it is the map itself, so dot is the
map. Write one line per variable, in sorted key order, in the form
`Environment="KEY=value"` (with the double quotes), each ending with a
newline. An empty map gives `""`.

```go
EnvironmentLines(map[string]string{"LOG_LEVEL": "info", "DB_HOST": "db.internal", "PORT": "8080"})
```

returns:

```text
Environment="DB_HOST=db.internal"
Environment="LOG_LEVEL=info"
Environment="PORT=8080"
```

As a Go string: `"Environment=\"DB_HOST=db.internal\"\nEnvironment=\"LOG_LEVEL=info\"\nEnvironment=\"PORT=8080\"\n"`.

Tip: `envTmpl` is a raw string (backquotes), so you can type `"` inside it
directly.

## Learn more

- [text/template: actions](https://pkg.go.dev/text/template#hdr-Actions)
- [text/template: variables](https://pkg.go.dev/text/template#hdr-Variables)
- [systemd.service manual](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html)
