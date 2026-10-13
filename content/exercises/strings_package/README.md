# The strings package

Ops work is full of text: URLs, label lists, command lines. The `strings`
package has a small set of functions that take apart, and put back together,
nearly any line a tool hands you.

**New here:** `strings.Split`, `Fields`, `Join`, `Cut`, `TrimPrefix`,
`TrimSuffix`, `TrimSpace`, `SplitSeq`, `strings.Builder` in a loop.

## Learn

### The everyday tools

You met `Contains`, `HasPrefix` and `ToUpper` in *Strings*. All of these
return a **new** value; Go strings never change in place.

| Function | What it does | Example | Result |
|---|---|---|---|
| `Split(s, sep)` | cut at every `sep` | `Split("a,b,c", ",")` | `["a" "b" "c"]` |
| `Fields(s)` | split on runs of whitespace | `Fields("  a   b ")` | `["a" "b"]` |
| `Join(parts, sep)` | glue a slice together | `Join([]string{"a", "b"}, "/")` | `"a/b"` |
| `Cut(s, sep)` | split at the **first** `sep` | `Cut("k=v=1", "=")` | `"k", "v=1", true` |
| `Index(s, sub)` | position of `sub`, or `-1` | `Index("api/v1", "/")` | `3` |
| `TrimPrefix(s, p)` | drop `p` from the start | `TrimPrefix("v1.2", "v")` | `"1.2"` |
| `TrimSuffix(s, p)` | drop `p` from the end | `TrimSuffix("app.log", ".log")` | `"app"` |
| `TrimSpace(s)` | drop whitespace at both ends | `TrimSpace("  web \n")` | `"web"` |
| `EqualFold(a, b)` | equal, ignoring case | `EqualFold("Prod", "PROD")` | `true` |

- `Split` keeps empty pieces: `Split("a,,b", ",")` is `["a" "" "b"]`, and
  `Split("", ",")` is `[""]`.
- `TrimPrefix`/`TrimSuffix` do nothing if the text isn't there, so no check
  is needed first.

Because they return a new string, keep the result: `name = strings.TrimSpace(name)`.
Calling `strings.TrimSpace(name)` on its own line throws the result away.

> **Why is it written like this?** `strings.` in front is required: it names
> the package the function lives in. The argument order (text first, then
> separator) is fixed by the package.

### Looping over pieces: `SplitSeq`

`strings.SplitSeq` gives the same pieces as `Split`, one at a time, without
building a slice. Range over it with one loop variable, the piece:

```go
for host := range strings.SplitSeq("web-1;web-2", ";") {
	fmt.Println(host) // web-1, then web-2
}
```

Use `Split` when you need a real slice (to check its length or index it).

> **Why is it written like this?** One loop variable is required here; there
> is no index. The name `host` is your choice. Your editor suggests
> `SplitSeq` when you range over `Split`.

### `strings.Cut`

`Cut` returns three values: the text *before* the separator, the text
*after* it, and a `bool` saying whether the separator was found.

```go
user, domain, found := strings.Cut("ops@example.com", "@")
// "ops", "example.com", true

user, domain, found = strings.Cut("root", "@")
// "root", "", false
```

Write `_` for a value you don't need: `user, _, _ := strings.Cut(s, "@")`.

> **Why is it written like this?** Go requires one name (or `_`) for each
> returned value. `:=` creates new names; `=` reuses existing ones.

### `strings.Builder`

You met it in *For loops: backoff and progress bars*. `WriteString` adds to
the end; `String` returns everything written.

```go
var b strings.Builder
b.WriteString("ls")
b.WriteString(" -la")
out := b.String() // "ls -la"
```

### Quotes inside a string

Inside `"..."`, write a double quote as `\"`. So `"\""` is a string holding
one `"` character, and `"say \"hi\""` is `say "hi"`. A *raw string* in
backticks, like `` `say "hi"` ``, takes everything literally.

## Your task

You're writing helpers for a deploy tool that reads service URLs and labels
and prints the `kubectl` commands it runs.

### 1. Get the host from a URL

**Goal:** make `HostFromURL` return just the host part of `url`.

Remove a leading `"https://"` or `"http://"`, then drop everything from the
first `"/"` onwards.

```go
HostFromURL("https://api.example.com/health")  // => "api.example.com"
HostFromURL("http://localhost:8080/metrics")   // => "localhost:8080"
HostFromURL("https://grafana.internal")        // => "grafana.internal"
HostFromURL("db.internal/admin")               // => "db.internal"
```

**Think it through:**
- Which tool removes text from the start only if it's there?
- Which tool keeps only the part before the first `"/"`?
- Are you keeping each result, or throwing it away?

### 2. Parse a label list

**Goal:** make `ParseLabels` turn `"app=web,env=prod"` into a map of keys to
values.

- Pairs are separated by `","`. Loop over them with `strings.SplitSeq`.
- Split each pair at the **first** `"="`: before is the key, after is the
  value. `"selector=a=b"` gives key `"selector"`, value `"a=b"`.
- Trim spaces from both key and value.
- Skip any pair without `"="`.
- An empty string gives an empty map.

```go
ParseLabels("app=web, env = prod")
// => map[app:web env:prod]

ParseLabels("team=core,canary,selector=a=b")
// => map[selector:a=b team:core]
```

**Think it through:**
- The map `labels` already exists; you only fill it.
- Which of `Cut`'s three results tells you to skip a pair?

### 3. Build a command line

**Goal:** make `BuildCommand` return `binary`, then each argument with one
space before it, quoting arguments that contain a space.

- Wrap an argument in double quotes if it contains a space
  (`strings.Contains(arg, " ")` tells you).
- No arguments: return just `binary`.

```go
BuildCommand("kubectl", []string{"get", "pods", "-l", "app=web tier"})
// => `kubectl get pods -l "app=web tier"`

BuildCommand("uptime", nil)
// => "uptime"
```

The backticks in the first result are a raw string; the result itself
contains the double quotes, not the backticks.

**Think it through:**
- The loop already writes `arg`. What must be written before it, every time?
- When quoting, which three pieces make the new argument: fixed text,
  the variable, fixed text?

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
- [Kubernetes: Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
