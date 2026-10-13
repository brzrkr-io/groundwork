# Variadic functions: a command builder

Deploy scripts are full of generated command lines. In this exercise you build
a small command builder for your team's runbook tool, using variadic functions
at every layer.

## Learn

### Recap

A `...T` parameter takes zero or more values and is a `[]T` inside the
function. Pass an existing slice with `slice...`. A call can't mix single
values with a spread slice, so build one slice with `append` first.
`strings.Contains(s, " ")` reports whether `s` contains a space, and a
`strings.Builder` collects text piece by piece, as in *For loops: backoff and
progress bars* (`b.WriteString(...)`, then `b.String()`). You'll need to add
`"strings"` to the imports.

A string in backticks is a *raw string*: everything between the backticks is
taken literally, so `` `"` `` is a string holding one double quote. (`"\""`
means the same thing.)

## Your task

Write the three helpers in `exercise.go`. Later helpers build on earlier
ones.

### 1. Build a command line

`Command` returns `bin`, followed by each argument, all separated by single
spaces. A shell would split an argument with a space in it, and would lose an
empty argument completely. So wrap an argument in double quotes (`"`) when it
**contains a space** or **is empty**. Leave every other argument as it is.

```go
Command("systemctl", "restart", "nginx")
// => `systemctl restart nginx`

Command("git", "commit", "-m", "fix deploy")
// => `git commit -m "fix deploy"`

Command("helm", "upgrade", "--set", "")
// => `helm upgrade --set ""`

Command("uptime")
// => `uptime`
```

The examples use raw strings so the double quotes inside are easy to read.
Test failures print strings in Go syntax, where a `"` inside a string shows
up as `\"`.

### 2. Target a namespace

`Kubectl` builds a `kubectl` command with `Command`. When `namespace` is not
empty, the arguments start with `-n` and the namespace, followed by `args`.
When `namespace` is `""`, it's just `args`.

```go
Kubectl("prod", "get", "pods")
// => `kubectl -n prod get pods`

Kubectl("", "get", "nodes")
// => `kubectl get nodes`

Kubectl("staging")
// => `kubectl -n staging`

Kubectl("")
// => `kubectl`
```

### 3. Chain the steps

`Chain` joins commands with `" && "`, so each step runs only if the previous
one worked. Skip empty commands. With no commands (or only empty ones),
return `""`.

```go
Chain(
	Kubectl("prod", "apply", "-f", "web.yaml"),
	Kubectl("prod", "rollout", "status", "deploy/web"),
)
// => `kubectl -n prod apply -f web.yaml && kubectl -n prod rollout status deploy/web`

Chain("make build", "", "make test")
// => `make build && make test`

Chain()
// => "" (the empty string)
```

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [strings.Contains](https://pkg.go.dev/strings#Contains)
- [kubectl: namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
