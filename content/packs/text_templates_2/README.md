# Text templates: systemd unit files

Every long-running service on a Linux host is described by a systemd *unit
file*. Here you fix and finish the templates a provisioning tool uses to
write them.

**New here:** reading template parse errors, `with`, `with ... else`,
ranging over a map with `$key, $value`.

**Heads-up:** the first test fails with a template parse error until you fix
task 1. That's part of the exercise.

## Learn

### Recap

From *Text templates: render an nginx upstream block*:

| Syntax | Does |
|---|---|
| `{{.Field}}` | prints a field |
| `{{range .List}}...{{end}}` | loops; dot is the element |
| `{{if .X}}...{{else}}...{{end}}` | branches on empty vs non-empty |
| `{{-` / `-}}` | trims spaces and newlines on that side |

The `render` helper in `exercise.go` already parses and executes for you. You
only edit the template constants.

### Reading a parse error

```text
template: unit:5: unexpected EOF
```

Read it left to right:

- `unit`: the template's name.
- `5`: the line in the template.
- `unexpected EOF`: the text ended (*EOF*, end of file) while the parser was
  still waiting for something.

Every `range`, `if` and `with` needs its own `{{end}}`. A missing one gives
exactly this error. As in *Reading compiler errors*, the line number is
where the problem was *noticed*, not always where the fix goes.

### with

`{{with .X}}...{{end}}` runs its body only when `.X` is non-empty. **Inside
the body, dot is `.X`.** It's a short `if` for "print this value if there is
one":

```text
{{with .Group}}Group={{.}}{{end}}
Group: "ops"  =>  "Group=ops"
Group: ""     =>  ""
```

Add `{{else}}` to give a default when the value is empty:

```text
Shell={{with .Shell}}{{.}}{{else}}/bin/sh{{end}}
Shell: "/bin/bash"  =>  "Shell=/bin/bash"
Shell: ""           =>  "Shell=/bin/sh"
```

> **Why is it written like this?** `with` changing dot is a rule of the
> template language. Using `with` instead of `if` is the author's choice: it
> saves typing `.Group` twice.

### Ranging over a map

`range` can put each entry into *variables*. Template variables start with
`$`. Over a map you get the key, then the value:

```text
{{range $k, $v := .}}{{$k}} is {{$v}}, {{end}}
map[string]string{"b": "2", "a": "1"}  =>  "a is 1, b is 2, "
```

Go maps have no order, but a template's `range` over a map **always visits
keys in sorted order**. So the output is the same every run, which is what
you want in a config file.

> **Why is it written like this?** The `$` and the `:=` are required
> template syntax. The names `$k` and `$v` are your choice.

### Quotes inside a raw string

The template constants use backquotes (a *raw string*). Inside one, `"` is
just a character, so you can type it directly. No `\"` needed.

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

Each task fixes one template constant. In every expected output, each line
ends with `\n` (including the last), and there are never blank lines.

### 1. Fix the [Unit] template

**Goal:** fix `unitTmpl` so it parses and writes one `After=` line per entry
in `After`.

There are two bugs:

1. The `range` has no `{{end}}`, so parsing fails with
   `template: unit:5: unexpected EOF`.
2. Once it parses, you'll see blank lines between the `After=` lines.

```go
UnitSection(Unit{Description: "Payments API", After: []string{"network.target", "postgres.service"}})
// => "[Unit]\nDescription=Payments API\nAfter=network.target\nAfter=postgres.service\n"

UnitSection(Unit{Description: "Log shipper", After: []string{"network-online.target"}})
// => "[Unit]\nDescription=Log shipper\nAfter=network-online.target\n"

UnitSection(Unit{Description: "Cache warmer"})
// => "[Unit]\nDescription=Cache warmer\n"
```

**Think it through:**
- Where does the loop body end? That's where `{{end}}` goes.
- Which actions leave a newline behind? Which side of them needs a dash?

### 2. Make User optional and give Restart a default

**Goal:** change `serviceTmpl` so the `User=` line only appears when `User`
is set, and `Restart` falls back to `on-failure`.

- `User` empty: leave out the whole `User=` line (no blank line either).
- `Restart` empty: write `Restart=on-failure`.

```go
ServiceSection(Service{ExecStart: "/usr/local/bin/payments --port=8080", User: "svc-payments", Restart: "always"})
// => "[Service]\nExecStart=/usr/local/bin/payments --port=8080\nUser=svc-payments\nRestart=always\n"

ServiceSection(Service{ExecStart: "/usr/bin/shipper", Restart: "always"})
// => "[Service]\nExecStart=/usr/bin/shipper\nRestart=always\n"

ServiceSection(Service{ExecStart: "/opt/warmer/run", User: "warmer"})
// => "[Service]\nExecStart=/opt/warmer/run\nUser=warmer\nRestart=on-failure\n"

ServiceSection(Service{ExecStart: "/usr/bin/true"})
// => "[Service]\nExecStart=/usr/bin/true\nRestart=on-failure\n"
```

**Think it through:**
- Which action prints something only when a value is non-empty? Which one
  adds a default?
- When `User` is empty, which newline must disappear too?

### 3. Write the Environment lines

**Goal:** fill in the empty `envTmpl` so it writes one
`Environment="KEY=value"` line per map entry, sorted by key.

- The data is the map itself, so dot is the map.
- Each line: fixed text `Environment="`, the key, `=`, the value, a closing
  `"`, then a newline.
- An empty map gives `""`.

```go
EnvironmentLines(map[string]string{"LOG_LEVEL": "info", "DB_HOST": "db.internal", "PORT": "8080"})
// => "Environment=\"DB_HOST=db.internal\"\nEnvironment=\"LOG_LEVEL=info\"\nEnvironment=\"PORT=8080\"\n"

EnvironmentLines(map[string]string{"REGION": "eu-west-1"})
// => "Environment=\"REGION=eu-west-1\"\n"
```

Printed, the first one is:

```text
Environment="DB_HOST=db.internal"
Environment="LOG_LEVEL=info"
Environment="PORT=8080"
```

(The `\"` only appears in the Go string above. In your raw-string template,
type a plain `"`.)

**Think it through:**
- You need both the key and the value. How does `range` give you two names?
- Each pass writes one full line. Where does the newline go: inside or
  outside the loop?

## Learn more

- [text/template: actions](https://pkg.go.dev/text/template#hdr-Actions)
- [text/template: variables](https://pkg.go.dev/text/template#hdr-Variables)
- [systemd.service manual](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html)
