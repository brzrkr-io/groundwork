## Task 1: Fix the [Unit] template

Every `{{range}}` needs a matching `{{end}}`. Once it parses, look at where
the newlines come from: the line with the `range` action and the line with
`end` each leave one behind.

### Closer

Add `{{end}}` on its own line after `After={{.}}`. Then put a dash in
`{{- range .After}}` and `{{- end}}` so each action eats the newline in front
of it, like the upstream block in *Text templates*.

### Answer

```go
{{- range .After}}
After={{.}}
{{- end}}
```

## Task 2: Make User optional and give Restart a default

`with` runs its body only when the value is non-empty, and sets dot to that
value. It can also have an `else`.

### Closer

Wrap the `User=` line in `{{- with .User}}` ... `{{- end}}`, each on its own
line, and use `{{.}}` inside. For the restart policy, put a
`{{with .Restart}}...{{else}}...{{end}}` right after `Restart=`.

### Answer

```go
const serviceTmpl = `[Service]
ExecStart={{.ExecStart}}
{{- with .User}}
User={{.}}
{{- end}}
Restart={{with .Restart}}{{.}}{{else}}on-failure{{end}}
`
```

## Task 3: Write the Environment lines

Dot is the map, so `range` over `.`. A `range` over a map visits keys in
sorted order, and you can name both the key and the value.

### Closer

Use `{{range $key, $value := .}}` ... `{{end}}`. Put the newline inside the
loop body, before `{{end}}`, so each pass writes one full line.

### Answer

```go
const envTmpl = `{{range $key, $value := .}}Environment="{{$key}}={{$value}}"
{{end}}`
```

An empty map runs the body zero times, so the result is `""`.
