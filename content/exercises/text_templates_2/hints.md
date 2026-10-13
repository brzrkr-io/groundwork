## Task 1: close the range, then trim

Add `{{end}}` on its own line after `After={{.}}`. Then put a dash in
`{{- range` and `{{- end}}` so each action eats the newline in front of it.
That is the same pattern as the upstream block in *Text templates*.

## Task 2: with, and with/else

Wrap the `User=` line in `{{- with .User}}` ... `{{- end}}`, each on its own
line like the range in task 1. Inside, `{{.}}` is the user. For the restart
policy, use `{{with .Restart}}{{.}}{{else}}on-failure{{end}}`.

## Task 2: almost the answer

```go
const serviceTmpl = `[Service]
ExecStart={{.ExecStart}}
{{- with .User}}
User={{.}}
{{- end}}
Restart={{with .Restart}}{{.}}{{else}}on-failure{{end}}
`
```

## Task 3: range over the map

Put the newline inside the loop body, so each pass writes one full line:

```go
const envTmpl = `{{range $key, $value := .}}Environment="{{$key}}={{$value}}"
{{end}}`
```
