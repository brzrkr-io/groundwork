# The strings package: fix the bootstrap parser

A server bootstrap script reads a few config lines before it starts any
services. Each of its parsing helpers has one bug, and every bug comes from
picking the wrong `strings` function.

**New here:** nothing; practice choosing the right `strings` function.

## Learn

### Recap

| Function | Good for |
|---|---|
| `strings.Split(s, sep)` | cutting at **every** `sep` |
| `strings.Cut(s, sep)` | cutting at the **first** `sep` only |
| `strings.Fields(s)` | splitting on runs of whitespace; never empty pieces |
| `strings.TrimSpace(s)` | dropping whitespace at both ends |
| `strings.EqualFold(a, b)` | comparing, ignoring upper/lower case |

### Why `Split(line, " ")` breaks on real files

People line config up with extra spaces or tabs. `Split` treats every single
space as a separator:

```go
strings.Split("eth0   up", " ")
// => ["eth0" "" "" "up"]

strings.Fields("eth0   up")
// => ["eth0" "up"]
```

### Calls inside calls

You can pass one function's result straight into another. The inner call
runs first:

```go
strings.HasPrefix(strings.ToLower(name), "web")
// lowercase name first, then check the prefix
```

> **Why is it written like this?** Go requires the inner call to return the
> type the outer one expects (here a `string`). Nesting the calls or using a
> temporary variable is your choice.

## Your task

Each function has one bug. Find it and fix it; keep the rest of the code.

### 1. Recognise production, however it is typed

**Goal:** make `IsProduction` return `true` for `"production"` in any case,
with any whitespace around it.

Right now only the exact lowercase word matches. Spaces, tabs and newlines
around it must be ignored. Other words, including `"prod"` and `""`, give
`false`.

```go
IsProduction("production")      // => true
IsProduction(" Production\n")   // => true
IsProduction("PRODUCTION")      // => true
IsProduction("staging")         // => false
IsProduction("prod")            // => false
```

**Think it through:**
- Two problems: the whitespace and the case. Which tool fixes each?
- Which one has to happen to `env` first?

### 2. Keep the whole value of an env line

**Goal:** make `ParseEnvLine` keep a value whole even when it contains `"="`.

It returns the key, the value and `true`, with spaces trimmed from key and
value. A line without `"="` returns `"", "", false`. An empty value is fine:
`"PROXY="` gives `"PROXY", "", true`.

```go
ParseEnvLine("PORT=8080")
// => "PORT", "8080", true

ParseEnvLine("DB_URL=postgres://db/app?sslmode=disable")
// => "DB_URL", "postgres://db/app?sslmode=disable", true

ParseEnvLine("JAVA_OPTS = -Xmx512m -Dfile.encoding=UTF-8")
// => "JAVA_OPTS", "-Xmx512m -Dfile.encoding=UTF-8", true

ParseEnvLine("DEBUG")
// => "", "", false
```

**Think it through:**
- Count the `"="` in the URL line. What does `Split` do with each one?
- Which tool cuts at the first one only, and how does it tell you "not
  found"?

### 3. Read `/etc/hosts` lines with any spacing

**Goal:** make `ParseHostsLine` work no matter how many spaces or tabs
separate the words.

A line is an IP address followed by one or more host names. It returns the
IP, the names and `true`. Anything after `#` is a comment (the code already
removes it). If fewer than two words are left (blank line, comment only, IP
only), it returns `"", nil, false`.

```go
ParseHostsLine("127.0.0.1   localhost  localhost.localdomain")
// => "127.0.0.1", ["localhost" "localhost.localdomain"], true

ParseHostsLine("\t10.0.0.5\tdb   # primary database")
// => "10.0.0.5", ["db"], true

ParseHostsLine("# managed by ansible")
// => "", nil, false
```

**Think it through:**
- Look at the line that makes `words`. What does it produce for two spaces
  in a row, or a tab?

## Learn more

- [Package strings: Cut](https://pkg.go.dev/strings#Cut)
- [Package strings: Fields](https://pkg.go.dev/strings#Fields)
- [Package strings: EqualFold](https://pkg.go.dev/strings#EqualFold)
