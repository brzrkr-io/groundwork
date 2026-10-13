# The strings package: fix the bootstrap parser

A server bootstrap script reads a few config lines before it starts any
services. Its parsing helpers each have one bug, and every bug comes from
picking the wrong `strings` function.

## Learn

### Recap

`strings.Split` cuts at **every** separator, while `strings.Cut` cuts at the
**first** one only. `strings.Fields` splits on runs of whitespace and never
gives you empty pieces. `strings.TrimSpace` drops whitespace at both ends, and
`strings.EqualFold` compares two strings ignoring upper and lower case.

### Why `Split(line, " ")` breaks on real files

Config files are written by people, and people line things up with extra
spaces or tabs. `Split` treats every single space as a separator:

```go
strings.Split("10.0.0.5   db", " ")
// => ["10.0.0.5" "" "" "db"]

strings.Fields("10.0.0.5   db")
// => ["10.0.0.5" "db"]
```

## Your task

Each function has one bug. Find it, fix it, and keep the rest of the code.

### 1. Recognise production, however it is typed

`IsProduction` should report whether `env` names the production environment.
It must match `"production"` in any mix of upper and lower case, and ignore
spaces, tabs and newlines around it. Right now only the exact lowercase word
matches.

```go
IsProduction("production")      // => true
IsProduction(" Production\n")   // => true
IsProduction("PRODUCTION")      // => true
IsProduction("staging")         // => false
IsProduction("prod")            // => false
```

### 2. Keep the whole value of an env line

`ParseEnvLine` reads an environment file line like `PORT=8080`. It returns the
key, the value and `true`, with spaces trimmed from the key and value. A line
without `"="` returns `"", "", false`. The bug: values that contain `"="`
themselves get chopped off.

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

### 3. Read `/etc/hosts` lines with any spacing

`ParseHostsLine` reads one line of an `/etc/hosts` file: an IP address
followed by one or more host names. It returns the IP, the names and `true`.
Anything after a `#` is a comment and is ignored. If what's left has fewer
than two words (a blank line, or a comment-only line), it returns
`"", nil, false`. The bug: lines that use several spaces, tabs, or leading
spaces come out wrong.

```go
ParseHostsLine("127.0.0.1   localhost  localhost.localdomain")
// => "127.0.0.1", ["localhost" "localhost.localdomain"], true

ParseHostsLine("\t10.0.0.5\tdb   # primary database")
// => "10.0.0.5", ["db"], true

ParseHostsLine("# managed by ansible")
// => "", nil, false
```

## Learn more

- [Package strings: Cut](https://pkg.go.dev/strings#Cut)
- [Package strings: Fields](https://pkg.go.dev/strings#Fields)
- [Package strings: EqualFold](https://pkg.go.dev/strings#EqualFold)
