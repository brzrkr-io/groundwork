# Variadic functions: a command builder

Deploy scripts are full of generated command lines. Here you build a small
command builder for your team's runbook tool, using variadic functions at
every layer.

**New here:** nothing; practice with variadic functions. (Raw strings in backticks are from *Formatting with fmt*, `strings.Contains` from *Strings*.)

## Learn

### Recap

- A `...T` parameter takes zero or more values and is a `[]T` inside.
- Pass an existing slice with `slice...`.
- A call can't mix single values with a spread slice, so build one slice
  with `append` first (as in *Variadic functions: fix the pager bot*).
- A `strings.Builder` collects text piece by piece: `b.WriteString(...)`,
  then `b.String()`.
- Joining with a separator but skipping empty items is what `JoinPath` did
  in *Variadic functions*.

### Importing a package

The starter file has no imports. To use `strings.Contains` or
`strings.Builder`, add this under the `package` line:

```go
import "strings"
```

> **Why is it written like this?** Go requires every package you use to be
> imported, and rejects imports you don't use.

### Does a string contain something?

`strings.Contains(s, sub)` returns `true` when `sub` appears anywhere in `s`:

```go
strings.Contains("eu-west-1", "west") // => true
strings.Contains("eu-west-1", "us")   // => false
```

### Raw strings

A string in backticks is a *raw string*: everything between the backticks is
taken literally. That makes text with double quotes inside easy to write:

```go
msg := `say "hi"`   // holds: say "hi"
quote := `"`        // holds one double quote
same := "\""        // also one double quote, written with \
```

> **Why is it written like this?** Both forms are valid Go. Picking backticks
> when the text contains `"` is a readability choice.

### Appending several values at once

`append` takes any number of values after the slice, and a `nil` slice is a
fine starting point:

```go
var flags []string
flags = append(flags, "--port", "8080") // => []string{"--port", "8080"}
```

## Your task

Write the three helpers in `exercise.go`. Later helpers build on earlier
ones.

### 1. Build a command line

**Goal:** make `Command` return `bin`, then each argument, all separated by single spaces, quoting the arguments that need it.

```go
Command("systemctl", "restart", "nginx")
// => `systemctl restart nginx`

Command("git", "commit", "-m", "fix deploy")
// => `git commit -m "fix deploy"`

Command("helm", "upgrade", "--set", "")
// => `helm upgrade --set ""`

Command("echo", "", "done")
// => `echo "" done`

Command("uptime")
// => `uptime`
```

- Wrap an argument in double quotes (`"`) when it **contains a space** or
  **is empty**. A shell would split the first kind and lose the second.
- Leave every other argument as it is.
- No space at the start or end.

The examples are raw strings so the quotes are easy to read. Test failures
print strings in normal Go syntax, where a `"` inside shows up as `\"`.

**Think it through:**
- Which piece comes before every argument, but never before `bin`?
- Which two checks decide whether an argument gets quotes?

### 2. Target a namespace

**Goal:** make `Kubectl` return a `kubectl` command built with `Command`, putting `-n` and `namespace` before `args` when `namespace` is not empty.

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

`"kubectl"` and `"-n"` are fixed text; `namespace` and `args` come from the
parameters. Arguments with spaces are still quoted, because `Command` does it.

**Think it through:**
- `Command` can't take single values *and* `args...` in one call. What do you
  build first?
- When should `-n` and `namespace` be added, and when not?

### 3. Chain the steps

**Goal:** make `Chain` return the non-empty `commands` joined with `" && "`.

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

- Skip empty commands, wherever they are.
- No `" && "` at the start or end.
- No commands, or only empty ones: return `""`.

**Think it through:**
- Which earlier helper had the same shape, with a different separator?

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [strings.Contains](https://pkg.go.dev/strings#Contains)
- [kubectl: namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
