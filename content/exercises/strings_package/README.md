# The strings package

Ops work is full of text: URLs, label lists, command lines, config lines. The
`strings` package has a small set of functions that handle most of it. Learn
these and you can take apart (and put back together) nearly any line a tool
hands you.

## Learn

You met `strings.Contains`, `strings.HasPrefix` and `strings.ToUpper` in
*Strings*. Here are the rest of the everyday tools. All of them return new
values; Go strings never change in place.

| Function | What it does | Example | Result |
|---|---|---|---|
| `Split(s, sep)` | cut `s` at every `sep` | `Split("a,b,c", ",")` | `["a" "b" "c"]` |
| `Fields(s)` | split on runs of spaces, tabs and newlines | `Fields("  a   b ")` | `["a" "b"]` |
| `Join(parts, sep)` | glue a slice back together | `Join([]string{"a", "b"}, "/")` | `"a/b"` |
| `Cut(s, sep)` | split at the **first** `sep` only | `Cut("k=v=1", "=")` | `"k", "v=1", true` |
| `Index(s, sub)` | position of the first `sub`, or `-1` | `Index("api/v1", "/")` | `3` |
| `TrimPrefix(s, p)` | drop `p` from the start, if it is there | `TrimPrefix("v1.2", "v")` | `"1.2"` |
| `TrimSuffix(s, p)` | drop `p` from the end, if it is there | `TrimSuffix("app.log", ".log")` | `"app"` |
| `TrimSpace(s)` | drop spaces, tabs and newlines at both ends | `TrimSpace("  web \n")` | `"web"` |
| `EqualFold(a, b)` | equal, ignoring upper/lower case | `EqualFold("Prod", "PROD")` | `true` |

A few details are worth knowing:

- `Split` keeps empty pieces: `Split("a,,b", ",")` is `["a" "" "b"]`, and
  `Split("", ",")` is `[""]`, a slice with one empty string.
- `TrimPrefix` and `TrimSuffix` do nothing when the prefix or suffix isn't
  there, so you can call them without checking first.

### Looping over the pieces: `SplitSeq` and `FieldsSeq`

When you only want to loop over the pieces, use `strings.SplitSeq` instead of
`Split` (and `strings.FieldsSeq` instead of `Fields`). They hand you the
pieces one at a time without building a slice first. Range over them with a
single loop variable, the piece itself:

```go
for host := range strings.SplitSeq("web-1,web-2", ",") {
	fmt.Println(host) // web-1, then web-2
}
```

They give exactly the same pieces as `Split` and `Fields`. Use
`Split` or `Fields` when you need an actual slice, for example to check its
length or index into it. Your editor suggests the `Seq` version when you range
over `Split` directly.

### `strings.Cut`

`Cut` returns three values: the part *before* the separator, the part *after*
it, and a `bool` that says whether the separator was found at all.

```go
key, value, found := strings.Cut("env=prod", "=")
// key == "env", value == "prod", found == true

key, value, found = strings.Cut("debug", "=")
// key == "debug", value == "", found == false
```

Use the `_` blank identifier for a value you don't need, as in
`before, _, _ := strings.Cut(s, "/")`.

### `strings.Builder`

You met `strings.Builder` in *For loops: backoff and progress bars*. It is the
standard way to assemble a string piece by piece:

```go
var b strings.Builder   // ready to use, no setup needed
b.WriteString("kubectl")
b.WriteString(" ")
b.WriteString("get")
cmd := b.String()       // "kubectl get"
```

`WriteString` adds text to the end. `String` gives you everything written so
far. Prefer it to `s += ...` in a loop, which copies the whole string every
time.

## Your task

You're writing helpers for a deploy tool that reads service URLs and labels and
prints the `kubectl` commands it runs.

### 1. Get the host from a URL

Change `HostFromURL` to remove a leading `"https://"` or `"http://"`, and then
everything from the first `"/"` onwards.

```go
HostFromURL("https://api.example.com/health")  // => "api.example.com"
HostFromURL("http://localhost:8080/metrics")   // => "localhost:8080"
HostFromURL("https://grafana.internal")        // => "grafana.internal"
HostFromURL("db.internal/admin")               // => "db.internal"
```

### 2. Parse a label list

Change `ParseLabels` to turn a string like `"app=web,env=prod"` into a map.

- Split the string on `","` to get the pairs. Loop over them with
  `strings.SplitSeq`.
- Split each pair at the **first** `"="`. The part before is the key, the part
  after is the value, so `"selector=a=b"` gives key `"selector"` and value
  `"a=b"`.
- Trim spaces from both the key and the value.
- Skip any pair without an `"="`.
- An empty string gives an empty map.

```go
ParseLabels("app=web, env = prod")
// => map[app:web env:prod]

ParseLabels("team=core,canary,selector=a=b")
// => map[selector:a=b team:core]
```

### 3. Build a command line

Change `BuildCommand` to return `binary` followed by each argument, with a
single space before each argument. Wrap any argument that contains a space in
double quotes (`strings.Contains(arg, " ")` tells you whether it does). With no
arguments, return just `binary`. Inside a Go string, a double quote is written
`\"`, so `"\""` is a string holding one double quote.

```go
BuildCommand("kubectl", []string{"get", "pods", "-l", "app=web tier"})
// => `kubectl get pods -l "app=web tier"`

BuildCommand("uptime", nil)
// => "uptime"
```

The backticks in the first result (and in the tests) make a *raw string*:
everything between them is taken literally, so the double quotes inside need
no `\`.

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
- [Kubernetes: Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
