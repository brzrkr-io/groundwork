# Errors are values: parse health checks

A monitoring agent reads its health checks from a small text format, one check
per line. A typo in that file should produce an error that points straight at
the bad line, not a silent wrong check.

## Learn

Recap: return `error` last, `nil` means success, return early with zero values
on failure, and add context in front of an error with
`fmt.Errorf("context: %v", err)`.

The zero value of a struct is the struct with every field at its zero value,
written `Check{}`. Return that alongside an error. For a slice the zero value
is `nil`.

Context stacks up as an error travels up through your functions. Each layer
adds what it knows:

```text
line 2: check "api x /healthz": port "x" is not a number
^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^
ParseAll  ParseCheck            the actual problem
```

## Your task

Each line of the file holds a service name, a port and an HTTP path,
separated by spaces:

```text
api 8080 /healthz
```

You need `strings`, `strconv` and `fmt`, so add those imports.

### 1. Parse one line

`ParseCheck` splits the line with `strings.Fields` and returns a `Check`.
Check these in order. On any error, return `Check{}` and the error. In the
messages, `<line>` is the whole input line, quoted with `%q`:

| Problem | Error message |
|---|---|
| not exactly 3 fields | `check <line>: want 3 fields, got 2` |
| port is not a number | `check <line>: port "x" is not a number` |
| port outside 1-65535 | `check <line>: port 0 is out of range 1-65535` |
| path doesn't start with `/` | `check <line>: path "healthz" must start with "/"` |

```go
ParseCheck("api 8080 /healthz")
// => Check{Name: "api", Port: 8080, Path: "/healthz"}, nil
ParseCheck("api 8080")
// => Check{}, error "check \"api 8080\": want 3 fields, got 2"
ParseCheck("api 8080 healthz")
// => Check{}, error "check \"api 8080 healthz\": path \"healthz\" must start with \"/\""
```

### 2. Parse the whole file

`ParseAll` parses every line and returns the checks in order. It skips blank
lines (lines that are empty or only spaces; `strings.TrimSpace` helps). Line
numbers start at 1 and count every line, blank ones too.

At the first bad line, stop and return `nil` and the error from `ParseCheck`
with `line <n>: ` in front:

```go
ParseAll([]string{"api 8080 /healthz", "", "web 80 /"})
// => [{api 8080 /healthz} {web 80 /}], nil
ParseAll([]string{"api 8080 /healthz", "web x /"})
// => nil, error "line 2: check \"web x /\": port \"x\" is not a number"
```

## Learn more

- [pkg.go.dev: strings.Fields](https://pkg.go.dev/strings#Fields)
- [pkg.go.dev: fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
- [Kubernetes: Configure liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
